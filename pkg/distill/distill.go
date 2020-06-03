package distill

import (
	"fmt"
	net "net/url"
	"regexp"
	"strings"
	"time"

	"github.com/noandrea/distill/config"
	"github.com/noandrea/distill/pkg/common"
	"github.com/noandrea/distill/pkg/datastore"
	"github.com/noandrea/distill/pkg/model"
	log "github.com/sirupsen/logrus"
)

var (
	settings config.Schema
	ds       datastore.URLDatastore
)

// NewSession starts a new distill session
func NewSession(cfg config.Schema, datastore datastore.URLDatastore) {
	settings = cfg
	ds = datastore
}

// generateID generates a new id
// it is guaranteed that returns an id of at least 1 character
func generateID(alphabet string, length int) (shortID string) {
	// a and l are validated before
	shortID, err := common.RandomString(alphabet, length)
	if err != nil {
		log.Warn("generateID: error generating IDs", err)
	}
	return
}

// UpsertURLSimple insert or update an url
// shortcut for UpsertURL(true, true, time.Now())
func UpsertURLSimple(url *model.URLReq) (id string, err error) {
	return UpsertURL(url, true, true, time.Now())
}

// UpsertURL insert or udpdate a url mapping
func UpsertURL(url *model.URLReq, forceAlphabet, forceLength bool, boundAt time.Time) (id string, err error) {
	// preprocess the url and generates the id if necessary
	// chech that the target url is a valid url
	if _, err = net.Parse(url.URL); err != nil {
		log.Info(url.URL)
		log.Error(err)
		return
	}
	// check that, if set ExhaustedURL is a valid url
	if len(url.ExhaustedURL) > 0 {
		if _, err = net.Parse(url.ExhaustedURL); err != nil {
			log.Error(err)
			return
		}
	}
	// check that, if set ExhaustedURL is a valid url
	if len(url.ExpiredURL) > 0 {
		if _, err = net.Parse(url.ExpiredURL); err != nil {
			log.Error(err)
			return
		}
	}

	// check if the ID is set
	id = strings.TrimSpace(url.ID)
	if len(id) == 0 {
		id = generateID(settings.ShortID.Alphabet, settings.ShortID.Length)
	} else {
		// TODO: check longest allowed key in badger
		p := fmt.Sprintf("[^%s]", regexp.QuoteMeta(settings.ShortID.Alphabet))
		m, _ := regexp.MatchString(p, url.ID)
		if forceAlphabet && m {
			err = fmt.Errorf("ID %v doesn't match alphabet and forceAlphabet is active", url.ID)
			return "", err
		}
		if forceLength && len(url.ID) != settings.ShortID.Length {
			err = fmt.Errorf("ID %v doesn't match length and forceLength len %v, required %v", url.ID, len(url.ID), settings.ShortID.Length)
			return "", err
		}
	}

	// set the binding date
	u := model.NewURLInfo(id, url.URL)

	// set exhausted url
	u.ExhaustedRedirectURL = url.ExhaustedURL
	// set expired url
	u.ExpiredRedirectURL = url.ExpiredURL
	// TODO: set inactive url
	u.InactiveRedirectURL = url.ExhaustedURL

	// now set constraints
	u.TTL = url.TTL
	// the local expiration always take priority
	u.SetExpirationTime(calculateExpiration(u, url.TTL, url.ExpireOn))
	if u.GetExpirationTime().IsZero() {
		// global expiration
		u.SetExpirationTime(calculateExpiration(u, settings.ShortID.TTL, settings.ShortID.ExpireOn))
	}
	// set max requests, the local version always has priority
	u.ResolveLimit = url.MaxRequests
	if u.ResolveLimit == 0 {
		u.ResolveLimit = settings.ShortID.MaxRequests
	}
	// cleanup the string id
	u.Id = strings.TrimSpace(url.ID)
	// process url id

	if err = ds.Upsert(id, u); err == nil {
		// collect statistics
		datastore.PushEvent(&model.URLOp{
			Opcode: model.OpcodeInsert,
			ID:     u.Id,
			Err:    err,
		})
	} else {
		log.Error("Error inserting new id")
	}
	return
}

