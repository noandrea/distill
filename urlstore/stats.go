package urlstore

import (
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/jbrodriguez/mlog"

	"github.com/dgraph-io/badger"
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
		s.LastRequest = time.Now()
		s.Gets++
	case opcodeExpired:
		s.LastRequest = time.Now()
		s.GetsExpired++
	}
	UpdateStats(s)
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
			globalStatistics.record(1, 0, 0, 0, 0)
		case opcodeInsert:
			// TODO: check if existed already
			globalStatistics.record(0, 1, 0, 1, 0)
		case opcodeDelete:
			globalStatistics.record(0, 0, 1, -1, 0)
		case opcodeExpired:
			globalStatistics.record(0, 0, 0, 0, 1)
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

func (s *Statistics) record(get, upsert, delete, urls, getExpired int64) {
	statsMutex.Lock()
	// this is confusing but actually correct
	// if the input number is negative will work just the same
	s.Gets += uint64(get)
	s.GetsExpired += uint64(getExpired)
	s.Upserts += uint64(upsert)
	s.Deletes += uint64(delete)
	s.Urls += uint64(urls)
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
	gcLimit := Config.Tuning.DbGCDeletesCount
	gcCount := uint64(0)
	// retrieve the gcCount from the db
	db.View(func(txn *badger.Txn) (err error) {
		gcCount = dbGetUint64(txn, sysKeyGCCount)
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

		db.RunValueLogGC(Config.Tuning.DbGCDiscardRation)
		mlog.Info("End maintenance n %d for deletes %d > %d", gcCount, deletes-latestGC, gcLimit)
		// update the gcCount
		db.Update(func(txn *badger.Txn) (err error) {
			gcCount++
			dbSetUint64(txn, sysKeyGCCount, gcCount)
			return
		})
	}
}
