package distill

import (
	"fmt"

	"strings"
	"time"

	"github.com/PuerkitoBio/purell"
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

func validateURL(url *string, allowEmpty bool) (err error) {
	if len(strings.TrimSpace(*url)) == 0 && allowEmpty {
		return
	}
	// normalize url
	nu, err := purell.NormalizeURLString(*url,
		purell.FlagLowercaseScheme|
			purell.FlagLowercaseHost|purell.FlagUppercaseEscapes)
	if err != nil {
		log.Error(err)
		return
	}
	url = &nu
	return
}

// calculateExpiration calculate the expiration of a url
// returns the highest date between the date binding + ttl and the date expiration date
func calculateExpiration(activeFrom time.Time, ttl int64, expiresOn time.Time) (expirationDate time.Time, err error) {
	// add the ttl
	expirationDate = activeFrom.Add(time.Duration(ttl) * time.Second)
	// check if the dates make sense
	if !expiresOn.IsZero() && activeFrom.After(expiresOn) {
		err = fmt.Errorf("URL expired on submission, activeFrom = %s, expiresOn = %s", activeFrom, expiresOn)
		return
	}
	// get the latest date
	if expiresOn.After(expirationDate) {
		expirationDate = expiresOn
	}
	return
}

func buildURLInfo(url *model.URLReq, shortIDAlphabet string, shortIDLength int) (u *model.URLInfo, err error) {
	u = model.URLInfoFromURLReq(*url)
	// preprocess the url and generates the id if necessary
	// check that the target url is a valid url
	if err = validateURL(&u.RedirectURL, false); err != nil {
		log.Error("RedirectURL  is not a valid url: ", err)
		return
	}
	// check the exhausted url
	if err = validateURL(&u.ExhaustedRedirectURL, true); err != nil {
		log.Error("ExhaustedRedirectURL  is not a valid url: ", err)
		return
	}
	// check if expired redirect url is set and valid
	if err = validateURL(&u.ExpiredRedirectURL, true); err != nil {
		log.Error("ExpiredRedirectURL  is not a valid url: ", err)
		return
	}
	// check if inactive redirect url is set and valid
	if err = validateURL(&url.InactiveRedirectURL, true); err != nil {
		log.Error("InactiveRedirectURL  is not a valid url: ", err)
		return
	}

	// check if the ID is set
	u.Id = strings.TrimSpace(u.Id)
	if len(u.Id) == 0 {
		u.Id = generateID(shortIDAlphabet, shortIDLength)
	}
	// } else {
	// 	// TODO: check longest allowed key in badger
	// 	p := fmt.Sprintf("[^%s]", regexp.QuoteMeta(settings.ShortID.Alphabet))
	// 	m, _ := regexp.MatchString(p, url.ID)
	// 	if forceAlphabet && m {
	// 		err = fmt.Errorf("ID %v doesn't match alphabet and forceAlphabet is active", url.ID)
	// 		return "", err
	// 	}
	// 	if forceLength && len(url.ID) != settings.ShortID.Length {
	// 		err = fmt.Errorf("ID %v doesn't match length and forceLength len %v, required %v", url.ID, len(url.ID), settings.ShortID.Length)
	// 		return "", err
	// 	}
	// }

	// TODO: normalize urls

	// work with expiration times
	// compute the active from date
	activeFrom := common.ProtoTime(u.ActiveFrom)
	recordedOn := common.ProtoTime(u.RecordedOn)
	if activeFrom.Before(recordedOn) {
		activeFrom := recordedOn
		u.ActiveFrom = common.TimeProto(activeFrom)
	}
	// compute the expiration date
	ttl := u.TTL
	expiresOn := common.ProtoTime(u.ExpiresOn)
	expiresOn, err = calculateExpiration(activeFrom, ttl, expiresOn)
	if err != nil {
		log.Error("Error calculating expiration:", err)
		return
	}
	u.ExpiresOn = common.TimeProto(expiresOn)

	// set max requests, the local version always has priority
	if u.ResolveLimit == 0 {
		u.ResolveLimit = settings.ShortID.MaxRequests
	}
	return
}

// UpsertURLSimple insert or update an url
// shortcut for UpsertURL(true, true, time.Now())
func UpsertURLSimple(url *model.URLReq) (id string, err error) {
	return UpsertURL(url, true, true)
}

// UpsertURL insert or udpdate a url mapping
func UpsertURL(url *model.URLReq, forceAlphabet, forceLength bool) (id string, err error) {
	u, err := buildURLInfo(url, settings.ShortID.Alphabet, settings.ShortID.Length)
	if err = ds.Upsert(id, u); err != nil {
		log.Error("Error inserting new id: ", err)
		return
	}
	id = u.Id
	// collect statistics
	datastore.PushEvent(&model.URLOp{
		Opcode: model.OpcodeInsert,
		ID:     id,
		Err:    err,
	})
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

// GetURLRedirect retrieve the redirect url associated to an id
// it also fire an event of type opcodeGet
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