// calculateExpiration calculate the expiration of a url
// returns the highest date between the date binding + ttl and the date expiration date
func calculateExpiration(u *model.URLInfo, ttl int64, expireDate time.Time) (expire time.Time) {
	if ttl > 0 {
		expire = common.ProtoTime(u.ActiveFrom).Add(time.Duration(ttl) * time.Second)
	}
	if expireDate.After(expire) {
		expire = expireDate
	}
	return
}

// DeleteURL delete a url mapping
func DeleteURL(id string) (err error) {
	err = ds.Delete(id)
	if err != nil {
		return
	}
	// collect statistics
	datastore.PushEvent(&model.URLOp{
		Opcode: model.OpcodeDelete,
		ID:     id,
	})
	return
}

// GetURLRedirect retrieve the redicrect url associated to an id
// it also fire an event of tipe opcodeGet
func GetURLRedirect(id string) (redirectURL string, err error) {
	// urlInfo
	u, err := ds.Hit(id)
	if err != nil {
		return
	}

	urlop := &model.URLOp{ID: u.Id}

	idExpiration := common.ProtoTime(u.ExpiresOn)

	if !idExpiration.IsZero() && time.Now().After(idExpiration) {
		log.Debugf("Expire date for %v, limit %v, requests %v", u.Id, u.Hits, u.ResolveLimit)
		err = model.ErrURLExpired
		redirectURL = common.IfEmptyThen(u.ExpiredRedirectURL, settings.ShortID.ExpiredRedirectURL)

		urlop.Err, urlop.Opcode = err, model.OpcodeExpired
		datastore.PushEvent(urlop)
		return
	}
	if u.ResolveLimit > 0 && u.Hits > u.ResolveLimit {
		log.Tracef("Expire max request for %v, limit %v, requests %v",
			u.Id,
			u.Hits,
			u.ResolveLimit)
		err = model.ErrURLExhausted
		redirectURL = common.IfEmptyThen(u.ExhaustedRedirectURL, settings.ShortID.ExhaustedRedirectURL)
		// push event
		urlop.Err, urlop.Opcode = err, model.OpcodeExpired
		datastore.PushEvent(urlop)
		return
	}

	// collect statistics
	urlop.Err, urlop.Opcode = err, model.OpcodeGet
	datastore.PushEvent(urlop)
	// return the redirectUrl
	redirectURL = u.RedirectURL
	return
}

// GetURLInfo retrieve the url info associated to an id
func GetURLInfo(id string) (u model.URLInfo, err error) {
	u, err = ds.Peek(id)
	return
}

// ImportCSV import urls from a csv file
func ImportCSV(inFile string) (rows int, err error) {
	// fp, err := os.Open(inFile)
	// if err != nil {
	// 	return
	// }
	// defer fp.Close()
	// start := time.Now()
	// csvR := csv.NewReader(fp)
	// for {
	// 	record, err := csvR.Read()
	// 	if err == io.EOF {
	// 		log.Error(err)
	// 		break
	// 	}
	// 	if err != nil {
	// 		log.Error(err)
	// 		break
	// 	}
	// 	if rows == 0 && common.IsEqStr(record[0], "url") {
	// 		// header, skip
	// 		continue
	// 	}
	// 	u := &model.URLReq{}
	// 	err = u.UnmarshalRecord(record)
	// 	if err != nil {
	// 		log.Error(err)
	// 		break
	// 	}
	// 	_, err = UpsertURL(u, false, false, time.Now())
	// 	if err != nil {
	// 		log.Error(err)
	// 		break
	// 	}
	// 	rows++
	// }
	// log.Infof("Import complete with %d rows in %s", rows, time.Since(start))
	return
}
