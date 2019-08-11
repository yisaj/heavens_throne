package twitlisten

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/yisaj/heavens_throne/config"
	"github.com/yisaj/heavens_throne/database"
	"github.com/yisaj/heavens_throne/twitspeak"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
)

const (
	maxWebhooksRegistrationAttempts = 5
)

// TODO: pass a message parser to the twitter listener
func Listen(conf *config.Config, speaker twitspeak.TwitterSpeaker, resource database.Resource, logger *logrus.Logger) {
	// check for webhooks id in database
	webhooksID, err := resource.GetWebhooksID(context.TODO())
	if err != nil {
		logger.WithError(err).Fatal("error while performing initial webhooks id check")
	}

	// if not, register the webhook
	if webhooksID == "" {
		for attempts := 1; attempts <= maxWebhooksRegistrationAttempts; attempts++ {
			id, err := speaker.RegisterWebhook()
			if err != nil {
				logger.WithError(err).Fatal("error while registering webhooks url")
			}
			if id != "" {
				webhooksID = id
				break
			}

			time.Sleep(time.Second)
		}
	}

	// autocert manager
	manager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(conf.Domains...),
		Cache:      autocert.DirCache("certs/cache"),
	}

	// auto cert challenge server
	challengeServer := &http.Server{
		Handler: manager.HTTPHandler(nil),
		Addr:    ":http",
	}

	// run challenge server
	go func() {
		logger.Info("starting autocert challenge server")
		err := challengeServer.ListenAndServe()
		if err != nil {
			logger.WithError(err).Fatal("autocert challenge server died")
		}
	}()

	// build the twitter webhooks server
	server := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      NewHandler(conf, logger),
		Addr:         ":https",
	}

	// start listening on https socket
	tlsConf := &tls.Config{
		GetCertificate: manager.GetCertificate,
	}
	listener, err := tls.Listen("tcp", ":https", tlsConf)
	if err != nil {
		logger.WithError(err).Fatal("failed listening on https socket")
	}

	// trigger a CRC manually
	err = speaker.TriggerCRC(webhooksID)
	if err != nil {
		logger.WithError(err).Fatal("error while triggering crc")
	}

	// start serving on the twitter webhooks listener
	logger.Info("starting twitter listener")
	err = server.Serve(listener)
	if err != nil {
		logger.WithError(err).Fatal("twitter listener server died")
	}
}
