package internal

// ConfigSchema define the configuration object
type ConfigSchema struct {
	Server struct {
		APIKey       string `yaml:"api-key"`
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		RootRedirect string `yaml:"root-redirect"`
		DbPath       string `yaml:"db-path"`
	} `yaml:"server"`
	ShortID struct {
		CharacterSet string `yaml:"character-set"`
		Length       int    `yaml:"length"`
		MaxRequests  int    `yaml:"max-requests"`
		TTL          int    `yaml:"ttl"`
	} `yaml:"short-id"`
}

// Config sytem configuration
var Config ConfigSchema
