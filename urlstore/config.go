package urlstore

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/noandrea/distill/pkg/common"
	yaml "gopkg.in/yaml.v2"
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
	StatsEventsWorkerNum   int     `yaml:"statsEventsWorkerNum"`
	StatsCaheSize          int     `yaml:"statsCacheSize"`
	DbPurgeWritesCount     int     `yaml:"dbPurgeWritesCount"`
	DbGCDeletesCount       int     `yaml:"dbGCDeletesCount"`
	DbGCDiscardRation      float64 `yaml:"dbGCDiscardRation"`
	URLCaheSize            int     `yaml:"URLCaheSize"`
	BckCSVIterPrefetchSize int     `yaml:"exportIteratorPrefetchSize"`
	APIKeyHeaderName       string  `yaml:"apiKeyHeaderName"`
}

// ConfigSchema define the configuration object
type ConfigSchema struct {
	Server  ServerConfig  `yaml:"server"`
	ShortID ShortIDConfig `yaml:"shortId"`
	Tuning  TuningConfig  `yaml:"tuning"`
}

func empty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

//Defaults generate configuration defaults
func (c *ConfigSchema) Defaults() {
	// for server
	common.DefaultIfEmptyStr(&c.Server.Host, "0.0.0.0")
	common.DefaultIfEmptyInt(&c.Server.Port, 1804)
	common.DefaultIfEmptyStr(&c.Server.DbPath, "distill.db")
	common.DefaultIfEmptyStr(&c.Server.RootRedirect, "https://github.com/noandrea/distill/wikis/welcome")
	common.DefaultIfEmptyStr(&c.Server.ExpiredRedirect, "https://github.com/noandrea/distill/wikis/Expired-URL")

	// for short id
	common.DefaultIfEmptyStr(&c.ShortID.Alphabet, "abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789")
	common.DefaultIfEmptyInt(&c.ShortID.Length, 6)

	// For tuning
	common.DefaultIfEmptyInt(&c.Tuning.StatsEventsWorkerNum, 1)
	common.DefaultIfEmptyInt(&c.Tuning.StatsCaheSize, 1024)
	common.DefaultIfEmptyInt(&c.Tuning.DbPurgeWritesCount, 2000)
	common.DefaultIfEmptyInt(&c.Tuning.DbGCDeletesCount, 500)
	if c.Tuning.DbGCDiscardRation <= 0 || c.Tuning.DbGCDiscardRation > 1 {
		c.Tuning.DbGCDiscardRation = 0.5
	}
	common.DefaultIfEmptyInt(&c.Tuning.URLCaheSize, 2048)
	common.DefaultIfEmptyInt(&c.Tuning.BckCSVIterPrefetchSize, 2048)
	common.DefaultIfEmptyStr(&c.Tuning.APIKeyHeaderName, "X-API-KEY")

}

//Validate configuration
func (c *ConfigSchema) Validate() {

	if common.IsEmptyStr(c.Server.APIKey) {
		panic("server.apy_key cannot be empty")
	}

	if c.ShortID.Length < 3 {
		panic("short_id.length must be at least 3")
	}

	if len(c.ShortID.Alphabet) < c.ShortID.Length {
		panic(fmt.Sprint("short_id.alphabet must be at least ", c.ShortID.Length, " characters long"))
	}
}

// Config system configuration
var Config ConfigSchema

// GenerateDefaultConfig generate a default configuration file an writes it in the outFile
func GenerateDefaultConfig(outFile, version string) {
	Config.Defaults()
	Config.Server.APIKey = common.GenerateSecret()
	b, _ := yaml.Marshal(Config)
	data := strings.Join([]string{
		"#",
		fmt.Sprintf("# Default configuration for Distill v%s", version),
		"# http://github.com/noandrea/distill",
		"#\n",
		fmt.Sprintf("%s", b),
		"#",
		"# Config end",
		"#",
	}, "\n")
	ioutil.WriteFile(outFile, []byte(data), 0600)
}
