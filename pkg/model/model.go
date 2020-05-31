package model

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	opcodeInsert  = 0
	opcodeGet     = 1
	opcodeDelete  = 2
	opcodeExpired = 3
	opcodeStore   = 4
)

const (
	keySysPrefix  = 0x00
	keyStatPrefix = 0x02
	keyURLPrefix  = 0x04
)

var (
	numberZero = itoa(0)
)

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
	opcode int
	ID     string
	err    error
}

// ShortID used in reply and channel comunication
type ShortID struct {
	ID string `json:"id"`
}

// URLReq request from a client to register an url
type URLReq struct {
	ID           string    `json:"id"`
	URL          string    `json:"url"`
	MaxRequests  uint64    `json:"max_requests,omitempty"`
	ExhaustedURL string    `json:"url_exhausted"`
	TTL          uint64    `json:"ttl,omitempty"`
	ExpireOn     time.Time `json:"expire_on,omitempty"`
	ExpiredURL   string    `json:"url_expired"`
}

// Statistics contains the global statistics
type Statistics struct {
	Urls        uint64    `json:"urls"`
	Gets        uint64    `json:"gets"`
	GetsExpired uint64    `json:"gets_expired"`
	Upserts     uint64    `json:"upserts"`
	Deletes     uint64    `json:"deletes"`
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

func (s *Statistics) record(get, upsert, delete, urls, getExpired int64) {
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
	return u.BountAt.Add(time.Duration(u.TTL) * time.Second)
}

// String version of urlinfo
func (u URLInfo) String() string {
	//return fmt.Sprint("%#v", u)
	return fmt.Sprintf("%v c:%d %v [mr:%d, exp:%v] --> %v", u.ID, u.Counter, u.BountAt.Format(time.Stamp), u.MaxRequests, u.ExpireOn.Format(time.RFC3339Nano), u.URL)
}

// Bind will run after the unmarshalling is complete
func (u *URLReq) Bind(r *http.Request) error {
	return nil
}

// Bind will run after the unmarshalling is complete
func (u *ShortID) Bind(r *http.Request) error {
	return nil
}

//     ______   ______  ____   ____
//   .' ___  |.' ____ \|_  _| |_  _|
//  / .'   \_|| (___ \_| \ \   / /
//  | |        _.____`.   \ \ / /
//  \ `.___.'\| \____) |   \ ' /
//   `.____ .' \______.'    \_/
//

//MarshalRecord marshal a urlinfo to a string array (for csv)
func (u *URLInfo) MarshalRecord() []string {
	pieces := make([]string, 9)
	pieces[0] = u.ID
	pieces[1] = u.URL
	pieces[2] = fTime(u.BountAt)
	pieces[3] = fUint64(u.Counter)
	pieces[4] = fUint64(u.MaxRequests)
	pieces[5] = u.ExhaustedURL
	pieces[6] = fUint64(u.TTL)
	pieces[7] = fTime(u.ExpireOn)
	pieces[8] = u.ExpiredURL
	return pieces
}

//UnmarshalRecord unmarshal a string array into a urlinfo
func (u *URLInfo) UnmarshalRecord(pieces []string) (err error) {
	pl := len(pieces)
	if pl != 9 {
		return fmt.Errorf("Invalid backup record! record corrupted")
	}
	u.ID = pieces[0]
	u.URL = pieces[1]
	if u.BountAt, err = pTime(pieces, 2, pl); err != nil {
		return
	}
	if u.Counter, err = pUint64(pieces, 3, pl); err != nil {
		return
	}
	if u.MaxRequests, err = pUint64(pieces, 4, pl); err != nil {
		return
	}
	u.ExhaustedURL = pieces[5]
	if u.TTL, err = pUint64(pieces, 6, pl); err != nil {
		return
	}
	if u.BountAt, err = pTime(pieces, 7, pl); err != nil {
		return
	}
	u.ExpiredURL = pieces[8]
	return
}

//UnmarshalRecord unmarshal a string array (csv record) to URLReq pointer
func (u *URLReq) UnmarshalRecord(pieces []string) (err error) {
	u.URL = pieces[0]
	p := len(pieces)
	if p > 1 {
		u.ID = pieces[1]
	}
	if u.MaxRequests, err = pUint64(pieces, 2, p); err != nil {
		return
	}
	if u.TTL, err = pUint64(pieces, 3, p); err != nil {
		return
	}
	if u.ExpireOn, err = pTime(pieces, 4, p); err != nil {
		return
	}
	return
}

//   ____  ____  ________  _____ 4   _______  ________  _______     ______
//  |_   ||   _||_   __  ||_   _|   |_   __ \|_   __  ||_   __ \  .' ____ \
//    | |__| |    | |_ \_|  | |       | |__) | | |_ \_|  | |__) | | (___ \_|
//    |  __  |    |  _| _   | |   _   |  ___/  |  _| _   |  __ /   _.____`.
//   _| |  | |_  _| |__/ | _| |__/ | _| |_    _| |__/ | _| |  \ \_| \____) |
//  |____||____||________||________||_____|  |________||____| |___|\______.'
//

// fUint64 for csv printing
func fUint64(v uint64) (str string) {
	if v == 0 {
		return
	}
	str = strconv.FormatUint(v, 10)
	return
}

// fTime for csv printing
func fTime(v time.Time) (str string) {
	if v.IsZero() {
		return
	}
	str = v.Format(time.RFC3339)
	return
}

// pInt64 parse an int64 from a string
func pInt64(src []string, idx, srcLen int) (v int64, err error) {
	if idx < srcLen && len(src[idx]) > 0 {
		v, err = strconv.ParseInt(src[idx], 10, 64)
	}
	return
}

// pUint64 parse an int64 from a string
func pUint64(src []string, idx, srcLen int) (v uint64, err error) {
	if idx < srcLen && len(src[idx]) > 0 {
		v, err = strconv.ParseUint(src[idx], 10, 64)
	}
	return
}

// pTime parse a time.Time from a string
func pTime(src []string, idx, srcLen int) (v time.Time, err error) {
	if idx < srcLen && len(src[idx]) > 0 {
		v, err = time.Parse(time.RFC3339, src[idx])
	}
	return
}

func itoa(i uint64) (b []byte) {
	b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return
}

func atoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func opcodeToString(opcode int) (label string) {
	switch opcode {
	case opcodeInsert:
		label = "UPS"
	case opcodeGet:
		label = "GET"
	case opcodeDelete:
		label = "DEL"
	case opcodeExpired:
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


