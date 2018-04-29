package iljl

import (
	"fmt"
	net "net/url"
	"regexp"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/jbrodriguez/mlog"
	"gitlab.com/lowgroundandbigshoes/iljl/internal"
)

var (
	db *badger.DB
)

// NewSession opens the underling storage
func NewSession() {
	// open the badger database
	opts := badger.DefaultOptions
	opts.Dir = internal.Config.Server.DbPath
	opts.ValueDir = internal.Config.Server.DbPath
	var err error
	db, err = badger.Open(opts)
	if err != nil {
		mlog.Fatal(err)
	}
	// initialize the worker pool
	if internal.Config.Server.EnableStats {
		err = NewStatistics()
		if err != nil {
			mlog.Fatal(err)
			//panic(fmt.Sprintf("cannot start staistics %v", err))
		}
	}
}

// CloseSession closes the underling storage
func CloseSession() {
	db.Close()
}

// PreprocessURL preprocess and validate url
func PreprocessURL(url *URLReq, forceAlphabet, forceLength bool) (err error) {
	// chech that the target url is a valid url
	if _, err := net.ParseRequestURI(url.URL); err != nil {
		return err
	}
	url.ID = strings.TrimSpace(url.ID)
	// process url id
	if len(url.ID) == 0 {
		url.ID = generateID()
	} else {
		p := fmt.Sprintf("[^%s]", regexp.QuoteMeta(internal.Config.ShortID.Alphabet))
		m, _ := regexp.MatchString(p, url.ID)
		if forceAlphabet && m {
			err = fmt.Errorf("ID %v doesn't match alphabet and forceAlphabet is active", url.ID)
			return err
		}
		if forceLength && len(url.ID) != internal.Config.ShortID.Length {
			err = fmt.Errorf("ID %v doesn't match length and forceLength len %v, required %v", url.ID, len(url.ID), internal.Config.ShortID.Length)
			return err
		}
	}
	// default ttl
	if url.TTL == 0 {
		url.TTL = internal.Config.ShortID.TTL
	}
	// default max requests
	if url.MaxRequests == 0 {
		url.MaxRequests = internal.Config.ShortID.MaxRequests
	}

	return nil
}

// UpsertURL insert or udpdate a url mapping
func UpsertURL(url *URLReq, forceAlphabet, forceLength bool) (id string, err error) {
	// preprocess the url and generates the id if necessary
	err = PreprocessURL(url, forceAlphabet, forceLength)
	if err != nil {
		return
	}
	// save the id
	id = url.ID
	urlInfo := &URLInfo{
		ID:          url.ID,
		URL:         url.URL,
		TTL:         url.TTL,
		Counter:     0,
		MaxRequests: url.MaxRequests,
		BountAt:     time.Now(),
	}
	// insert/update the mapping
	err = db.Update(func(txn *badger.Txn) error {
		var err error
		urlData, err := urlInfo.MarshalBinary()
		if err != nil {
			return err
		}
		if urlInfo.TTL > 0 {
			err = txn.SetWithTTL(keyURL(id), urlData, ttl(url.TTL))
		} else {
			err = txn.Set(keyURL(id), urlData)
		}
		// collect statistics
		pushEvent(&URLOp{
			opcode: opcodeInsert,
			url:    *urlInfo,
			err:    err,
		})

		return err
	})
	return
}

// DeleteURL delete a url mapping
func DeleteURL(id string) (err error) {
	err = db.Update(func(txn *badger.Txn) (err error) {
		key := keyURL(id)
		// first check if the url exists
		_, err = dbGet(txn, key)
		if err == badger.ErrKeyNotFound {
			return
		}
		err = txn.Delete(key)
		if err != nil {
			return
		}
		// delete the counter and discard the error
		txn.Delete(keyURLStatCount(id))
		// collect statistics
		pushEvent(&URLOp{
			opcode: opcodeDelete,
		})
		return err
	})
	return
}

// GetURL get an url mapping by it's id
func GetURL(id string, withStats bool) (url URLInfo, err error) {
	err = db.View(func(txn *badger.Txn) (err error) {
		val, err := dbGet(txn, keyURL(id))
		if err != nil {
			return err
		}
		url = URLInfo{}
		err = url.UnmarshalBinary(val)
		// if is without statistics push the event and return
		if !withStats {
			// collect statistics
			pushEvent(&URLOp{
				opcode: opcodeGet,
				url:    url,
				err:    err,
			})

			return
		}
		// if it is with statistics retrieve the statitsts
		val, err = dbGet(txn, keyURLStatCount(id))
		if err != nil {
			val = numberZero
		}
		url.Counter = atoi(val) + 1
		return

	})
	return
}

// dbGet helper functin
func dbGet(txn *badger.Txn, k []byte) (val []byte, err error) {
	item, err := txn.Get(k)
	if err != nil {
		return
	}
	val, err = item.Value()
	return
}
