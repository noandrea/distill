package distill

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bluele/gcache"
	"github.com/dgraph-io/badger"
	"github.com/jbrodriguez/mlog"
	"gitlab.com/welance/distill/internal"
)

const (
	backupExtBin = ".bin"
	backupExtCsv = ".csv"
)

var (
	db *badger.DB
	uc gcache.Cache
)

// NewSession opens the underling storage
func NewSession() {
	// open the badger database
	opts := badger.DefaultOptions
	opts.SyncWrites = true
	opts.Dir = internal.Config.Server.DbPath
	opts.ValueDir = internal.Config.Server.DbPath
	var err error
	db, err = badger.Open(opts)
	if err != nil {
		mlog.Fatal(err)
	}
	// initialzie internal cache
	uc = gcache.New(internal.Config.Tuning.URLCaheSize).
		EvictedFunc(whenRemoved).
		PurgeVisitorFunc(whenRemoved).
		ARC().
		Build()
	// inintialize statistics
	err = NewStatistics()
	if err != nil {
		mlog.Fatal(err)
	}
}

// CloseSession closes the underling storage
func CloseSession() {
	StopStatistics()
	uc.Purge()
	db.Close()
}

func whenRemoved(key, value interface{}) {
	ui := value.(*URLInfo)
	Upsert(ui)
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
			opts.PrefetchSize = internal.Config.Tuning.BckCSVIterPrefetchSize
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
func Restore(inFile string) (err error) {
	ext := filepath.Ext(inFile)
	switch ext {
	case backupExtBin:
		fp, err := os.Open(inFile)
		if err != nil {
			return err
		}
		db.Load(fp)
		fp.Close()
	case backupExtCsv:
		fp, err := os.Open(inFile)
		if err != nil {
			return err
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
		}
		fp.Close()
	default:
		err = fmt.Errorf("Unrecoginzed backup format %v", ext)
		mlog.Warning("Unrecoginzed backup format %v", ext)
	}
	return
}

// Helper fucntions

// dbGet helper functin
func dbDel(txn *badger.Txn, keys ...[]byte) (err error) {
	for _, k := range keys {
		err = txn.Delete(k)
		mlog.Trace("dbDel write %s", k)
	}
	return
}

func dbSetInt64(txn *badger.Txn, k []byte, val int64) {
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
	mlog.Trace("dbGet read %s v:%d ", k, item.Version)
	return
}

func dbGetInt64(txn *badger.Txn, k []byte) (i int64) {
	item, err := txn.Get(k)
	if err != nil {
		return
	}
	val, err := item.Value()
	mlog.Trace("dbGetInt64 read %s v:%d ", k, item.Version)
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
