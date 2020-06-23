package datastore

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/noandrea/distill/config"
	"github.com/noandrea/distill/pkg/common"
	"github.com/noandrea/distill/pkg/model"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack"
)

type EmbedStore struct {
	db *badger.DB
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
	return store
}

// Close closes the underling storage
func (es *EmbedStore) Close() {
	es.db.Close()
}

// Put store arbitrary data (serialized in json)
func (es *EmbedStore) Put(key string, data interface{}) (err error) {
	err = es.db.Update(func(txn *badger.Txn) (err error) {
		// marshal and store the data
		v, err := msgpack.Marshal(&data)
		if err != nil {
			return
		}
		return txn.Set([]byte(key), v)
	})
	return
}

// Get get existing data
func (es *EmbedStore) Get(key string, data interface{}) (found bool, err error) {
	err = es.db.View(func(txn *badger.Txn) (err error) {
		v, err := dbGet(txn, []byte(key))
		if err != nil {
			return
		}
		found = true
		err = msgpack.Unmarshal(v, data)
		return
	})
	return
}

// CounterSet set a counter value
func (es *EmbedStore) CounterSet(key string, val int64) (err error) {
	err = es.db.Update(func(txn *badger.Txn) (err error) {
		k := []byte(key)
		// set the value
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(val))
		err = txn.Set(k, b)
		return
	})
	return
}

// CounterGet get a counter value
func (es *EmbedStore) CounterGet(key string) (val int64, err error) {
	err = es.db.View(func(txn *badger.Txn) (err error) {
		b, err := dbGet(txn, []byte(key))
		if err != nil {
			log.Warn("key ", key, " not found, counter set to 0")
			b = make([]byte, 8)
		}
		val = int64(binary.LittleEndian.Uint64(b))
		return
	})
	return
}

// CounterPlus increase a counter
func (es *EmbedStore) CounterPlus(key string) (err error) {
	err = es.db.Update(func(txn *badger.Txn) (err error) {
		k := []byte(key)
		// get the value
		b, err := dbGet(txn, k)
		if err != nil {
			log.Warn("key ", key, " not found, counter set to 0")
		}
		// read and update
		counter := int64(binary.LittleEndian.Uint64(b))
		counter++
		// now set
		binary.LittleEndian.PutUint64(b, uint64(counter))
		err = txn.Set(k, b)
		return
	})
	return
}

// CounterMinus decrease a counter
func (es *EmbedStore) CounterMinus(key string) (err error) {
	err = es.db.Update(func(txn *badger.Txn) (err error) {
		k := []byte(key)
		// get the value
		b, err := dbGet(txn, k)
		if err != nil {
			log.Warn("key ", key, " not found, counter set to 0")
		}
		// read and update
		counter := int64(binary.LittleEndian.Uint64(b))
		counter--
		// now set
		binary.LittleEndian.PutUint64(b, uint64(counter))
		err = txn.Set(k, b)
		return
	})
	return
}

// Hit hit an url, return the url that has been hit (with updated hit counter inclusive)
func (es *EmbedStore) Hit(key string) (u model.URLInfo, err error) {
	err = es.db.Update(func(txn *badger.Txn) (err error) {
		// key to byte slice
		k := []byte(key)
		// get the element
		v, err := dbGet(txn, k)
		if err != nil {
			return
		}
		// unmarshal the result
		err = msgpack.Unmarshal(v, &u)
		if err != nil {
			return
		}
		// this is ugly, the problem is to update
		// correctly the counters for inactive urls
		// this should be used by every implementation
		// for the storage
		UpdateCounters(&u)
		// marshal the update
		data, err := msgpack.Marshal(&u)
		if err != nil {
			return
		}
		// execute the update
		err = txn.Set(k, data)
		return
	})
	return
}

// Peek get an url without updating the hit count
func (es *EmbedStore) Peek(key string) (u model.URLInfo, err error) {
	err = es.db.View(func(txn *badger.Txn) (err error) {
		v, err := dbGet(txn, []byte(key))
		if err != nil {
			return
		}
		err = msgpack.Unmarshal(v, &u)
		return
	})
	return
}

// Insert an url into the url store
func (es *EmbedStore) Insert(key string, u *model.URLInfo) (err error) {
	err = es.db.Update(func(txn *badger.Txn) (err error) {
		k := []byte(key)
		// if the key exists return error
		if _, e := dbGet(txn, k); e == nil {
			err = fmt.Errorf("duplicated key %s", key)
			return
		}
		v, err := msgpack.Marshal(u)
		if err != nil {
			return
		}
		err = txn.Set(k, v)
		return
	})
	return err
}

// Upsert an url into the the urlstore
func (es *EmbedStore) Upsert(key string, u *model.URLInfo) (err error) {
	err = es.db.Update(func(txn *badger.Txn) (err error) {
		k := []byte(key)
		v, err := msgpack.Marshal(u)
		if err != nil {
			return
		}
		err = txn.Set(k, v)
		return
	})
	return err
}

// Delete deletes an url
func (es *EmbedStore) Delete(key string) (err error) {
	err = es.db.Update(func(txn *badger.Txn) (err error) {
		// then delete the keys
		err = txn.Delete([]byte(key))
		log.Debug("Delete(", key, ") error:", err)
		return err
	})
	log.Trace("Delete() 02:", err)
	return
}

