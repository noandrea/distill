package datastore

import (
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/noandrea/distill/pkg/model"
	log "github.com/sirupsen/logrus"
)

var (
	wg               sync.WaitGroup
	sc               gcache.Cache
	opEventsQueue    chan *model.URLOp
	globalStatistics *model.Statistics
	// sytem keys
	sysKeyPurgeCount []byte
	sysKeyGCCount    []byte
)

// PushEvent in the url operaiton queue
func PushEvent(urlop *model.URLOp) {
	s := model.Statistics{}
	switch urlop.Opcode {
	case model.OpcodeDelete:
		s.Deletes++
		s.Urls--
	case model.OpcodeInsert:
		s.Upserts++
		s.Urls++
	case model.OpcodeGet:
		s.LastRequest = time.Now()
		s.Gets++
	case model.OpcodeExpired:
		s.LastRequest = time.Now()
		s.GetsExpired++
	}
	//UpdateStats(s)
}

// Process is an implementation of wp.Job.Process()
func processEvents(workerID int) {

	for {
		uo, isChannelOpen := <-opEventsQueue
		if !isChannelOpen {
			break
		}
		log.Tracef(">>> Event pid: %v, opcode:%v  %v", workerID, model.OpcodeToString(uo.Opcode), uo.ID)
		switch uo.Opcode {
		case model.OpcodeGet:
			globalStatistics.Record(1, 0, 0, 0, 0)
		case model.OpcodeInsert:
			// TODO: check if existed already
			globalStatistics.Record(0, 1, 0, 1, 0)
		case model.OpcodeDelete:
			globalStatistics.Record(0, 0, 1, -1, 0)
		case model.OpcodeExpired:
			globalStatistics.Record(0, 0, 0, 0, 1)
		}
		log.Tracef("<<< Event pid: %v, opcode:%v  %v",
			workerID,
			model.OpcodeToString(uo.Opcode),
			uo.ID)
	}
	// complete task
	log.Tracef("Stop event processor id: %v", workerID)
	wg.Done()

}

// runDbMaintenance
var (
	statsMutex sync.Mutex
)

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
	// if isMaintenanceRunning() {
	// 	return
	// }
	// setRunMaintenance(true)
	// defer setRunMaintenance(false)
	// wg.Add(1)
	// defer wg.Done()

	// // caluclate if gc is necessary
	// deletes := uint64(globalStatistics.Deletes)
	// gcLimit := settings.Tuning.DbGCDeletesCount
	// gcCount := uint64(0)
	// // retrieve the gcCount from the db
	// db.View(func(txn *badger.Txn) (err error) {
	// 	gcCount = dbGetUint64(txn, sysKeyGCCount)
	// 	return
	// })

	// latestGC := gcCount * gcLimit
	// if latestGC > deletes {
	// 	// there was a reset should reset in the stats
	// 	gcCount, latestGC = 0, 0
	// }

	// if deletes-latestGC > gcLimit {
	// 	log.Infof("Start maintenance n %d for deletes %d > %d", gcCount, deletes-latestGC, gcLimit)

	// 	log.Info("")

	// 	db.RunValueLogGC(settings.Tuning.DbGCDiscardRation)
	// 	log.Infof("End maintenance n %d for deletes %d > %d", gcCount, deletes-latestGC, gcLimit)
	// 	// update the gcCount
	// 	db.Update(func(txn *badger.Txn) (err error) {
	// 		gcCount++
	// 		dbSetUint64(txn, sysKeyGCCount, gcCount)
	// 		return
	// 	})
	// }
}
