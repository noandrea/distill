package distill

import (
	"fmt"
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/jbrodriguez/mlog"

	"github.com/dgraph-io/badger"
	"gitlab.com/welance/distill/internal"
)

var (
	wg               sync.WaitGroup
	sc               gcache.Cache
	opEventsQueue    chan *URLOp
	globalStatistics *Statistics
	// sytem keys
	sysKeyPurgeCount []byte
	sysKeyGCCount    []byte
	// stats keys
	statsKeyGlobalURLCount []byte
	statsKeyGlobalGetCount []byte
	statsKeyGlobalDelCount []byte
	statsKeyGlobalUpdCount []byte
)

// NewStatistics starts the statistics collector worker pool
func NewStatistics() (err error) {
	// initializae system key
	sysKeyPurgeCount = keySys("ilij_sys_purge_count")
	sysKeyGCCount = keySys("ilij_sys_gc_count")
	// initialize stats keys
	statsKeyGlobalURLCount = keyGlobalStat("ilij_global_url_count")
	statsKeyGlobalGetCount = keyGlobalStat("ilij_global_get_count")
	statsKeyGlobalDelCount = keyGlobalStat("ilij_global_del_count")
	statsKeyGlobalUpdCount = keyGlobalStat("ilij_global_upd_count")
	// read the current statistics
	globalStatistics, err = loadGlobalStatistics()
	if err != nil {
		return
	}
	// initialzie internal cache to provide idempotency for events
	sc = gcache.New(internal.Config.Tuning.StatsCaheSize).
		ARC().
		Build()
	// Initialize channel of events
	mlog.Trace("intialize queue size %d", internal.Config.Tuning.StatsEventsQueueSize)
	opEventsQueue = make(chan *URLOp, internal.Config.Tuning.StatsEventsQueueSize)
	// start the routines
	for i := 0; i < internal.Config.Tuning.StatsEventsWorkerNum; i++ {
		wg.Add(1)
		go processEvents(i)
	}

	return
}

// StopStatistics stops the statistics
func StopStatistics() {
	// stop the running
	mlog.Info("Stop statistics, lost %d events", len(opEventsQueue))
	close(opEventsQueue)
	wg.Wait()
}

// GetStats retrieve the global statistics
func GetStats() (s *Statistics) {
	return globalStatistics
}

func loadGlobalStatistics() (s *Statistics, err error) {
	s = &Statistics{}
	err = db.View(func(txn *badger.Txn) (err error) {
		s.Urls = dbGetInt64(txn, statsKeyGlobalURLCount)
		s.Gets = dbGetInt64(txn, statsKeyGlobalGetCount)
		s.Deletes = dbGetInt64(txn, statsKeyGlobalDelCount)
		s.Upserts = dbGetInt64(txn, statsKeyGlobalUpdCount)
		return
	})
	globalStatistics = s
	mlog.Info("Statistics are %v", s)
	return
}

func resetGlobalStatistics() (err error) {
	s := &Statistics{}
	// run the update
	err = db.Update(func(txn *badger.Txn) (err error) {
		// find all the urls
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 200
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		ucp := []byte{keyURLStatCountPrefix}
		for it.Seek(ucp); it.ValidForPrefix(ucp); it.Next() {
			s.Urls++
			s.Upserts++
			// update gets
			v, err := it.Item().Value()
			if err != nil {
				v = numberZero
			}
			s.Gets += atoi(v)
		}

		dbSetInt64(txn, statsKeyGlobalURLCount, s.Urls)
		dbSetInt64(txn, statsKeyGlobalGetCount, s.Gets)
		dbSetInt64(txn, statsKeyGlobalDelCount, 0)
		dbSetInt64(txn, statsKeyGlobalUpdCount, s.Upserts)
		// update global statistics
		globalStatistics = s
		return
	})
	if err != nil {
		mlog.Warning("Error while rest stats %v", err)
	}
	return
}

// pushEvent in the url operaiton queue
func pushEvent(urlop *URLOp) {
	switch urlop.opcode {
	case opcodeDelete, opcodeExpired:
		key := fmt.Sprint(urlop.opcode, ":", urlop.ID)
		// if it's in cache do not queue the job
		if _, err := sc.Get(key); err == nil {
			return
		}
		sc.SetWithExpire(key, nil, time.Duration(60)*time.Second)
	case opcodeInsert:
		for _, key := range []int{opcodeDelete, opcodeExpired} {
			sc.Remove(key)
		}
	}
	opEventsQueue <- urlop
}

