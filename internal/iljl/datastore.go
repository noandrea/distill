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
	rq chan ShortUrlGet
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
	rq = make(chan ShortUrlGet)
}

// CloseSession closes the underling storage
func CloseSession() {
	db.Close()
}

// PreprocessURL preprocess and validate url
func PreprocessURL(url *URLReq, forceAlphabet bool) error {
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
		if !m && forceAlphabet {
			err := fmt.Errorf("ID %v doesn't match alphabet and forceAlphabet is active", url.ID)
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

func UpsertUrl(url URLReq) error {
	err := db.Update(func(txn *badger.Txn) error {
		var err error
		urlData, err := url.MarshalBinary()
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
	return err
}

func DeleteUrl(url URLReq) error {
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(url.ID))
		return err
	})
	return err
}

func GetUrlByID(id string) (url URLInfo) {
	err := db.View(func(txn *badger.Txn) (err error) {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return
		}
		val, err := item.Value()
		if err != nil {
			return
		}
		url = URLInfo{}
		url.UnmarshalBinary(val)
		if err != nil {
			return
		}
		fmt.Printf("The answer is: %s\n", val)
		return
	})
	if err != nil {
		return
	}
	return
}
