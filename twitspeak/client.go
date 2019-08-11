package twitspeak

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yisaj/heavens_throne/config"
	"github.com/yisaj/heavens_throne/entities"

	"github.com/pkg/errors"
)

const (
	apiPrefix = "https://api.twitter.com/1.1/account_activity/all"
)

type Speaker struct {
	client *http.Client
	conf   *config.Config
}

func NewSpeaker(conf *config.Config) *Speaker {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	return &Speaker{
		client,
		conf,
	}
}

func (s *Speaker) triggerCRC(webhookID string) (*entities.TwitterResponse, error) {
	triggerCRCPath := fmt.Sprintf("/%s/webhooks/%s.json", s.conf.TwitterEnvName, webhookID)

	req, err := http.NewRequest("PUT", apiPrefix+triggerCRCPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed building trigger CRC request")
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed requesting trigger CRC")
	}

	body := &entities.TwitterResponse{}
	err = json.NewDecoder(res.Body).Decode(body)
	if err != nil {
		return nil, errors.Wrap(err, "failed decoding trigger CRC response")
	}

	return body, nil
}
