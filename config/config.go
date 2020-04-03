package config

import (
	"os"
	"strings"
)

const (
	prefix               = "HTHRONE_"
	dbURIKey             = "DB_URI"
	domainsKey           = "DOMAIN"
	endpointKey          = "ENDPOINT"
	consumerKeyKey       = "CONSUMER_KEY"
	consumerKeySecretKey = "CONSUMER_KEY_SECRET"
	twitterEnvNameKey    = "TWITTER_ENV_NAME"
	accessTokenKey       = "ACCESS_TOKEN"
	accessTokenSecretKey = "ACCESS_TOKEN_SECRET"
	debugKey             = "DEBUG"
)

// Config defines the database and twitter configuration for the app
type Config struct {
	DatabaseURI       string
	Domains           []string
	Endpoint          string
	ConsumerKey       string
	ConsumerKeySecret string
	TwitterEnvName    string
	AccessToken       string
	AccessTokenSecret string
	Debug             string
}

// New returns a new config object constructed from environment variables
func New() *Config {
	domains := strings.Split(os.Getenv(prefix+domainsKey), ",")

	return &Config{
		DatabaseURI:       os.Getenv(prefix + dbURIKey),
		Domains:           domains,
		Endpoint:          os.Getenv(prefix + endpointKey),
		ConsumerKey:       os.Getenv(prefix + consumerKeyKey),
		ConsumerKeySecret: os.Getenv(prefix + consumerKeySecretKey),
		TwitterEnvName:    os.Getenv(prefix + twitterEnvNameKey),
		AccessToken:       os.Getenv(prefix + accessTokenKey),
		AccessTokenSecret: os.Getenv(prefix + accessTokenSecretKey),
		Debug:             os.Getenv(prefix + debugKey),
	}
}
