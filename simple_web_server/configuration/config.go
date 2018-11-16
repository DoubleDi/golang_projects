package configuration

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	DBUser string `yaml:"DBUser"`
	DBPass string `yaml:"DBPass"`
	DBHost string `yaml:"DBHost"`
	DBPort string `yaml:"DBPort"`
	DBName string `yaml:"DBName"`
}

func Init(configPath *string) (*Config, error) {
	config := &Config{}

	configFile, err := ioutil.ReadFile(*configPath)
	if err != nil {
		return nil, err
	}
	yaml.Unmarshal(configFile, &config)

	return config, nil
}
