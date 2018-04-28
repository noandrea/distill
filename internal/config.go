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
}

//ShortIDConfig configureaiont for the short id
type ShortIDConfig struct {
	Alphabet    string `yaml:"alphabet"`
	Length      int    `yaml:"length"`
	MaxRequests int64  `yaml:"maxRequests"`
	TTL         int64  `yaml:"ttl"`
	Domain      string `yaml:"domain"`
}

// ConfigSchema define the configuration object
type ConfigSchema struct {
	Server  ServerConfig  `yaml:"server"`
	ShortID ShortIDConfig `yaml:"shortId"`
}

//Validate configuration
func (c ConfigSchema) Validate() {

	mlog.Trace(spew.Sprint(c))

	if c.ShortID.Length < 3 {
		panic("short_id.length must be at least 3")
	}

	if len(c.ShortID.Alphabet) < c.ShortID.Length {
		panic(fmt.Sprint("short_id.alphabet must be at least ", c.ShortID.Length, " characters long"))
	}
}

// Config sytem configuration
var Config ConfigSchema
