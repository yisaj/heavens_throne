package config

import (
	"os"
)

const (
	prefix         = "HTHRONE_"
	placeholderKey = "PLACEHOLDER"
)

type Config struct {
	Placeholder string
}

func New() *Config {
	return &Config{
		os.Getenv(prefix + placeholderKey),
	}
}
