package distill

import (
	"encoding/csv"
	"fmt"
	"io"
	net "net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jbrodriguez/mlog"
	gonanoid "github.com/matoous/go-nanoid"
	"gitlab.com/welance/oss/distill/internal"
)

func randomString(alphabet string, length int) (string, error) {
	return gonanoid.Generate(alphabet, length)
}

// generateID generates a new id
func generateID() (shortID string) {
	a := internal.Config.ShortID.Alphabet
	l := internal.Config.ShortID.Length
	// a and l are validated before
	shortID, _ = gonanoid.Generate(a, l)
	return
}

// UpsertURLSimple insert or updae an url
// shortcut for UpsertURL(true, true, time.Now())
func UpsertURLSimple(url *URLReq) (id string, err error) {
	return UpsertURL(url, true, true, time.Now())
}

// UpsertURL insert or udpdate a url mapping
func UpsertURL(url *URLReq, forceAlphabet, forceLength bool, boundAt time.Time) (id string, err error) {
	// preprocess the url and generates the id if necessary
	// chech that the target url is a valid url
	if _, err = net.ParseRequestURI(url.URL); err != nil {
		mlog.Info("%s", url.URL)
		mlog.Error(err)
		return
	}
	// set the binding date
	u := &URLInfo{
		BountAt: boundAt,
		URL:     url.URL,
		TTL:     url.TTL,
	}
	// the local expiration always take priority
	u.ExpireOn = calculateExpiration(u, url.TTL, url.ExpireOn)
	if u.ExpireOn.IsZero() {
		// global expiration
		u.ExpireOn = calculateExpiration(u, internal.Config.ShortID.TTL, internal.Config.ShortID.ExpireOn)
	}
	// set max requests, the local version always has priority
	u.MaxRequests = url.MaxRequests
	if u.MaxRequests == 0 {
		u.MaxRequests = internal.Config.ShortID.MaxRequests
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
			ID:     u.ID,
			err:    err,
		})
	}
	return u.ID, err
}

// calculateExpiration calculate the expiration of a url
// returns the highest date betwwen the date binding + ttl and the date expiration date
func calculateExpiration(u *URLInfo, ttl int64, expireDate time.Time) (expire time.Time) {
	if ttl > 0 {
		expire = u.BountAt.Add(time.Duration(ttl) * time.Second)
	}
	if expireDate.After(expire) {
		expire = expireDate
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
		ID:     id,
	})
	return
}

// GetURLRedirect retrieve the redicrect url associated to an id
// it also fire an event of tipe opcodeGet
func GetURLRedirect(id string) (redirectURL string, err error) {
	urlInfo, err := Get(id)
	if err != nil {
		return
	}
	expired := false
	if !urlInfo.ExpireOn.IsZero() && time.Now().After(urlInfo.ExpireOn) {
		mlog.Trace("Expire date for %v, limit %v, requests %v", urlInfo.ID, urlInfo.Counter, urlInfo.MaxRequests)
		expired = true
	}
	if urlInfo.MaxRequests > 0 && urlInfo.Counter > urlInfo.MaxRequests {
		mlog.Trace("Expire max request for %v, limit %v, requests %v", urlInfo.ID, urlInfo.Counter, urlInfo.MaxRequests)
		expired = true
	}

	if expired {
		err = ErrURLExpired // collect statistics
		pushEvent(&URLOp{
			opcode: opcodeExpired,
			ID:     urlInfo.ID,
			err:    err,
		})
		return
	}
	// collect statistics
	pushEvent(&URLOp{
		opcode: opcodeGet,
		ID:     urlInfo.ID,
		err:    err,
	})
	// return the redirectUrl
	redirectURL = urlInfo.URL
	return
}

// GetURLInfo retrieve the url info associated to an id
func GetURLInfo(id string) (urlInfo *URLInfo, err error) {
	urlInfo, err = Peek(id)
	return
}

// ImportCSV import urls from a csv file
func ImportCSV(inFile string) (rows int, err error) {
	fp, err := os.Open(inFile)
	if err != nil {
		return
	}
	defer fp.Close()
	start := time.Now()
	csvR := csv.NewReader(fp)
	for {
		record, err := csvR.Read()
		if err == io.EOF {
			mlog.Error(err)
			break
		}
		if err != nil {
			mlog.Error(err)
			break
		}
		if rows == 0 && record[0] == "url" {
			// header, skip
			continue
		}
		u := &URLReq{}
		err = u.UnmarshalRecord(record)
		if err != nil {
			mlog.Error(err)
			break
		}
		_, err = UpsertURL(u, false, false, time.Now())
		if err != nil {
			mlog.Error(err)
			break
		}
		rows++
	}
	mlog.Info("Import complete with %d rows in %s", rows, time.Since(start))
	return
}
