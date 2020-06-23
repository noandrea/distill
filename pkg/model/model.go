package model

import (
	"fmt"
	"net/http"
	"time"

	"github.com/noandrea/distill/pkg/common"
)

// Opcodes for stats
const (
	OpcodeInsert  = 0
	OpcodeGet     = 1
	OpcodeDelete  = 2
	OpcodeExpired = 3
	OpcodeStore   = 4
)

// opcodes for key prefix
const (
	KeySysPrefix  = 0x00
	KeyStatPrefix = 0x02
	KeyURLPrefix  = 0x04
)

// others
var (
	NumberZero = common.Itoa(0)
)

// NewURLInfo create a new URLInfo
func NewURLInfo(ID, redirectURL string) *URLInfo {
	return &URLInfo{
		ID:          ID,
		RedirectURL: redirectURL,
		RecordedOn:  time.Now(),
	}
}

// URLInfoFromURLReq copy fields from URLReq to URLInfo
func URLInfoFromURLReq(r URLReq) (u *URLInfo) {
	return &URLInfo{
		ID:                   r.ID,
		RedirectURL:          r.RedirectURL,
		RecordedOn:           r.ReceivedOn,
		ExhaustedRedirectURL: r.ExhaustedRedirectURL,
		ExpiredRedirectURL:   r.ExpiredRedirectURL,
		InactiveRedirectURL:  r.InactiveRedirectURL,
		TTL:                  r.TTL,
		ResolveLimit:         r.ResolveLimit,
		ExpiresOn:            r.ExpiresOn,
		ActiveFrom:           r.ActiveFrom,
	}
}

// URLInfo main entity to store
type URLInfo struct {
	ID                   string    `json:"id,omitempty"`
	RedirectURL          string    `json:"redirectURL,omitempty"`
	Hits                 int64     `json:"hits,omitempty"`
	RecordedOn           time.Time `json:"recordedOn,omitempty"`
	ActiveFrom           time.Time `json:"activeFrom,omitempty"`
	ExpiresOn            time.Time `json:"expiresOn,omitempty"`
	ResolveCount         int64     `json:"resolveCount,omitempty"`
	ResolveLimit         int64     `json:"resolveLimit,omitempty"`
	TTL                  int64     `json:"TTL,omitempty"`
	ExpiredRedirectURL   string    `json:"expiredRedirectURL,omitempty"`
	ExhaustedRedirectURL string    `json:"exhaustedRedirectURL,omitempty"`
	InactiveRedirectURL  string    `json:"inactiveRedirectURL,omitempty"`
}

// URLOp to track events on urls
type URLOp struct {
	Opcode int
	ID     string
	Err    error
}

// ShortID used in reply and channel comunication
type ShortID struct {
	ID string `json:"id"`
}

// URLReq request from a client to register an url
type URLReq struct {
	ID                   string    `json:"id"`
	ReceivedOn           time.Time `json:"received_on,omitempty"`
	RedirectURL          string    `json:"redirect_url"`
	ResolveLimit         int64     `json:"max_requests,omitempty"`
	TTL                  int64     `json:"ttl,omitempty"`
	ExpiresOn            time.Time `json:"expire_on,omitempty"`
	ActiveFrom           time.Time `json:"active_from,omitempty"`
	ExpiredRedirectURL   string    `json:"redirect_expired_url"`
	ExhaustedRedirectURL string    `json:"redirect_exhausted_url"`
	InactiveRedirectURL  string    `json:"redirect_inactive_url"`
}

// Statistics contains the global statistics
type Statistics struct {
	Urls        int64     `json:"urls"`
	Gets        int64     `json:"gets"`
	GetsExpired int64     `json:"gets_expired"`
	Upserts     int64     `json:"upserts"`
	Deletes     int64     `json:"deletes"`
	LastRequest time.Time `json:"last_request"`
}

// Bind will run after the unmarshalling is complete
func (u *URLReq) Bind(r *http.Request) error {
	return nil
}

// Bind will run after the unmarshalling is complete
func (u *ShortID) Bind(r *http.Request) error {
	return nil
}

// //     ______   ______  ____   ____
// //   .' ___  |.' ____ \|_  _| |_  _|
// //  / .'   \_|| (___ \_| \ \   / /
// //  | |        _.____`.   \ \ / /
// //  \ `.___.'\| \____) |   \ ' /
// //   `.____ .' \______.'    \_/
// //

// OpcodeToString opcode to sting
func OpcodeToString(opcode int) (label string) {
	switch opcode {
	case OpcodeInsert:
		label = "UPS"
	case OpcodeGet:
		label = "GET"
	case OpcodeDelete:
		label = "DEL"
	case OpcodeExpired:
		label = "EXP"
	}
	return
}

//   ________  _______     _______
//  |_   __  ||_   __ \   |_   __ \
//    | |_ \_|  | |__) |    | |__) |
//    |  _| _   |  __ /     |  __ /
//   _| |__/ | _| |  \ \_  _| |  \ \_
//  |________||____| |___||____| |___|
//

// ErrURLExpired when url is expired
var ErrURLExpired = fmt.Errorf("URL expired")

// ErrURLInactive when url is inactive
var ErrURLInactive = fmt.Errorf("URL expired")

// ErrURLExhausted when url is expired
var ErrURLExhausted = fmt.Errorf("URL exhausted")

// ErrInvalidBackupRecord when a csv record from backup is different from expected
var ErrInvalidBackupRecord = fmt.Errorf("Invalid backup record")
