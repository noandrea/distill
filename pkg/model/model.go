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
		RecordedOn:           time.Now(),
		ExhaustedRedirectURL: r.ExhaustedRedirectURL,
		ExpiredRedirectURL:   r.ExpiredRedirectURL,
		InactiveRedirectURL:  r.InactiveRedirectURL,
		TTL:                  r.TTL,
		ResolveLimit:         r.ResolveLimit,
		ExpiresOn:            r.ExpiresOn,
		ActiveFrom:           r.ActiveFrom,
	}
}

type URLInfo struct {
	ID                   string    `json:"id,omitempty"`
	RedirectURL          string    `json:"redirectURL,omitempty"`
	Hits                 int64     `json:"hits,omitempty"`
	RecordedOn           time.Time `json:"recordedOn,omitempty"`
	ActiveFrom           time.Time `json:"activeFrom,omitempty"`
	ExpiresOn            time.Time `json:"expiresOn,omitempty"`
	ResolveLimit         int64     `json:"resolveLimit,omitempty"`
	TTL                  int64     `json:"TTL,omitempty"`
	ExpiredRedirectURL   string    `json:"expiredRedirectURL,omitempty"`
	ExhaustedRedirectURL string    `json:"exhaustedRedirectURL,omitempty"`
	InactiveRedirectURL  string    `json:"inactiveRedirectURL,omitempty"`
}

// BinSerializable interface for binary serializable structs
type BinSerializable interface {
	MarshalBinary() (data []byte, err error)
	UnmarshalBinary(data []byte) error
}

// CSVSerializable interface for binary serializable structs
type CSVSerializable interface {
	MarshalRecord() (data []string, err error)
	UnmarshalRecord(data []string) error
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

func (s *Statistics) String() string {
	return fmt.Sprintf("URLs: %d, GETs: %d, Inserts: %d, Deletes: %d, GetsExpired: %d",
		s.Urls,
		s.Gets,
		s.Upserts,
		s.Deletes,
		s.GetsExpired,
	)
}

// Record statistics
func (s *Statistics) Record(get, upsert, delete, urls, getExpired int64) {
	// statsMutex.Lock()
	// // this is confusing but actually correct
	// // if the input number is negative will work just the same
	// s.Gets += uint64(get)
	// s.GetsExpired += uint64(getExpired)
	// s.Upserts += uint64(upsert)
	// s.Deletes += uint64(delete)
	// s.Urls += uint64(urls)
	// statsMutex.Unlock()
}

// ExpirationDate return the expiration date of the URLInfo
func (u URLInfo) ExpirationDate() time.Time {
	return u.ActiveFrom.Add(time.Duration(u.TTL) * time.Second)
}

// // String version of urlinfo
// func (u URLInfo) String() string {
// 	//return fmt.Sprint("%#v", u)
// 	return fmt.Sprintf("%v c:%d %v [mr:%d, exp:%v] --> %v",
// 		u.Id, u.Hits,
// 		common.ProtoTime(u.ActiveFrom).Format(time.Stamp),
// 		u.ResolveLimit,
// 		common.ProtoTime(u.ActiveFrom).Format(time.RFC3339Nano),
// 		u.RedirectURL)
// }

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

// //MarshalRecord marshal a urlinfo to a string array (for csv)
// func (u *URLInfo) MarshalRecord() []string {
// 	pieces := make([]string, 9)
// 	pieces[0] = u.Id
// 	pieces[1] = u.RedirectURL
// 	pieces[2] = fTime(common.ProtoTime(u.RecordedOn))
// 	pieces[3] = fUint64(u.Hits)
// 	pieces[4] = fUint64(u.ResolveLimit)
// 	pieces[5] = u.ExhaustedRedirectURL
// 	pieces[6] = fUint64(u.TTL)
// 	pieces[7] = fTime(common.ProtoTime(u.ExpiresOn))
// 	pieces[8] = u.ExpiredRedirectURL
// 	return pieces
// }

// //UnmarshalRecord unmarshal a string array into a urlinfo
// func (u *URLInfo) UnmarshalRecord(pieces []string) (err error) {
// 	pl := len(pieces)
// 	if pl != 9 {
// 		return fmt.Errorf("Invalid backup record! record corrupted")
// 	}
// 	u.Id = pieces[0]
// 	u.RedirectURL = pieces[1]
// 	if u.BountAt, err = protoTime(pieces, 2, pl); err != nil {
// 		return
// 	}
// 	if u.Counter, err = pUint64(pieces, 3, pl); err != nil {
// 		return
// 	}
// 	if u.MaxRequests, err = pUint64(pieces, 4, pl); err != nil {
// 		return
// 	}
// 	u.ExhaustedURL = pieces[5]
// 	if u.TTL, err = pUint64(pieces, 6, pl); err != nil {
// 		return
// 	}
// 	if u.BountAt, err = protoTime(pieces, 7, pl); err != nil {
// 		return
// 	}
// 	u.ExpiredURL = pieces[8]
// 	return
// }

// //UnmarshalRecord unmarshal a string array (csv record) to URLReq pointer
// func (u *URLReq) UnmarshalRecord(pieces []string) (err error) {
// 	u.URL = pieces[0]
// 	p := len(pieces)
// 	if p > 1 {
// 		u.ID = pieces[1]
// 	}
// 	if u.MaxRequests, err = pUint64(pieces, 2, p); err != nil {
// 		return
// 	}
// 	if u.TTL, err = pUint64(pieces, 3, p); err != nil {
// 		return
// 	}
// 	if u.ExpireOn, err = protoTime(pieces, 4, p); err != nil {
// 		return
// 	}
// 	return
// }

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
var ErrURLExpired = fmt.Errorf("url expired")

// ErrURLExhausted when url is expired
var ErrURLExhausted = fmt.Errorf("url exhausted")

// ErrInvalidBackupRecord when a csv record from backup is different from expected
var ErrInvalidBackupRecord = fmt.Errorf("Invalid backup record")
