package config

import (
	"github.com/kelseyhightower/envconfig"
)

var config = New()

// Config holds the configuration of the service
type Config struct {
	DBName     string `split_words:"true"`
	DBUser     string `split_words:"true"`
	DBPassword string `split_words:"true"`
	DBHost     string `split_words:"true"`
	DBPort     string `split_words:"true"`
	HTTPPort   string `split_words:"true"`
}

// New returns a new config with defaults
func New() *Config {
	return &Config{
		HTTPPort: "8000",
	}
}

// Load loads the config from environment variables
func (c *Config) Load(appName string) error {
	return envconfig.Process(appName, c)
}

// Get returns the global config instance
func Get() *Config {
	return config
}
