package internal

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/jbrodriguez/mlog"
)

// ServerConfig configuration for the server
type ServerConfig struct {
	APIKey       string `yaml:"apiKey"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	RootRedirect string `yaml:"rootRedirect"`
	DbPath       string `yaml:"dbPath"`
	EnableStats  bool   `yaml:"enableStats"`
}

//ShortIDConfig configureaiont for the short id
type ShortIDConfig struct {
	Alphabet    string `yaml:"alphabet"`
	Length      int    `yaml:"length"`
	MaxRequests int64  `yaml:"maxRequests"`
	TTL         int64  `yaml:"ttl"`
}

// TuningConfig fine tuning configuration
type TuningConfig struct {
	StatsWorkerPoolSize  int `yaml:"statsWorkerPoolSize"`
	StatsWorkerQueueSize int `yaml:"statsWorkerQueueSize"`
	StatsEventsWorkerNum int `yaml:"statsEventsWorkerNum"`
	StatsEventsQueueSize int `yaml:"statsEventsQueueSize"`
}

// ConfigSchema define the configuration object
type ConfigSchema struct {
	Server  ServerConfig  `yaml:"server"`
	ShortID ShortIDConfig `yaml:"shortId"`
	Tuning  TuningConfig  `yaml:"tuning"`
}

//Validate configuration
func (c *ConfigSchema) Validate() {

	mlog.Trace(spew.Sprint(c))

	if c.ShortID.Length < 3 {
		panic("short_id.length must be at least 3")
	}

	if len(c.ShortID.Alphabet) < c.ShortID.Length {
		panic(fmt.Sprint("short_id.alphabet must be at least ", c.ShortID.Length, " characters long"))
	}

	if c.Tuning.StatsWorkerPoolSize == 0 {
		c.Tuning.StatsWorkerPoolSize = 1
	}

	if c.Tuning.StatsWorkerQueueSize == 0 {
		c.Tuning.StatsWorkerQueueSize = 2
	}

	if c.Tuning.StatsEventsQueueSize == 0 {
		c.Tuning.StatsEventsQueueSize = 2048
	}

	if c.Tuning.StatsEventsWorkerNum == 0 {
		c.Tuning.StatsEventsWorkerNum = 1
	}
}

// Config sytem configuration
var Config ConfigSchema