// Backup the database as csv
func (es *EmbedStore) Backup(outFile string) (err error) {
	// ext := filepath.Ext(outFile)
	// switch ext {
	// case backupExtBin:
	// 	// create output file
	// 	fp, err := os.Create(outFile)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	ts, err := es.db.Backup(fp, 0)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	log.Info("Backup completed at ", ts)
	// case backupExtCsv:
	// 	err = store.db.View(func(txn *badger.Txn) (err error) {
	// 		// create output file
	// 		fp, err := os.Create(outFile)
	// 		if err != nil {
	// 			return
	// 		}
	// 		defer fp.Close()
	// 		// open the csv writer
	// 		csvW := csv.NewWriter(fp)
	// 		defer csvW.Flush()

	// 		// open the iterator
	// 		opts := badger.DefaultIteratorOptions
	// 		opts.PrefetchSize = settings.Tuning.BckCSVIterPrefetchSize
	// 		opts.PrefetchValues = true
	// 		it := txn.NewIterator(opts)
	// 		defer it.Close()

	// 		p := []byte{keyURLPrefix}
	// 		for it.Seek(p); it.ValidForPrefix(p); it.Next() {
	// 			// retrieve values
	// 			err := it.Item().Value(func(v []byte) error {
	// 				u := &model.URLInfo{}
	// 				u.UnmarshalBinary(v)
	// 				return csvW.Write(u.MarshalRecord())
	// 			})
	// 			if err != nil {
	// 				break
	// 			}
	// 		}
	// 		return
	// 	})
	// default:
	// 	err = fmt.Errorf("Unrecognized backup format %v", ext)
	// 	log.Warning("Unrecognized backup format:", ext)
	// }
	return
}

// Restore the database from a backup file
func (es *EmbedStore) Restore(inFile string) (count int, err error) {
	// ext := filepath.Ext(inFile)
	// switch ext {
	// case backupExtBin:
	// 	fp, err := os.Open(inFile)
	// 	if err != nil {
	// 		return 0, err
	// 	}
	// 	es.db.Load(fp, 16)
	// 	fp.Close()
	// case backupExtCsv:
	// 	fp, err := os.Open(inFile)
	// 	if err != nil {
	// 		return 0, err
	// 	}
	// 	csvR := csv.NewReader(fp)
	// 	for {
	// 		record, err := csvR.Read()
	// 		if err == io.EOF {
	// 			break
	// 		}
	// 		if err != nil {
	// 			break
	// 		}
	// 		u := &model.URLInfo{}
	// 		if err = u.UnmarshalRecord(record); err != nil {
	// 			break
	// 		}
	// 		if err = store.Upsert(u); err != nil {
	// 			break
	// 		}
	// 		count++
	// 	}
	// 	fp.Close()
	// default:
	// 	err = fmt.Errorf("Unrecognized backup format %v", ext)
	// 	log.Warning("Unrecognized backup format:", ext)
	// }
	return
}

// NewURLIterator return an url iterator over the database
// func (es *EmbedStore) NewURLIterator() *URLIterator {
// 	// txn := es.db.NewTransaction(false)
// 	// it := txn.NewIterator(badger.DefaultIteratorOptions)
// 	// px := []byte{keyURLPrefix}
// 	// it.Seek(px)
// 	// return &URLIterator{
// 	// 	Transaction: txn,
// 	// 	Iterator:    it,
// 	// 	KeyPrefix:   px,
// 	// }
// }

// // URLIterator an iterator over URLs
// type URLIterator struct {
// 	Transaction *badger.Txn
// 	Iterator    *badger.Iterator
// 	KeyPrefix   []byte
// }

// // HasNext tells if there are still elements in the list
// func (i *URLIterator) HasNext() bool {
// 	i.Iterator.Next()
// 	return i.Iterator.ValidForPrefix(i.KeyPrefix)
// }

// // NextURL get the next URL from the iterator
// func (i *URLIterator) NextURL() (u *model.URLInfo, err error) {
// 	u = &model.URLInfo{}
// 	err = i.Iterator.Item().Value(func(v []byte) error {
// 		return u.UnmarshalBinary(v)
// 	})
// 	return
// }

// // Close the iterator
// func (i *URLIterator) Close() {
// 	i.Iterator.Close()
// 	i.Transaction.Discard()
// }

// Helper functions

// dbGet helper functin
func dbDel(txn *badger.Txn, keys ...[]byte) (err error) {
	for _, k := range keys {
		err = txn.Delete(k)
		log.Trace("dbDel write:", k)
	}
	return
}

// dbGet retrieve the value of a key
// ErrKeyNotFound is returned if the key is not found
func dbGet(txn *badger.Txn, k []byte) (val []byte, err error) {
	item, err := txn.Get(k)
	if err != nil {
		return
	}
	err = item.Value(func(v []byte) error {
		val = v
		return nil
	})
	log.Debug("dbGet read ", k, " v:", item.Version())
	return
}

func dbGetUint64(txn *badger.Txn, k []byte) (i uint64) {
	item, err := txn.Get(k)
	if err != nil {
		return
	}
	val := []byte{0}
	err = item.Value(func(v []byte) error {
		val = v
		return nil
	})
	log.Tracef("dbGetInt64 read %s v:%d ", k, item.Version())
	i = common.Atoi(val)
	return
}
