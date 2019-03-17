// Package urlstore provides the main functionalities for distill
package urlstore

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/bluele/gcache"
	"github.com/dgraph-io/badger"
	"github.com/jbrodriguez/mlog"
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
	db  *badger.DB
	uc  gcache.Cache
	st  *Statistics
	stM sync.Mutex
)

// NewSession opens the underling storage
func NewSession() {
	// open the badger database
	opts := badger.DefaultOptions
	opts.SyncWrites = true
	opts.Dir = Config.Server.DbPath
	err := os.MkdirAll(Config.Server.DbPath, os.ModePerm)
	if err != nil {
		mlog.Fatal(err)
	}
	opts.ValueDir = Config.Server.DbPath
	db, err = badger.Open(opts)
	if err != nil {
		mlog.Fatal(err)
	}
	// initialzie internal cache
	uc = gcache.New(Config.Tuning.URLCaheSize).
		EvictedFunc(whenRemoved).
		PurgeVisitorFunc(whenRemoved).
		ARC().
		Build()
	// inintialize statistics
	err = LoadStats()
	if err != nil {
		mlog.Fatal(err)
	}
}

// CloseSession closes the underling storage
func CloseSession() {
	SaveStats()
	uc.Purge()
	db.Close()
}

// whenRemoved gets called by the memory cache
// it will check the value, if the value is nil
// means that the key has been deleted
// so it will delete it also from the persistent store
func whenRemoved(key, value interface{}) {
	if value == nil {
		Delete(key.(string))
		return
	}
	ui := value.(*URLInfo)
	Upsert(ui)
}

// SaveStats write the URL's statistics
func SaveStats() (err error) {
	err = db.Update(func(txn *badger.Txn) (err error) {
		// find all the urls
		dbSetUint64(txn, statsKeyGlobalURLCount, st.Urls)
		dbSetUint64(txn, statsKeyGlobalGetCount, st.Gets)
		dbSetUint64(txn, statsKeyGlobalDelCount, st.Deletes)
		dbSetUint64(txn, statsKeyGlobalUpdCount, st.Upserts)
		// update global statistics
		return
	})
	return
}

// LoadStats write the URL's statistics
func LoadStats() (err error) {
	// initialize object
	st = &Statistics{}
	err = db.View(func(txn *badger.Txn) (err error) {
		st.Urls = dbGetUint64(txn, statsKeyGlobalURLCount)
		st.Gets = dbGetUint64(txn, statsKeyGlobalGetCount)
		st.Deletes = dbGetUint64(txn, statsKeyGlobalDelCount)
		st.Upserts = dbGetUint64(txn, statsKeyGlobalUpdCount)
		return
	})
	return
}

// UpdateStats uppdate urls statistics
func UpdateStats(s Statistics) {
	stM.Lock()
	defer stM.Unlock()
	st.Urls += s.Urls
	st.Gets += s.Gets
	st.Deletes += s.Deletes
	st.Upserts += s.Upserts
	st.GetsExpired += s.GetsExpired
	st.LastRequest = s.LastRequest
}

// ResetStats reset global statistcs
func ResetStats() (err error) {
	stM.Lock()
	defer stM.Unlock()
	st = &Statistics{}
	// iterate over the urls
	i := NewURLIterator()
	for i.HasNext() {
		u, err := i.NextURL()
		if err != nil {
			mlog.Warning("Warning looping through the URLs")
		}
		st.Urls++
		st.Upserts++
		st.Gets += u.Counter
	}
	// close the iterator
	i.Close()
	// run the update
	err = SaveStats()
	if err != nil {
		mlog.Warning("Error while rest stats %v", err)
	}
	return
}

// GetStats get the statistics
func GetStats() (s *Statistics) {
	return st
}

// Insert an url into the the urlstore
func Insert(u *URLInfo) (err error) {
	err = db.Update(func(txn *badger.Txn) (err error) {
		u.ID = generateID()
		key := keyURL(u.ID)
		// TODO: need another limit (numeber of retries)
		// TODO: also check the type of error
		for _, err = dbGet(txn, key); err == nil; {
			u.ID = generateID()
			key = keyURL(u.ID)
		}
		err = dbSetBin(txn, key, u)
		return
	})
	return err
}

// Upsert an url into the the urlstore
func Upsert(u *URLInfo) (err error) {
	err = db.Update(func(txn *badger.Txn) (err error) {
		err = dbSetBin(txn, keyURL(u.ID), u)
		return err
	})
	return err
}

// Peek retrive a url without incrementing the counter
func Peek(id string) (u *URLInfo, err error) {
	uic, err := uc.Get(id)
	if err == gcache.KeyNotFoundError {
		mlog.Trace("cache miss for %s", id)
		err = db.View(func(txn *badger.Txn) (err error) {
			u = &URLInfo{}
			ku := keyURL(id)
			err = dbGetBin(txn, ku, u)
			if err != nil {
				return
			}
			return
		})
	} else {
		u = uic.(*URLInfo)
	}
	return
}

// Get an url from the datastore
func Get(id string) (u *URLInfo, err error) {
	u, err = Peek(id)
	if err != nil {
		return
	}
	// increase the counter
	u.Counter++
	uc.Set(id, u)
	return
}

