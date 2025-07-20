package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Environment string `yaml:"env" env:"ENV" env-default:"development"`
	LogLevel    string `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
}

func New() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./config/config.yml", cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
