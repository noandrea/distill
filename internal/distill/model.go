package distill

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"sync"
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
	keySysPrefix          = 0x00
	keyStatPrefix         = 0x02
	keyURLPrefix          = 0x04
	keyURLStatCountPrefix = 0x05
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
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	MaxRequests int64     `json:"max_requests,omitempty"`
	TTL         int64     `json:"ttl,omitempty"`
	ExpireOn    time.Time `json:"expire_on,omitempty"`
}

// Statistics contains the global statistics
type Statistics struct {
	mutex   sync.Mutex
	Urls    int64 `json:"urls"`
	Gets    int64 `json:"gets"`
	Upserts int64 `json:"upserts"`
	Deletes int64 `json:"deletes"`
}

func (s *Statistics) String() string {
	return fmt.Sprintf("urls: %d, gets: %d, upserts:%d, deletes:%d",
		s.Urls,
		s.Gets,
		s.Upserts,
		s.Deletes,
	)
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

//     ______   ______  ____   ____
//   .' ___  |.' ____ \|_  _| |_  _|
//  / .'   \_|| (___ \_| \ \   / /
//  | |        _.____`.   \ \ / /
//  \ `.___.'\| \____) |   \ ' /
//   `.____ .' \______.'    \_/
//

//MarshalRecord marshal a urlinfo to a string array (for csv)
func (u *URLInfo) MarshalRecord() []string {
	pieces := make([]string, 7)
	pieces[0] = u.ID
	pieces[1] = u.URL
	pieces[2] = fTime(u.BountAt)
	pieces[3] = fInt64(u.Counter)
	pieces[4] = fInt64(u.MaxRequests)
	pieces[5] = fInt64(u.TTL)
	pieces[6] = fTime(u.ExpireOn)
	return pieces
}

//UnmarshalRecord unmarshal a string array into a urlinfo
func (u *URLInfo) UnmarshalRecord(pieces []string) (err error) {
	pl := len(pieces)
	if pl != 7 {
		return fmt.Errorf("Invalid backup record! record corrupted")
	}
	u.URL = pieces[0]
	u.ID = pieces[1]
	if u.BountAt, err = pTime(pieces, 2, pl); err != nil {
		return
	}
	if u.Counter, err = pInt64(pieces, 3, pl); err != nil {
		return
	}
	if u.MaxRequests, err = pInt64(pieces, 4, pl); err != nil {
		return
	}
	if u.TTL, err = pInt64(pieces, 5, pl); err != nil {
		return
	}
	if u.BountAt, err = pTime(pieces, 6, pl); err != nil {
		return
	}
	return
}

//UnmarshalRecord unmarshal a string array (csv record) to URLReq pointer
func (u *URLReq) UnmarshalRecord(pieces []string) (err error) {
	u.URL = pieces[0]
	p := len(pieces)
	if p > 1 {
		u.ID = pieces[1]
	}
	if u.MaxRequests, err = pInt64(pieces, 2, p); err != nil {
		return
	}
	if u.TTL, err = pInt64(pieces, 3, p); err != nil {
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

// fInt64 for csv printing
func fInt64(v int64) (str string) {
	if v == 0 {
		return
	}
	str = strconv.FormatInt(v, 10)
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

// pTime parse a time.Time from a string
func pTime(src []string, idx, srcLen int) (v time.Time, err error) {
	if idx < srcLen && len(src[idx]) > 0 {
		v, err = time.Parse(time.RFC3339, src[idx])
	}
	return
}

func itoa(i int64) (b []byte) {
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return
}

func atoi(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
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

// ErrInvalidBackupRecord when a csv record from backup is different from expected
var ErrInvalidBackupRecord = fmt.Errorf("Invalid backup record")

//   ___  ____   ________  ____  ____   ______
//  |_  ||_  _| |_   __  ||_  _||_  _|.' ____ \
//    | |_/ /     | |_ \_|  \ \  / /  | (___ \_|
//    |  __'.     |  _| _    \ \/ /    _.____`.
//   _| |  \ \_  _| |__/ |   _|  |_   | \____) |
//  |____||____||________|  |______|   \______.'
//

func keyURL(id string) (k []byte) {
	return key(keyURLPrefix, id)
}

func keyURLStatCount(id string) (k []byte) {
	return key(keyURLStatCountPrefix, id)
}

func keySys(id string) (k []byte) {
	return key(keySysPrefix, id)
}

func keyGlobalStat(id string) (k []byte) {
	return key(keyStatPrefix, id)
}

func key(prefix byte, id string) (k []byte) {
	idb := []byte(id)
	k = make([]byte, len(idb)+1)
	k[0] = prefix
	copy(k[1:], idb)
	return
}
