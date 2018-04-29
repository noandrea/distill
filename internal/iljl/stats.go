package iljl

import (
	"fmt"
	"sync"

	"github.com/jbrodriguez/mlog"

	"github.com/dgraph-io/badger"
	"gitlab.com/lowgroundandbigshoes/iljl/internal"
)

var (
	wg               sync.WaitGroup
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
		g := func(key []byte) (count int64, err error) {
			item, err := txn.Get(key)
			if err != nil {
				if err == badger.ErrKeyNotFound {
					return 0, nil
				}
				return
			}
			val, err := item.Value()
			if err != nil {
				return
			}
			count = atoi(val)
			return
		}

		s.Urls, err = g(statsKeyGlobalURLCount)
		if err != nil {
			return
		}
		s.Gets, err = g(statsKeyGlobalGetCount)
		if err != nil {
			return
		}
		s.Deletes, err = g(statsKeyGlobalDelCount)
		if err != nil {
			return
		}
		s.Upserts, err = g(statsKeyGlobalUpdCount)
		if err != nil {
			return
		}
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
		set := func(key []byte, v int64) (err error) {
			err = txn.Set(key, itoa(v))
			return
		}

		// find all the urls
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 200
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		ucp := []byte{KeyURLStatCountPrefix}
		for it.Seek(ucp); it.ValidForPrefix(ucp); it.Next() {
			v, err := it.Item().Value()
			if err != nil {
				v = numberZero
			}
			s.Urls++
			s.Upserts++
			s.Gets += atoi(v)
		}

		err = set(statsKeyGlobalURLCount, s.Urls)
		if err != nil {
			return
		}
		err = set(statsKeyGlobalGetCount, s.Gets)
		if err != nil {
			return
		}
		err = set(statsKeyGlobalDelCount, 0)
		if err != nil {
			return
		}
		err = set(statsKeyGlobalUpdCount, s.Upserts)
		if err != nil {
			return
		}
		// update global statistics
		globalStatistics = s
		return
	})
	if err != nil {
		mlog.Error(err)
	}
	return
}

// pushEvent in the url operaiton queue
func pushEvent(urlop *URLOp) {
	opEventsQueue <- urlop
}

// Process is an implementation of wp.Job.Process()
func processEvents(workerID int) {
	for {
		uo, ok := <-opEventsQueue
		if !ok {
			break
		}

		switch uo.opcode {
		case opcodeGet:
			db.Update(func(txn *badger.Txn) (err error) {
				globalStatistics.Gets++
				// update gets count
				err = txn.Set(statsKeyGlobalGetCount, itoa(globalStatistics.Gets))
				if err != nil {
					mlog.Error(fmt.Errorf("global stats update error %v ", err))
				}
				// update url get count
				k := keyURLStatCount(uo.url.ID)
				v, err := dbGet(txn, k)
				if err != nil {
					v = numberZero
				}
				// Set the same ttl in case
				if uo.url.TTL > 0 {
					err = txn.SetWithTTL(k, itoa(atoi(v)+1), ttl(uo.url.TTL))
				} else {
					err = txn.Set(k, itoa(atoi(v)+1))
				}
				// TODO: if the count is > than the maxRequests delete the url
				return
			})
		case opcodeInsert:
			db.Update(func(txn *badger.Txn) (err error) {
				globalStatistics.Urls++
				globalStatistics.Upserts++
				// update urls count
				err = txn.Set(statsKeyGlobalURLCount, itoa(globalStatistics.Urls))
				if err != nil {
					mlog.Error(fmt.Errorf("global stats update error %v ", err))
				}
				// update upserts count
				err = txn.Set(statsKeyGlobalUpdCount, itoa(globalStatistics.Upserts))
				if err != nil {
					mlog.Error(fmt.Errorf("global stats update error %v ", err))
				}
				return
			})
		case opcodeDelete:
			db.Update(func(txn *badger.Txn) (err error) {
				globalStatistics.Urls--
				globalStatistics.Deletes++
				// update urls count
				err = txn.Set(statsKeyGlobalURLCount, itoa(globalStatistics.Urls))
				if err != nil {
					mlog.Error(fmt.Errorf("global stats update error %v ", err))
				}
				// update upserts count
				err = txn.Set(statsKeyGlobalDelCount, itoa(globalStatistics.Deletes))
				if err != nil {
					mlog.Error(fmt.Errorf("global stats update error %v ", err))
				}
				return
			})

			// TODO: run database maintenance here
		}
		// run the maintenance
		go runDbMaintenance()
	}
	// complete task
	wg.Done()
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
		val, err := dbGet(txn, sysKeyGCCount)
		if err != nil {
			val = numberZero
		}
		gcCount = atoi(val)
		return
	})

	latestGC := gcCount * gcLimit
	if latestGC > deletes {
		// there was a reset should reset in the stats
		gcCount, latestGC = 0, 0
	}

	if deletes-latestGC > gcLimit {
		mlog.Info("Start maintenance n %d for deletes %d > %d", gcCount, deletes-latestGC, gcLimit)
		err := db.PurgeOlderVersions()
		if err != nil {
			mlog.Warning("Purge db failed %v", err)
			return
		}
		db.RunValueLogGC(internal.Config.Tuning.DbGCDiscardRation)
		mlog.Info("End maintenance n %d for deletes %d > %d", gcCount, deletes-latestGC, gcLimit)
		// updaete the gcCount
		db.Update(func(txn *badger.Txn) (err error) {
			gcCount++
			err = txn.Set(sysKeyGCCount, itoa(gcCount))
			if err != nil {
				mlog.Warning("Error updating %X : %v", sysKeyGCCount, err)
			}
			return
		})
		// unlock the maintenance
		maintenanceRunning = false
	}
	wg.Done()
}