// Process is an implementation of wp.Job.Process()
func processEvents(workerID int) {

	for {
		uo, isChannelOpen := <-opEventsQueue
		if !isChannelOpen {
			break
		}
		mlog.Trace(">>> Event pid: %v, opcode:%v  %v", workerID, opcodeToString(uo.opcode), uo.ID)
		switch uo.opcode {
		case opcodeGet:
			db.Update(func(txn *badger.Txn) (err error) {
				// update global gets count
				globalStatistics.recGet()
				dbSetInt64(txn, statsKeyGlobalGetCount, globalStatistics.Gets)
				return
			})
		case opcodeInsert:
			db.Update(func(txn *badger.Txn) (err error) {
				globalStatistics.recUpsert()
				// update urls count
				dbSetInt64(txn, statsKeyGlobalURLCount, globalStatistics.Urls)
				// update upserts count
				dbSetInt64(txn, statsKeyGlobalUpdCount, globalStatistics.Upserts)
				return
			})
		case opcodeDelete:
			db.Update(func(txn *badger.Txn) (err error) {
				globalStatistics.recDelete()
				// update urls count
				dbSetInt64(txn, statsKeyGlobalURLCount, globalStatistics.Urls)
				// update deletes count
				dbSetInt64(txn, statsKeyGlobalDelCount, globalStatistics.Deletes)
				return
			})
		case opcodeExpired:
			db.Update(func(txn *badger.Txn) (err error) {
				// first verify that the url has not been already deleted
				err = Delete(uo.ID)
				if err != nil {
					return
				}
				globalStatistics.recDelete()
				// update urls count
				dbSetInt64(txn, statsKeyGlobalURLCount, globalStatistics.Urls)
				// update deletes count
				dbSetInt64(txn, statsKeyGlobalDelCount, globalStatistics.Deletes)
				return
			})
		}
		mlog.Trace("<<< Event pid: %v, opcode:%v  %v", workerID, opcodeToString(uo.opcode), uo.ID)
		// run the maintenance
		go runDbMaintenance()
	}
	// complete task
	mlog.Trace("Stop event processor id: %v", workerID)
	wg.Done()

}

func (s *Statistics) recUpsert() {
	s.mutex.Lock()
	s.Upserts++
	s.Urls++
	s.mutex.Unlock()
}

func (s *Statistics) recDelete() {
	s.mutex.Lock()
	s.Deletes++
	s.Urls--
	s.mutex.Unlock()
}

func (s *Statistics) recGet() {
	s.mutex.Lock()
	s.Gets++
	s.mutex.Unlock()
}

// runDbMaintenance
var maintenanceRunning = false

func runDbMaintenance() {
	if maintenanceRunning {
		return
	}

	maintenanceRunning = true
	wg.Add(1)
	// caluclate if gc is necessary
	deletes := globalStatistics.Deletes
	gcLimit := internal.Config.Tuning.DbGCDeletesCount
	gcCount := int64(0)
	// retrieve the gcCount from the db
	db.View(func(txn *badger.Txn) (err error) {
		gcCount = dbGetInt64(txn, sysKeyGCCount)
		return
	})

	latestGC := gcCount * gcLimit
	if latestGC > deletes {
		// there was a reset should reset in the stats
		gcCount, latestGC = 0, 0
	}

	if deletes-latestGC > gcLimit {
		mlog.Info("Start maintenance n %d for deletes %d > %d", gcCount, deletes-latestGC, gcLimit)

		mlog.Info("")

		db.RunValueLogGC(internal.Config.Tuning.DbGCDiscardRation)
		mlog.Info("End maintenance n %d for deletes %d > %d", gcCount, deletes-latestGC, gcLimit)
		// updaete the gcCount
		db.Update(func(txn *badger.Txn) (err error) {
			gcCount++
			dbSetInt64(txn, sysKeyGCCount, gcCount)
			return
		})
		// unlock the maintenance
		maintenanceRunning = false
	}
	wg.Done()
}
