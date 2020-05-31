package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/noandrea/distill/pkg/common"
	"github.com/spf13/viper"
)

// ServerConfig configuration for the server
type ServerConfig struct {
	APIKey string `yaml:"api_key" mapstructure:"api_key"`
	Host   string `yaml:"host" mapstructure:"host"`
	Port   int    `yaml:"port" mapstructure:"port"`
}

// DatastoreConfig configure the datastore
type DatastoreConfig struct {
	Name string `yaml:"name" mapstructure:"name"`
	URI  string `yaml:"uri" mapstructure:"uri"`
}

//ShortIDConfig configuration for the short id
type ShortIDConfig struct {
	Alphabet             string    `yaml:"alphabet" mapstructure:"alphabet"`
	Length               int       `yaml:"length" mapstructure:"length"`
	MaxRequests          uint64    `yaml:"max_requests" mapstructure:"max_requests"`
	TTL                  uint64    `yaml:"ttl" mapstructure:"ttl"`
	ExpireOn             time.Time `yaml:"expire_on" mapstructure:"expire_on"`
	RootRedirectURL      string    `yaml:"root_redirect_url" mapstructure:"root_redirect_url"`
	ExpiredRedirectURL   string    `yaml:"expired_redirect_url" mapstructure:"expired_redirect_url"`
	ExhaustedRedirectURL string    `yaml:"exhausted_redirect_url" mapstructure:"exhausted_redirect_url"`
}

// TuningConfig fine tuning configuration
type TuningConfig struct {
	StatsEventsWorkerNum   int     `yaml:"stats_events_worker_num" mapstructure:"stats_events_worker_num"`
	StatsCacheSize         int     `yaml:"stats_cache_size" mapstructure:"stats_cache_size"`
	DbPurgeWritesCount     uint64  `yaml:"db_purge_writes_count" mapstructure:"db_purge_writes_count"`
	DbGCDeletesCount       uint64  `yaml:"db_gc_deletes_count" mapstructure:"db_gc_deletes_count"`
	DbGCDiscardRation      float64 `yaml:"db_gc_discard_ration" mapstructure:"db_gc_discard_ration"`
	URLCacheSize           int     `yaml:"url_cache_size" mapstructure:"url_cache_size"`
	BckCSVIterPrefetchSize int     `yaml:"export_iterator_prefetch_size" mapstructure:"export_iterator_prefetch_size"`
	APIKeyHeaderName       string  `yaml:"api_key_header_name" mapstructure:"api_key_header_name"`
}

// Schema define the configuration object
type Schema struct {
	Server         ServerConfig    `yaml:"server" mapstructure:"server"`
	Datastore      DatastoreConfig `yaml:"datastore" mapstructure:"datastore"`
	ShortID        ShortIDConfig   `yaml:"short_id" mapstructure:"short_id"`
	Tuning         TuningConfig    `yaml:"tuning" mapstructure:"tuning"`
	RuntimeVersion string          `yaml:"-" mapstructure:"-"`
}

func empty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

//Defaults set the defaults for the configuration
func Defaults() {
	// for server
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 1804)
	// datastore
	viper.SetDefault("datastore.name", "embed")
	viper.SetDefault("datastore.uri", "distill.db")
	// for short id
	viper.SetDefault("short_id.root_redirect_url", "https://discover.distill.plus")
	viper.SetDefault("short_id.expired_redirect_url", "https://discover.distill.plus")
	viper.SetDefault("short_id.alphabet", "abcdefghkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789")
	viper.SetDefault("short_id.length", 6)
	// for tuning
	viper.SetDefault("tuning.stats_events_worker_num", 1)
	viper.SetDefault("tuning.stats_cache_size", 1024)
	viper.SetDefault("tuning.db_purge_writes_count", 2000)
	viper.SetDefault("tuning.db_gc_deletes_count", 500)
	viper.SetDefault("tuning.db_gc_discard_ration", 0.5)
	//
	viper.SetDefault("tuning.url_cache_size", 2048)
	viper.SetDefault("tuning.bck_csv_iter_prefetch_size", 2048)
	viper.SetDefault("tuning.api_key_header_name", "X-API-KEY")
}

//Validate configuration
func (c *Schema) Validate() {

	if common.IsEmptyStr(c.Server.APIKey) {
		panic("server.api_key cannot be empty")
	}

	if c.ShortID.Length < 3 {
		panic("short_id.length must be at least 3")
	}

	if len(c.ShortID.Alphabet) < c.ShortID.Length {
		panic(fmt.Sprint("short_id.alphabet must be at least ", c.ShortID.Length, " characters long"))
	}

	if c.Tuning.DbGCDiscardRation <= 0 || c.Tuning.DbGCDiscardRation > 1 {
		panic(fmt.Sprint("tuning.db_gc_discard_ration must be > 0 and < 1"))
	}
}
