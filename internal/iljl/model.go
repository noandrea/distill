package iljl

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
