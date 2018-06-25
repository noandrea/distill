package distill

import (
	"sync"

	"github.com/bluele/gcache"
	"github.com/jbrodriguez/mlog"

	"github.com/dgraph-io/badger"
	"gitlab.com/welance/oss/distill/internal"
)

var (
	wg               sync.WaitGroup
	sc               gcache.Cache
	opEventsQueue    chan *URLOp
	globalStatistics *Statistics
	// sytem keys
	sysKeyPurgeCount []byte
	sysKeyGCCount    []byte
)

// pushEvent in the url operaiton queue
func pushEvent(urlop *URLOp) {
	s := Statistics{}
	switch urlop.opcode {
	case opcodeDelete:
		s.Deletes++
		s.Urls--
	case opcodeInsert:
		s.Upserts++
		s.Urls++
	case opcodeGet:
		s.Gets++
	case opcodeExpired:
		s.Deletes++
		s.Urls--
		Delete(urlop.ID)
	}
	UpdateStats(s)
	// switch urlop.opcode {
	// case opcodeDelete, opcodeExpired:
	// 	key := fmt.Sprint(urlop.opcode, ":", urlop.ID)
	// 	// if it's in cache do not queue the job
	// 	if _, err := sc.GetIFPresent(key); err == nil {
	// 		// if not set a nil value for some seconds so it
	// 		sc.SetWithExpire(key, nil, time.Duration(60)*time.Second)
	// 		return
	// 	}
	// 	// if not set a nil value for some seconds so it
	// 	sc.SetWithExpire(key, nil, time.Duration(60)*time.Second)
	// case opcodeInsert:
	// 	for _, key := range []int{opcodeDelete, opcodeExpired} {
	// 		sc.Remove(key)
	// 	}
	// }
	//opEventsQueue <- urlop
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
			globalStatistics.record(1, 0, 0, 0)
		case opcodeInsert:
			// TODO: check if existed already
			globalStatistics.record(0, 1, 0, 1)
		case opcodeDelete:
			globalStatistics.record(0, 0, 1, -1)
		case opcodeExpired:
			globalStatistics.record(0, 0, 1, -1)
		}
		mlog.Trace("<<< Event pid: %v, opcode:%v  %v", workerID, opcodeToString(uo.opcode), uo.ID)
	}
	// complete task
	mlog.Trace("Stop event processor id: %v", workerID)
	wg.Done()

}

// runDbMaintenance
var (
	statsMutex sync.Mutex
)

func (s *Statistics) record(get, upsert, delete, urls int64) {
	statsMutex.Lock()
	s.Gets += get
	s.Upserts += upsert
	s.Deletes += delete
	s.Urls += urls
	statsMutex.Unlock()
}

// runDbMaintenance
var (
	maintenanceMutex   sync.Mutex
	maintenanceRunning = false
)

// setRunMaintenance change maintenance status
func setRunMaintenance(val bool) {
	maintenanceMutex.Lock()
	defer maintenanceMutex.Unlock()
	maintenanceRunning = val
}

// isMaintenanceRunning check if there is already a routine doing maintenance
func isMaintenanceRunning() bool {
	maintenanceMutex.Lock()
	defer maintenanceMutex.Unlock()
	return maintenanceRunning
}

// runDbMaintenance runs the database maintenance
// TODO: add tests for this function
func runDbMaintenance() {
	if isMaintenanceRunning() {
		return
	}
	setRunMaintenance(true)
	defer setRunMaintenance(false)
	wg.Add(1)
	defer wg.Done()

	// caluclate if gc is necessary
	deletes := globalStatistics.Deletes
	gcLimit := internal.Config.Tuning.DbGCDeletesCount
	gcCount := 0
	// retrieve the gcCount from the db
	db.View(func(txn *badger.Txn) (err error) {
		gcCount = int(dbGetInt64(txn, sysKeyGCCount))
		return
	})

	latestGC := gcCount * gcLimit
	if latestGC > int(deletes) {
		// there was a reset should reset in the stats
		gcCount, latestGC = 0, 0
	}

	if int(deletes)-latestGC > gcLimit {
		mlog.Info("Start maintenance n %d for deletes %d > %d", gcCount, int(deletes)-latestGC, gcLimit)

		mlog.Info("")

		db.RunValueLogGC(internal.Config.Tuning.DbGCDiscardRation)
		mlog.Info("End maintenance n %d for deletes %d > %d", gcCount, int(deletes)-latestGC, gcLimit)
		// update the gcCount
		db.Update(func(txn *badger.Txn) (err error) {
			gcCount++
			dbSetInt64(txn, sysKeyGCCount, int64(gcCount))
			return
		})
	}
}
