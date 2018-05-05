package iljl

import (
	"github.com/bluele/gcache"
	"github.com/dgraph-io/badger"
	"github.com/jbrodriguez/mlog"
	"gitlab.com/lowgroundandbigshoes/iljl/internal"
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
