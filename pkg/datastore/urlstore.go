package datastore

import (
	"fmt"

	"github.com/noandrea/distill/config"
	"github.com/noandrea/distill/pkg/model"
)

const (
	backupExtBin = ".bin"
	backupExtCsv = ".csv"
)

var (
	// initialize stats keys
	statsKeyGlobalURLCount = keyGlobalStat("distill_global_url_count")
	statsKeyGlobalGetCount = keyGlobalStat("distill_global_get_count")
	statsKeyGlobalDelCount = keyGlobalStat("distill_global_del_count")
	statsKeyGlobalUpdCount = keyGlobalStat("distill_global_upd_count")
)

var (
	settings config.Schema
)

// URLDatastore implements the datastore for the short urls
type URLDatastore interface {
	Close()
	// for items
	Insert(u *model.URLInfo) (err error)
	Upsert(u *model.URLInfo) (err error)
	Peek(id string) (u *model.URLInfo, err error)
	Get(id string) (u *model.URLInfo, err error)
	Delete(id string) (err error)
	// Backup the database as csv
	Backup(outFile string) (err error)
	Restore(inFile string) (count int, err error)
}

//   ___  ____   ________  ____  ____   ______
//  |_  ||_  _| |_   __  ||_  _||_  _|.' ____ \
//    | |_/ /     | |_ \_|  \ \  / /  | (___ \_|
//    |  __'.     |  _| _    \ \/ /    _.____`.
//   _| |  \ \_  _| |__/ |   _|  |_   | \____) |
//  |____||____||________|  |______|   \______.'
//

func keyURL(id string) (k []byte, err error) {
	return key(keyURLPrefix, id)
}

func keySys(id string) (k []byte) {
	k, _ = key(keySysPrefix, id)
	return
}

func keyGlobalStat(id string) (k []byte) {
	k, _ = key(keyStatPrefix, id)
	return
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
