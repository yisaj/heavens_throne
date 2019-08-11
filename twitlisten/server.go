package twitlisten

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/yisaj/heavens_throne/config"
	"github.com/yisaj/heavens_throne/twitspeak"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
)

func makeServer() *http.Server {
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// TODO: pass a message parser to the twitter listener
func Listen(conf *config.Config, speaker *twitspeak.Speaker, logger *logrus.Logger) {
	// autocert manager
	manager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(conf.Domain),
		Cache:      autocert.DirCache("certs/cache"),
	}

	// auto cert challenge server
	challengeServer := makeServer()
	challengeServer.Handler = manager.HTTPHandler(nil)
	challengeServer.Addr = ":http"

	// run challenge server
	go func() {
		logger.Info("starting autocert challenge server")
		err := challengeServer.ListenAndServe()
		if err != nil {
			logger.WithError(err).Fatal("autocert challenge server died")
		}
	}()

	// build the twitter webhooks server
	server := makeServer()
	server.Handler = NewHandler(conf, logger)
	server.Addr = ":https"
	/*
		server.TLSConfig = &tls.Config{
			GetCertificate: manager.GetCertificate,
		}
	*/

	// start listening on https socket
	tlsConf := &tls.Config{
		GetCertificate: manager.GetCertificate,
	}
	listener, err := tls.Listen("tcp", ":https", tlsConf)
	if err != nil {
		logger.WithError(err).Fatal("failed listening on https socket")
	}

	// trigger a CRC
	logger.Debug("I'm in between listening and serving")

	// start serving on the twitter webhooks listener
	logger.Info("starting twitter listener")
	err = server.Serve(listener)
	if err != nil {
		logger.WithError(err).Fatal("twitter listener server died")
	}
}
