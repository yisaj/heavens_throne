package main

import (
	"github.com/yisaj/heavens_throne/config"
	"github.com/yisaj/heavens_throne/database"
	"github.com/yisaj/heavens_throne/twitlisten"
	"github.com/yisaj/heavens_throne/twitspeak"

	"github.com/sirupsen/logrus"
)

// TODO: remember to take down game simulator on panic
func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true})

	conf := config.New()

	// spin up connection to database
	_, err := database.Connect(conf)
	if err != nil {
		logger.WithError(err).Fatal("failed database connection")
	}

	// spin up twitter client
	speaker := twitspeak.NewSpeaker(conf)

	// spin up game simulation task

	// spin up twitter webhooks server
	twitlisten.Listen(conf, speaker, logger)

	// stop game simulation task on exit

	logger.Fatal("END")
}
