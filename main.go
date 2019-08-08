package main

import (
	"github.com/sirupsen/logrus"

	"github.com/yisaj/heavens_throne/config"
	"github.com/yisaj/heavens_throne/database"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true})

	conf := config.New()
	_, err := database.Connect(conf)
	if err != nil {
		logger.WithError(err).Fatal("failed database connection")
	}

	logger.Fatal("END")
}
