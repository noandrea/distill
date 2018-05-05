package internal

import (
	"fmt"
	"time"
)

// ServerConfig configuration for the server
type ServerConfig struct {
	APIKey          string `yaml:"apiKey"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	DbPath          string `yaml:"dbPath"`
	RootRedirect    string `yaml:"rootRedirect"`
	ExpiredRedirect string `yaml:"expiredRedirect"`
}

//ShortIDConfig configureaiont for the short id
type ShortIDConfig struct {
	Alphabet    string    `yaml:"alphabet"`
	Length      int       `yaml:"length"`
	MaxRequests int64     `yaml:"maxRequests"`
	TTL         int64     `yaml:"ttl"`
	ExpireOn    time.Time `yaml:"expireOn"`
}

// TuningConfig fine tuning configuration
type TuningConfig struct {
	StatsEventsWorkerNum int     `yaml:"statsEventsWorkerNum"`
	StatsEventsQueueSize int     `yaml:"statsEventsQueueSize"`
	StatsCaheSize        int     `yaml:"statsCacheSize"`
	DbPurgeWritesCount   int64   `yaml:"dbPurgeWritesCount"`
	DbGCDeletesCount     int64   `yaml:"dbGCDeletesCount"`
	DbGCDiscardRation    float64 `yaml:"dbGCDiscardRation"`
	URLCaheSize          int     `yaml:"URLCaheSize"`
}

// ConfigSchema define the configuration object
type ConfigSchema struct {
	Server  ServerConfig  `yaml:"server"`
	ShortID ShortIDConfig `yaml:"shortId"`
	Tuning  TuningConfig  `yaml:"tuning"`
}

//Validate configuration
func (c *ConfigSchema) Validate() {

	if c.ShortID.Length < 3 {
		panic("short_id.length must be at least 3")
	}

	if len(c.ShortID.Alphabet) < c.ShortID.Length {
		panic(fmt.Sprint("short_id.alphabet must be at least ", c.ShortID.Length, " characters long"))
	}

	if c.Tuning.StatsEventsQueueSize <= 0 {
		c.Tuning.StatsEventsQueueSize = 2048
	}

	if c.Tuning.StatsEventsWorkerNum <= 0 {
		c.Tuning.StatsEventsWorkerNum = 1
	}

	if c.Tuning.StatsCaheSize <= 0 {
		c.Tuning.StatsCaheSize = 1024
	}

	if c.Tuning.DbPurgeWritesCount <= 0 {
		c.Tuning.DbPurgeWritesCount = 2000
	}

	if c.Tuning.DbGCDeletesCount <= 0 {
		c.Tuning.DbGCDeletesCount = 500
	}

	if c.Tuning.DbGCDiscardRation <= 0 || c.Tuning.DbGCDiscardRation > 1 {
		c.Tuning.DbGCDiscardRation = 0.5
	}

	if c.Tuning.URLCaheSize <= 0 || c.Tuning.URLCaheSize > 1 {
		c.Tuning.URLCaheSize = 2048
	}

}

// Config sytem configuration
var Config ConfigSchema
