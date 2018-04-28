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
	rq chan ShortID
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
	// TODO: load statistics from the database
	// initialize channel for request comunication
	rq = make(chan ShortID)
}

// CloseSession closes the underling storage
func CloseSession() {
	db.Close()
}

// PreprocessURL preprocess and validate url
func PreprocessURL(url *URLReq, forceAlphabet, forceLength bool) error {
	// chech that the target url is a valid url
	if _, err := net.ParseRequestURI(url.URL); err != nil {
		return err
	}
	url.ID = strings.TrimSpace(url.ID)
	// process url id
	if len(url.ID) == 0 {
		url.ID = GenerateID()
	} else {
		p := fmt.Sprintf("[%s]", internal.Config.ShortID.Alphabet)
		m, _ := regexp.MatchString(p, url.ID)
		if forceAlphabet && !m {
			err := fmt.Errorf("ID %v doesn't match alphabet and forceAlphabet is active", url.ID)
			return err
		}
		if forceLength && len(url.ID) != internal.Config.ShortID.Length {
			err := fmt.Errorf("ID %v doesn't match length and forceLength len %v, required %v", url.ID, len(url.ID), internal.Config.ShortID.Length)
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
		if url.TTL > 0 {
			d := time.Duration(url.TTL) * time.Second
			err = txn.SetWithTTL([]byte(url.ID), urlData, d)
		} else {
			err = txn.Set([]byte(url.ID), urlData)
		}
		return err
	})
	return
}

// DeleteURL delete a url mapping
func DeleteURL(id string) (err error) {
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
	return
}

// GetURL get an url mapping by it's id
func GetURL(id string) (url URLInfo, err error) {
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}
		url = URLInfo{}
		err = url.UnmarshalBinary(val)
		return err
	})
	return
}
