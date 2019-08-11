package config

import (
	"os"
)

const (
	prefix            = "HTHRONE_"
	dbURIKey          = "DB_URI"
	domainKey         = "DOMAIN"
	endpointKey       = "ENDPOINT"
	consumerSecretKey = "CONSUMER_SECRET"
	twitterEnvNameKey = "TWITTER_ENV_NAME"
)

type Config struct {
	DatabaseURI    string
	Domain         string
	Endpoint       string
	ConsumerSecret string
	TwitterEnvName string
}

func New() *Config {
	return &Config{
		DatabaseURI:    os.Getenv(prefix + dbURIKey),
		Domain:         os.Getenv(prefix + domainKey),
		Endpoint:       os.Getenv(prefix + endpointKey),
		ConsumerSecret: os.Getenv(prefix + consumerSecretKey),
		TwitterEnvName: os.Getenv(prefix + twitterEnvNameKey),
	}
}
