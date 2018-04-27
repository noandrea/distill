package internal

// ServerConfig configuration for the server
type ServerConfig struct {
	APIKey       string `yaml:"api-key"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	RootRedirect string `yaml:"root-redirect"`
	DbPath       string `yaml:"db-path"`
}

//ShortIDConfig configureaiont for the short id
type ShortIDConfig struct {
	Alphabet    string `yaml:"alphabet"`
	Length      int    `yaml:"length"`
	MaxRequests int    `yaml:"max-requests"`
	TTL         int    `yaml:"ttl"`
	Domain      string `yaml:"domain"`
}

// ConfigSchema define the configuration object
type ConfigSchema struct {
	Server  ServerConfig  `yaml:"server"`
	ShortID ShortIDConfig `yaml:"short-id"`
}

//Validate configuration
func (c ConfigSchema) Validate() {
	if c.ShortID.Length < 3 {
		panic("short-id.length must be at least 3")
	}

	if len(c.ShortID.Alphabet) < 10 {
		panic("short-id.alphabet must be at least 10 characters long")
	}
}

// Config sytem configuration
var Config ConfigSchema
