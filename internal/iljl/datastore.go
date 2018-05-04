package iljl

import (
	"fmt"
	net "net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jbrodriguez/mlog"
	"gitlab.com/lowgroundandbigshoes/iljl/internal"
)

// UpsertURL insert or udpdate a url mapping
func UpsertURL(url *URLReq, forceAlphabet, forceLength bool) (id string, err error) {
	// preprocess the url and generates the id if necessary
	// chech that the target url is a valid url
	if _, err := net.ParseRequestURI(url.URL); err != nil {
		return "", err
	}
	// set the binding date
	u := &URLInfo{
		BountAt: time.Now(),
		URL:     url.URL,
	}
	// global expiration
	globalExpire := calculateExpiration(u, internal.Config.ShortID.TTL, internal.Config.ShortID.ExpireOn)
	// local expiration
	localExpire := calculateExpiration(u, url.TTL, url.ExpireOn)
	// the expiration with the latest expiration (farthest in the future)
	u.ExpireOn = globalExpire
	if localExpire.After(globalExpire) {
		u.ExpireOn = localExpire
	}
	// set max requests
	u.MaxRequests = internal.Config.ShortID.MaxRequests
	if url.MaxRequests > u.MaxRequests {
		u.MaxRequests = url.MaxRequests
	}
	// cleanup the string id
	u.ID = strings.TrimSpace(url.ID)
	// process url id
	if len(u.ID) == 0 {
		err = Insert(u)
	} else {
		// TODO: check longest allowed key in badger
		p := fmt.Sprintf("[^%s]", regexp.QuoteMeta(internal.Config.ShortID.Alphabet))
		m, _ := regexp.MatchString(p, url.ID)
		if forceAlphabet && m {
			err = fmt.Errorf("ID %v doesn't match alphabet and forceAlphabet is active", url.ID)
			return "", err
		}
		if forceLength && len(url.ID) != internal.Config.ShortID.Length {
			err = fmt.Errorf("ID %v doesn't match length and forceLength len %v, required %v", url.ID, len(url.ID), internal.Config.ShortID.Length)
			return "", err
		}
		err = Upsert(u)
	}

	if err == nil {
		// collect statistics
		pushEvent(&URLOp{
			opcode: opcodeInsert,
			url:    u,
			err:    err,
		})
	}
	return u.ID, err
}

// calculateExpiration calculate the expiration of a url
// returns the highest date betwwen the date binding + ttl and the date expiration date
func calculateExpiration(u *URLInfo, ttl int64, expireDate time.Time) (expire time.Time) {
	if internal.Config.ShortID.TTL > 0 {
		expire = u.BountAt.Add(time.Duration(internal.Config.ShortID.TTL) * time.Second)
	}
	if !internal.Config.ShortID.ExpireOn.After(expire) {
		expire = internal.Config.ShortID.ExpireOn
	}
	return
}

// DeleteURL delete a url mapping
func DeleteURL(id string) (err error) {
	err = Delete(id)
	if err != nil {
		return
	}
	// collect statistics
	pushEvent(&URLOp{
		opcode: opcodeDelete,
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
	urlInfo, err = Get(id)
	if err != nil {
		return
	}

	expired := false
	if time.Now().After(urlInfo.ExpireOn) {
		mlog.Trace("Expire date for %v, limit %v, requests %v", urlInfo.ID, urlInfo.Counter, urlInfo.MaxRequests)
		expired = true
	}
	if urlInfo.Counter > urlInfo.MaxRequests {
		mlog.Trace("Expire max request for %v, limit %v, requests %v", urlInfo.ID, urlInfo.Counter, urlInfo.MaxRequests)
		expired = true
	}

	if expired {
		err = ErrURLExpired // collect statistics
		pushEvent(&URLOp{
			opcode: opcodeExpired,
			url:    urlInfo,
			err:    err,
		})
	}
	return

}
