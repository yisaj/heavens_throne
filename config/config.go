package config

import (
	"os"
	"strings"
)

const (
	prefix            = "HTHRONE_"
	dbURIKey          = "DB_URI"
	domainsKey        = "DOMAIN"
	endpointKey       = "ENDPOINT"
	consumerSecretKey = "CONSUMER_SECRET"
	twitterEnvNameKey = "TWITTER_ENV_NAME"
	accessTokenKey    = "ACCESS_TOKEN"
)

type Config struct {
	DatabaseURI    string
	Domains        []string
	Endpoint       string
	ConsumerSecret string
	TwitterEnvName string
	AccessToken    string
}

func New() *Config {
	domains := strings.Split(os.Getenv(prefix+domainsKey), ",")

	return &Config{
		DatabaseURI:    os.Getenv(prefix + dbURIKey),
		Domains:        domains,
		Endpoint:       os.Getenv(prefix + endpointKey),
		ConsumerSecret: os.Getenv(prefix + consumerSecretKey),
		TwitterEnvName: os.Getenv(prefix + twitterEnvNameKey),
		AccessToken:    os.Getenv(prefix + accessTokenKey),
	}
}
