package config

import (
	"os"
)

const (
	prefix   = "HTHRONE_"
	dbURIKey = "DB_URI"
)

type Config struct {
	DatabaseURI string
}

func New() *Config {
	return &Config{
		DatabaseURI: os.Getenv(prefix + dbURIKey),
	}
}