// Delete deletes an url
func Delete(id string) (err error) {
	err = db.Update(func(txn *badger.Txn) (err error) {
		// remove from cache
		uc.Remove(id)
		// remove from storage
		key := keyURL(id)
		// then delete the keys
		err = txn.Delete(key)
		mlog.Trace("Delete() 01 %v", err)
		return err
	})
	mlog.Trace("Delete() 02 %v", err)
	return
}

// Backup the database as csv
func Backup(outFile string) (err error) {
	ext := filepath.Ext(outFile)
	switch ext {
	case backupExtBin:
		// create output file
		fp, err := os.Create(outFile)
		if err != nil {
			return err
		}
		ts, err := db.Backup(fp, 0)
		if err != nil {
			return err
		}
		mlog.Info("Backup completed at %v", ts)
	case backupExtCsv:
		err = db.View(func(txn *badger.Txn) (err error) {
			// create output file
			fp, err := os.Create(outFile)
			if err != nil {
				return
			}
			defer fp.Close()
			// open the csv writer
			csvW := csv.NewWriter(fp)
			defer csvW.Flush()

			// open the iterator
			opts := badger.DefaultIteratorOptions
			opts.PrefetchSize = Config.Tuning.BckCSVIterPrefetchSize
			opts.PrefetchValues = true
			it := txn.NewIterator(opts)
			defer it.Close()

			p := []byte{keyURLPrefix}
			for it.Seek(p); it.ValidForPrefix(p); it.Next() {
				// retrieve values
				v, err := it.Item().Value()
				if err != nil {
					break
				}
				u := &URLInfo{}
				u.UnmarshalBinary(v)
				err = csvW.Write(u.MarshalRecord())
				if err != nil {
					break
				}
			}
			return
		})
	default:
		err = fmt.Errorf("Unrecoginzed backup format %v", ext)
		mlog.Warning("Unrecoginzed backup format %v", ext)
	}
	return
}

// Restore the database from a backup file
func Restore(inFile string) (count int, err error) {
	ext := filepath.Ext(inFile)
	switch ext {
	case backupExtBin:
		fp, err := os.Open(inFile)
		if err != nil {
			return 0, err
		}
		db.Load(fp)
		fp.Close()
	case backupExtCsv:
		fp, err := os.Open(inFile)
		if err != nil {
			return 0, err
		}
		csvR := csv.NewReader(fp)
		for {
			record, err := csvR.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			u := &URLInfo{}
			if err = u.UnmarshalRecord(record); err != nil {
				break
			}
			if err = Upsert(u); err != nil {
				break
			}
			count++
		}
		fp.Close()
	default:
		err = fmt.Errorf("Unrecoginzed backup format %v", ext)
		mlog.Warning("Unrecoginzed backup format %v", ext)
	}
	return
}

// NewURLIterator return an url iterator over the database
func NewURLIterator() *URLIterator {
	txn := db.NewTransaction(false)
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	px := []byte{keyURLPrefix}
	it.Seek(px)
	return &URLIterator{
		Transaction: txn,
		Iterator:    it,
		KeyPrefix:   px,
	}
}

// URLIterator an iterator over URLs
type URLIterator struct {
	Transaction *badger.Txn
	Iterator    *badger.Iterator
	KeyPrefix   []byte
}

// HasNext tells if there are still elements in the list
func (i *URLIterator) HasNext() bool {
	i.Iterator.Next()
	return i.Iterator.ValidForPrefix(i.KeyPrefix)
}

// NextURL get the next URL from the iterator
func (i *URLIterator) NextURL() (u *URLInfo, err error) {
	v, err := i.Iterator.Item().Value()
	if err != nil {
		return
	}
	u = &URLInfo{}
	err = u.UnmarshalBinary(v)
	return
}

// Close the iterator
func (i *URLIterator) Close() {
	i.Iterator.Close()
	i.Transaction.Discard()
}

// Helper functions

// dbGet helper functin
func dbDel(txn *badger.Txn, keys ...[]byte) (err error) {
	for _, k := range keys {
		err = txn.Delete(k)
		mlog.Trace("dbDel write %s", k)
	}
	return
}

func dbSetUint64(txn *badger.Txn, k []byte, val uint64) {
	err := txn.Set(k, itoa(val))
	mlog.Trace("dbSetInt64 write %s", k)
	if err != nil {
		mlog.Warning("update key %X error %v ", k, err)
	}
}

func dbSetBin(txn *badger.Txn, k []byte, val BinSerializable) (err error) {
	binData, err := val.MarshalBinary()
	if err != nil {
		return
	}
	err = txn.Set(k, binData)
	mlog.Trace("dbSetBin write %s", k)
	return
}

func dbGet(txn *badger.Txn, k []byte) (val []byte, err error) {
	item, err := txn.Get(k)
	if err != nil {
		return
	}
	val, err = item.Value()
	mlog.Trace("dbGet read %s v:%d ", k, item.Version())
	return
}

func dbGetUint64(txn *badger.Txn, k []byte) (i uint64) {
	item, err := txn.Get(k)
	if err != nil {
		return
	}
	val, err := item.Value()
	mlog.Trace("dbGetInt64 read %s v:%d ", k, item.Version())
	if err != nil {
		val = numberZero
	}
	i = atoi(val)
	return
}

func dbGetBin(txn *badger.Txn, k []byte, val BinSerializable) (err error) {
	v, err := dbGet(txn, k)
	if err != nil {
		return err
	}
	err = val.UnmarshalBinary(v)
	return
}
