package iljl

const (
	OpcodeInsert = 0
	OpcodeGet    = 1
	OpcodeDelete = 2
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

// Statistics
type Statistics struct {
	Urls    int64 `json:"urls"`
	Gets    int64 `json:"gets"`
	Upserts int64 `json:"upserts"`
	Deletes int64 `json:"deletes"`
}
