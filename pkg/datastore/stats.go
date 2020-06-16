package datastore

import (
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/noandrea/distill/pkg/model"
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
