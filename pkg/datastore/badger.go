package datastore

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/bluele/gcache"
	"github.com/dgraph-io/badger"
	"github.com/noandrea/distill/config"
	"github.com/noandrea/distill/pkg/model"
	log "github.com/sirupsen/logrus"
)

type EmbedStore struct {
	db  *badger.DB
	uc  gcache.Cache
	st  *model.Statistics
	stM sync.Mutex
}

var store *EmbedStore

//NewEmbedStore opens the underling storage
func NewEmbedStore(cfg config.DatastoreConfig) *EmbedStore {
	store = &EmbedStore{}
	dsSettings := cfg
	// open the badger database
	opts := badger.DefaultOptions(dsSettings.URI)
	opts.SyncWrites = true
	// opts.Dir = dsSettings.URI TODO: this should not be necessary anymore
	err := os.MkdirAll(dsSettings.URI, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	opts.ValueDir = dsSettings.URI
	store.db, err = badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	// initialize internal cache
	store.uc = gcache.New(settings.Tuning.URLCacheSize).
		EvictedFunc(whenRemoved).
		PurgeVisitorFunc(whenRemoved).
		ARC().
		Build()
	// initialize statistics
	err = store.LoadStats()
	if err != nil {
		log.Fatal(err)
	}
	return store
}

// CloseSession closes the underling storage
func (store *EmbedStore) Close() {
	store.SaveStats()
	store.uc.Purge()
	store.db.Close()
}

// UpdateStats uppdate urls statistics
func (store *EmbedStore) UpdateStats(s model.Statistics) {
	store.stM.Lock()
	defer store.stM.Unlock()
	store.st.Urls += s.Urls
	store.st.Gets += s.Gets
	store.st.Deletes += s.Deletes
	store.st.Upserts += s.Upserts
	store.st.GetsExpired += s.GetsExpired
	store.st.LastRequest = s.LastRequest
}

// ResetStats reset global statistcs
func (store *EmbedStore) ResetStats() (err error) {
	store.stM.Lock()
	defer store.stM.Unlock()
	store.st = &model.Statistics{}
	// iterate over the urls
	i := NewURLIterator()
	for i.HasNext() {
		u, err := i.NextURL()
		if err != nil {
			log.Warning("Warning looping through the URLs")
		}
		store.st.Urls++
		store.st.Upserts++
		store.st.Gets += u.Counter
	}
	// close the iterator
	i.Close()
	// run the update
	err = store.SaveStats()
	if err != nil {
		log.Warning("Error while rest stats:", err)
	}
	return
}

// GetStats get the statistics
func (store *EmbedStore) GetStats() (s *model.Statistics) {
	return store.st
}

// Insert an url into the url store
func (store *EmbedStore) Insert(u *model.URLInfo) (err error) {
	err = store.db.Update(func(txn *badger.Txn) (err error) {
		u.ID = generateID(settings.ShortID.Alphabet, settings.ShortID.Length)
		// generateID always return a valid id
		key, _ := keyURL(u.ID)
		// TODO: need another limit (numeber of retries)
		// TODO: also check the type of error
		for _, err = dbGet(txn, key); err == nil; {
			u.ID = generateID(settings.ShortID.Alphabet, settings.ShortID.Length)
			// generateID always return a valid id
			key, _ = keyURL(u.ID)
		}
		err = dbSetBin(txn, key, u)
		return
	})
	return err
}

// Upsert an url into the the urlstore
func (store *EmbedStore) Upsert(u *model.URLInfo) (err error) {
	err = store.db.Update(func(txn *badger.Txn) (err error) {
		key, err := keyURL(u.ID)
		if err != nil {
			return
		}
		err = dbSetBin(txn, key, u)
		return
	})
	return err
}

// Peek retrive a url without incrementing the counter
func (store *EmbedStore) Peek(id string) (u *model.URLInfo, err error) {
	uic, err := store.uc.Get(id)
	if err == gcache.KeyNotFoundError {
		log.Trace("cache miss for ", id)
		err = store.db.View(func(txn *badger.Txn) (err error) {
			u = &model.URLInfo{}
			ku, err := keyURL(id)
			if err != nil {
				return
			}
			err = dbGetBin(txn, ku, u)
			if err != nil {
				return
			}
			return
		})
	} else {
		u = uic.(*model.URLInfo)
	}
	return
}

// Get an url from the datastore
func (store *EmbedStore) Get(id string) (u *model.URLInfo, err error) {
	u, err = store.Peek(id)
	if err != nil {
		return
	}
	// increase the counter
	u.Counter++
	store.uc.Set(id, u)
	return
}

// Delete deletes an url
func (store *EmbedStore) Delete(id string) (err error) {
	err = store.db.Update(func(txn *badger.Txn) (err error) {
		// remove from cache
		store.uc.Remove(id)
		// remove from storage
		key, err := keyURL(id)
		if err != nil {
			return
		}
		// then delete the keys
		err = txn.Delete(key)
		log.Trace("Delete() 01:", err)
		return err
	})
	log.Trace("Delete() 02:", err)
	return
}

// Backup the database as csv
func (store *EmbedStore) Backup(outFile string) (err error) {
	ext := filepath.Ext(outFile)
	switch ext {
	case backupExtBin:
		// create output file
		fp, err := os.Create(outFile)
		if err != nil {
			return err
		}
		ts, err := store.db.Backup(fp, 0)
		if err != nil {
			return err
		}
		log.Info("Backup completed at ", ts)
	case backupExtCsv:
		err = store.db.View(func(txn *badger.Txn) (err error) {
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
			opts.PrefetchSize = settings.Tuning.BckCSVIterPrefetchSize
			opts.PrefetchValues = true
			it := txn.NewIterator(opts)
			defer it.Close()

			p := []byte{keyURLPrefix}
			for it.Seek(p); it.ValidForPrefix(p); it.Next() {
				// retrieve values
				err := it.Item().Value(func(v []byte) error {
					u := &model.URLInfo{}
					u.UnmarshalBinary(v)
					return csvW.Write(u.MarshalRecord())
				})
				if err != nil {
					break
				}
			}
			return
		})
	default:
		err = fmt.Errorf("Unrecognized backup format %v", ext)
		log.Warning("Unrecognized backup format:", ext)
	}
	return
}

// Restore the database from a backup file
func (store *EmbedStore) Restore(inFile string) (count int, err error) {
	ext := filepath.Ext(inFile)
	switch ext {
	case backupExtBin:
		fp, err := os.Open(inFile)
		if err != nil {
			return 0, err
		}
		store.db.Load(fp, 16)
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
			u := &model.URLInfo{}
			if err = u.UnmarshalRecord(record); err != nil {
				break
			}
			if err = store.Upsert(u); err != nil {
				break
			}
			count++
		}
		fp.Close()
	default:
		err = fmt.Errorf("Unrecognized backup format %v", ext)
		log.Warning("Unrecognized backup format:", ext)
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
func (i *URLIterator) NextURL() (u *model.URLInfo, err error) {
	u = &model.URLInfo{}
	err = i.Iterator.Item().Value(func(v []byte) error {
		return u.UnmarshalBinary(v)
	})
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
		log.Trace("dbDel write:", k)
	}
	return
}

func dbSetUint64(txn *badger.Txn, k []byte, val uint64) {
	err := txn.Set(k, itoa(val))
	log.Trace("dbSetInt64 write:", k)
	if err != nil {
		log.Warningf("update key %X error %v ", k, err)
	}
}

func dbSetBin(txn *badger.Txn, k []byte, val model.BinSerializable) (err error) {
	binData, err := val.MarshalBinary()
	if err != nil {
		return
	}
	err = txn.Set(k, binData)
	log.Trace("dbSetBin write:", k)
	return
}

func dbGet(txn *badger.Txn, k []byte) (val []byte, err error) {
	item, err := txn.Get(k)
	if err != nil {
		return
	}
	err = item.Value(func(v []byte) error {
		val = v
		return nil
	})
	log.Tracef("dbGet read %s v:%d ", k, item.Version())
	return
}

func dbGetUint64(txn *badger.Txn, k []byte) (i uint64) {
	item, err := txn.Get(k)
	if err != nil {
		return
	}
	val := numberZero
	err = item.Value(func(v []byte) error {
		val = v
		return nil
	})
	log.Tracef("dbGetInt64 read %s v:%d ", k, item.Version())
	i = atoi(val)
	return
}

func dbGetBin(txn *badger.Txn, k []byte, val model.BinSerializable) (err error) {
	v, err := dbGet(txn, k)
	if err != nil {
		return err
	}
	err = val.UnmarshalBinary(v)
	return
}
