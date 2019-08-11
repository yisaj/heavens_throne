package twitspeak

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yisaj/heavens_throne/config"
	"github.com/yisaj/heavens_throne/entities"

	"github.com/pkg/errors"
)

const (
	apiPrefix = "https://api.twitter.com/1.1/account_activity/all"
)

type speaker struct {
	client *http.Client
	conf   *config.Config
}

type TwitterSpeaker interface {
	TriggerCRC(webhookID string) error
	RegisterWebhook() (string, error)
}

func NewSpeaker(conf *config.Config) TwitterSpeaker {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	return &speaker{
		client,
		conf,
	}
}

func (s *speaker) TriggerCRC(webhookID string) error {
	triggerCRCPath := fmt.Sprintf("/%s/webhooks/%s.json", s.conf.TwitterEnvName, webhookID)

	req, err := http.NewRequest("PUT", apiPrefix+triggerCRCPath, nil)
	if err != nil {
		return errors.Wrap(err, "failed building trigger CRC request")
	}

	res, err := s.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed requesting trigger CRC")
	}

	var resBody map[string]string
	err = json.NewDecoder(res.Body).Decode(resBody)
	if err != nil {
		return errors.Wrap(err, "failed decoding trigger CRC response")
	}

	rawTwitterErrors, ok := resBody["errors"]
	if ok {
		var twitterErrors entities.TwitterErrors
		err = json.Unmarshal([]byte(rawTwitterErrors), &twitterErrors)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal twitter errors")
		}
		return errors.Wrap(twitterErrors, "received twitter response errors")
	}

	return nil
}

func (s *speaker) RegisterWebhook() (string, error) {
	registerWebhookPath := fmt.Sprintf("/%s/webhooks.json", s.conf.TwitterEnvName)
	url := s.conf.Domains[0] + s.conf.Endpoint

	res, err := s.client.Post(apiPrefix+registerWebhookPath, "application/json", strings.NewReader(url))
	if err != nil {
		return "", errors.Wrap(err, "failed requesting webhooks registration")
	}

	var resBody map[string]string
	err = json.NewDecoder(res.Body).Decode(resBody)
	if err != nil {
		return "", errors.Wrap(err, "failed decoding register webhook response")
	}

	rawTwitterErrors, ok := resBody["errors"]
	if ok {
		var twitterErrors entities.TwitterErrors
		err = json.Unmarshal([]byte(rawTwitterErrors), &twitterErrors)
		if err != nil {
			return "", errors.Wrap(err, "failed to unmarshal twitter errors")
		}
	}

	return resBody["id"], nil
}
