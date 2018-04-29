package iljl

import (
	"encoding/binary"
	"fmt"
	"time"
)

const (
	opcodeInsert = 0
	opcodeGet    = 1
	opcodeDelete = 2
)

const (
	KeySysPrefix          = 0x00
	KeyStatPrefix         = 0x02
	KeyURLPrefix          = 0x04
	KeyURLStatCountPrefix = 0x05
)

var (
	numberZero = itoa(0)
)

// URLOp to track events on urls
type URLOp struct {
	opcode int
	url    URLInfo
	err    error
}

// ShortID used in reply and channel comunication
type ShortID struct {
	ID string `json:"id"`
}

// URLReq request from a client to register an url
type URLReq struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	TTL         int64  `json:"ttl"`
	MaxRequests int64  `json:"max_requests"`
}

// Statistics contains the global statistics
type Statistics struct {
	Urls    int64 `json:"urls"`
	Gets    int64 `json:"gets"`
	Upserts int64 `json:"upserts"`
	Deletes int64 `json:"deletes"`
}

func (s Statistics) String() string {
	return fmt.Sprintf("urls: %d, gets: %d, upserts:%d, deletes:%d",
		s.Urls,
		s.Gets,
		s.Upserts,
		s.Deletes,
	)
}

func ttl(seconds int64) (d time.Duration) {
	return time.Duration(seconds) * time.Second
}

func itoa(i int64) (b []byte) {
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return
}

func atoi(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}

//   ___  ____   ________  ____  ____   ______
//  |_  ||_  _| |_   __  ||_  _||_  _|.' ____ \
//    | |_/ /     | |_ \_|  \ \  / /  | (___ \_|
//    |  __'.     |  _| _    \ \/ /    _.____`.
//   _| |  \ \_  _| |__/ |   _|  |_   | \____) |
//  |____||____||________|  |______|   \______.'
//

func keyURL(id string) (k []byte) {
	return key(KeyURLPrefix, id)
}

func keyURLStatCount(id string) (k []byte) {
	return key(KeyURLStatCountPrefix, id)
}

func keySys(id string) (k []byte) {
	return key(KeySysPrefix, id)
}

func keyGlobalStat(id string) (k []byte) {
	return key(KeyStatPrefix, id)
}

func key(prefix byte, id string) (k []byte) {
	idb := []byte(id)
	k = make([]byte, len(idb)+1)
	k[0] = prefix
	copy(k[1:], idb)
	return
}
