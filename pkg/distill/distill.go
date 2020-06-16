package distill

import (
	"encoding/hex"
	"fmt"
	"net/url"

	"strings"
	"time"

	"github.com/noandrea/distill/config"
	"github.com/noandrea/distill/pkg/common"
	"github.com/noandrea/distill/pkg/datastore"
	"github.com/noandrea/distill/pkg/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
)

const (
	configID = "//cfg//"
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

func genKey(tenant, id string) string {
	th := blake2b.Sum256([]byte(tenant))
	// k := append([]byte, th[0:16]..., []byte(id))
	return fmt.Sprint(hex.EncodeToString(th[0:16]), ":", id)

}

func validateURL(u *string, allowEmpty bool) (err error) {
	if len(strings.TrimSpace(*u)) == 0 && allowEmpty {
		return
	}
	urlS, err := url.Parse(*u)
	if err != nil {
		return
	}
	if len(urlS.Host) == 0 {
		err = fmt.Errorf("Empty hostname in URL")
		return
	}
	*u = urlS.String()
	return
}

// calculateExpiration calculate the expiration of a url
// returns the highest date between the date binding + ttl and the date expiration date
func calculateExpiration(activeFrom time.Time, ttl int64, expiresOn time.Time) (expirationDate time.Time, err error) {
	if ttl == 0 && expiresOn.IsZero() {
		expirationDate = expiresOn
		return
	}
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

func buildURLInfo(url *model.URLReq, cfg config.ShortIDConfig) (u *model.URLInfo, err error) {
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
	u.ID = strings.TrimSpace(u.ID)
	if len(u.ID) == 0 {
		u.ID = generateID(cfg.Alphabet, cfg.Length)
	}

	// TODO: normalize urls

	// work with expiration times
	// compute the active from date
	if u.ActiveFrom.Before(u.RecordedOn) {
		u.ActiveFrom = u.RecordedOn
	}
	// compute the expiration date
	u.ExpiresOn, err = calculateExpiration(u.ActiveFrom, u.TTL, u.ExpiresOn)
	if err != nil {
		log.Error("Error calculating expiration:", err)
		return
	}

	// set max requests, the local version always has priority
	if u.ResolveLimit == 0 {
		u.ResolveLimit = settings.ShortID.MaxRequests
	}
	return
}

// UpsertURLSimple insert or update an url
// shortcut for UpsertURL(true, true, time.Now())
func UpsertURLSimple(tenant string, url *model.URLReq) (id string, err error) {
	return UpsertURL(tenant, url)
}

// UpsertURL insert or udpdate a url mapping
func UpsertURL(tenant string, url *model.URLReq) (id string, err error) {
	log.Debug("UpsertURL for tenant: ", tenant)
	// retrieve the configuration
	cfgKey := genKey(tenant, configID)
	cfg := settings.ShortID
	if found, _ := ds.Get(cfgKey, &cfg); found {
		log.Debug("Loading local settings for ", tenant)
	}
	// TODO: add hard cap counter for number of requests allowed
	// build the URLInfo object
	u, err := buildURLInfo(url, cfg)
	// compute key
	k := genKey(tenant, u.ID)
	// store the result
	if err = ds.Upsert(k, u); err != nil {
		log.Error("Error inserting new id: ", err)
		return
	}
	id = u.ID
	//TODO: collect statistics
	return
}

// DeleteURL delete a url mapping
func DeleteURL(tenant, id string) (err error) {
	k := genKey(tenant, id)
	err = ds.Delete(k)
	if err != nil {
		return
	}
	//TODO: collect statistics
	return
}

// GetURLRedirect retrieve the redirect url associated to an id
// it also fire an event of type opcodeGet
func GetURLRedirect(tenant string, id string) (redirectURL string, err error) {
	// retrieve the configuration
	cfgKey := genKey(tenant, configID)
	cfg := settings.ShortID
	if found, _ := ds.Get(cfgKey, &cfg); found {
		log.Debug("Loading local settings for ", tenant)
	}
	// urlInfo
	k := genKey(tenant, id)
	u, err := ds.Hit(k)
	if err != nil {
		return
	}

	// test for inactive
	if u.ActiveFrom.Before(time.Now()) {
		log.Debugf("Inactive date for %v, limit %v, requests %v", u.ID, u.Hits, u.ResolveLimit)
		err = model.ErrURLExpired
		redirectURL = common.IfEmptyThen(u.InactiveRedirectURL, cfg.InactiveRedirectURL)
		//TODO: collect statistics
		return
	}
	// test for expiration
	idExpiration := u.ExpiresOn
	if !idExpiration.IsZero() && time.Now().After(idExpiration) {
		log.Debugf("Expire date for %v, limit %v, requests %v", u.ID, u.Hits, u.ResolveLimit)
		err = model.ErrURLExpired
		redirectURL = common.IfEmptyThen(u.ExpiredRedirectURL, cfg.ExpiredRedirectURL)
		//TODO: collect statistics
		return
	}
	// test for limits
	if u.ResolveLimit > 0 && u.ResolveCount > u.ResolveLimit {
		log.Tracef("Expire max request for %v, limit %v, requests %v",
			u.ID,
			u.Hits,
			u.ResolveLimit)
		err = model.ErrURLExhausted
		redirectURL = common.IfEmptyThen(u.ExhaustedRedirectURL, cfg.ExhaustedRedirectURL)
		// push event
		//TODO: collect statistics
		return
	}

	//TODO: collect statistics

	// return the redirectUrl
	redirectURL = u.RedirectURL
	return
}

// GetURLInfo retrieve the url info associated to an id
func GetURLInfo(tenant, id string) (u model.URLInfo, err error) {
	k := genKey(tenant, id)
	u, err = ds.Peek(k)
	return
}
