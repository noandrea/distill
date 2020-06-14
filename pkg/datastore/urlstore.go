package datastore

import (
	"fmt"
	"time"

	"github.com/noandrea/distill/config"
	"github.com/noandrea/distill/pkg/common"
	"github.com/noandrea/distill/pkg/model"
)

var (
	settings config.Schema
)

// URLDatastore implements the datastore for the short urls
type URLDatastore interface {
	// for general data
	Put(key string, data interface{}) error
	Get(key string, data interface{}) (bool, error)
	// for counters
	CounterSet(key string, val int64) (err error)
	CounterGet(key string) (val int64, err error)
	CounterPlus(key string) (err error)
	CounterMinus(key string) (err error)
	// for URLs
	Hit(key string) (model.URLInfo, error)
	Peek(key string) (model.URLInfo, error)
	Insert(key string, u *model.URLInfo) error
	Upsert(key string, u *model.URLInfo) error
	Delete(key string) error
	// TODO: Backup the database as csv
	// Backup(outFile string) error
	// Restore(inFile string) (int, error)
}

func key(prefix byte, id string) (k []byte, err error) {
	idb := []byte(id)
	idbl := len(idb)
	if idbl == 0 {
		err = fmt.Errorf("Empty id not allowed")
		return
	}
	k = make([]byte, len(idb)+1)
	k[0] = prefix
	copy(k[1:], idb)
	return
}

// UpdateCounters on retrieval from storage
func UpdateCounters(u *model.URLInfo) {
	// increase hit counter
	u.Hits++
	// deal with counter
	now := time.Now()
	if now.After(u.ActiveFrom) && now.Before(u.ExpiresOn) {
		if u.ResolveLimit > 0 {
			u.ResolveCount = common.Min(u.ResolveCount, u.ResolveLimit)
		}
		u.ResolveCount++
	}
}
