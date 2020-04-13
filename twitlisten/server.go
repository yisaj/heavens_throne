package twitlisten

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/yisaj/heavens_throne/config"
	"github.com/yisaj/heavens_throne/database"
	"github.com/yisaj/heavens_throne/input"
	"github.com/yisaj/heavens_throne/simulation"
	"github.com/yisaj/heavens_throne/twitspeak"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
)

// Listen spins up the HTTPS autocert server, hooks into the twitter api, and
// starts listening for twitter user events
func Listen(conf *config.Config, speaker twitspeak.TwitterSpeaker, resource database.Resource, logger *logrus.Logger, simLock *simulation.SimLock, simulator simulation.Simulator) {
	// check for webhooks id in database
	webhooksID, err := resource.GetWebhooksID(context.TODO())
	if err != nil {
		logger.WithError(err).Panic("failed querying database for webhook id")
	}

	// check twitter for webhooks id
	webhooksID, err = speaker.GetWebhook()
	if err != nil {
		logger.WithError(err).Panic("failed querying twitter for webhook id")
	}

	// autocert manager
	manager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(conf.Domains...),
		Cache:      autocert.DirCache("certs"),
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
			logger.WithError(err).Panic("autocert challenge server died")
		}
	}()

	// build the twitter webhooks server
	dmParser := input.NewDMParser(resource, speaker, logger, simulator)
	twitterHandler := newHandler(conf, logger, dmParser, speaker, simLock)
	server := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      twitterHandler,
		Addr:         ":https",
	}

	// start listening on https socket
	tlsConf := &tls.Config{
		GetCertificate: manager.GetCertificate,
	}
	listener, err := tls.Listen("tcp", ":https", tlsConf)
	if err != nil {
		logger.WithError(err).Panic("failed listening on https socket")
	}

	if webhooksID != "" {
		// trigger a CRC manually
		go func() {
			err = speaker.TriggerCRC(webhooksID)
			if err != nil {
				logger.WithError(err).Panic("error while triggering crc")
			}
		}()
	} else {
		// register the webhook
		go func(handler *handler) {
			id, err := speaker.RegisterWebhook()
			if err != nil || id == "" {
				logger.WithError(err).Panic("error while registering webhooks url")
			}

			err = resource.SetWebhooksID(context.TODO(), webhooksID)
			if err != nil {
				logger.WithError(err).Panic("error while setting webhooks id in database")
			}
			handler.WebhooksID = webhooksID
		}(twitterHandler.(*handler))
	}

	// start serving on the twitter webhooks listener
	logger.Info("starting twitter listener")
	err = server.Serve(listener)
	if err != nil {
		logger.WithError(err).Panic("twitter listener server died")
	}
}
