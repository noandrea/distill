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
	// inintialize statistics
	err = NewStatistics()
	if err != nil {
		mlog.Fatal(err)
	}
}

// CloseSession closes the underling storage
func CloseSession() {
	StopStatistics()
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
	err = db.Update(func(txn *badger.Txn) (err error) {
		err = dbSetBin(txn, keyURL(id), urlInfo)
		if err != nil {
			return err
		}
		// collect statistics
		pushEvent(&URLOp{
			opcode: opcodeInsert,
			url:    urlInfo,
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
		// then delete the keys
		err = dbDel(txn, key)
		if err != nil {
			return
		}
		// collect statistics
		pushEvent(&URLOp{
			opcode: opcodeDelete,
		})
		return err
	})
	return
}

// GetURLRedirect retrieve the redicrect url associated to an id
// it also fire an event of tipe opcodeGet
func GetURLRedirect(id string) (redirectURL string, err error) {
	urlInfo, err := GetURLInfo(id)
	if err != nil {
		return
	}
	// collect statistics
	pushEvent(&URLOp{
		opcode: opcodeGet,
		url:    urlInfo,
		err:    err,
	})
	// return the redirectUrl
	redirectURL = urlInfo.URL
	return
}

// GetURLInfo retrieve the url info associated to an id
func GetURLInfo(id string) (urlInfo *URLInfo, err error) {
	err = db.View(func(txn *badger.Txn) (err error) {
		urlInfo = &URLInfo{}
		err = dbGetBin(txn, keyURL(id), urlInfo)
		if err != nil {
			return
		}
		// validate the ttl
		expired := false
		if urlInfo.TTL > 0 && time.Now().After(urlInfo.ExpirationDate()) {
			expired = true
		}
		mlog.Trace("GetURLInfo() %v ", urlInfo)
		// validate the maxRequest limit
		if urlInfo.MaxRequests > 0 && urlInfo.Counter >= urlInfo.MaxRequests {
			mlog.Trace("Expire max request for %v, limit %v, requests %v", urlInfo.ID, urlInfo.Counter, urlInfo.MaxRequests)
			expired = true
		}
		if expired {
			err = badger.ErrKeyNotFound
			// collect statistics
			pushEvent(&URLOp{
				opcode: opcodeExpired,
				url:    urlInfo,
				err:    err,
			})
		}
		return
	})
	return
}

// dbGet helper functin

func dbDel(txn *badger.Txn, keys ...[]byte) (err error) {
	for _, k := range keys {
		err = txn.Delete(k)
	}
	return
}

func dbSet(txn *badger.Txn, k, val []byte, ttlSeconds int64) (err error) {
	if ttlSeconds > 0 {
		err = txn.SetWithTTL(k, val, time.Duration(ttlSeconds)*time.Second)
	} else {
		err = txn.Set(k, val)
	}
	return
}

func dbSetInt64(txn *badger.Txn, k []byte, val int64) {
	err := txn.Set(k, itoa(val))
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
	return
}

func dbGet(txn *badger.Txn, k []byte) (val []byte, err error) {
	item, err := txn.Get(k)
	if err != nil {
		return
	}
	val, err = item.Value()
	return
}

func dbGetInt64(txn *badger.Txn, k []byte) (i int64) {
	item, err := txn.Get(k)
	if err != nil {
		return
	}
	val, err := item.Value()
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
