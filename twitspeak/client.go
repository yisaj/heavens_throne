package twitspeak

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/yisaj/heavens_throne/config"
	"github.com/yisaj/heavens_throne/entities"

	"github.com/pkg/errors"
)

const (
	apiPrefix  = "https://api.twitter.com/1.1"
	nonceRunes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	nonceMax   = 6
	nonceMask  = 1<<uint(nonceMax) - 1
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
)

type speaker struct {
	client *http.Client
	conf   *config.Config
}

type TwitterSpeaker interface {
	TriggerCRC(webhookID string) error
	GetWebhook() (string, error)
	RegisterWebhook() (string, error)
	SendDM(userID string, msg string) error
	SubscribeUser() error
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

func generateNonce(length int) string {
	nonce := make([]byte, length)

	for i, cache, remain := 0, randSource.Int63(), nonceMax; i < length; remain-- {
		if remain == 0 {
			cache, remain = randSource.Int63(), nonceMax
		}
		if index := int(cache & nonceMask); index < len(nonceRunes) {
			nonce[i] = nonceRunes[index]
			i++
		}
		cache >>= nonceMax
	}

	return *(*string)(unsafe.Pointer(&nonce))
}

func (s *speaker) authorizeRequest(req *http.Request) error {
	// collect the values for the authorization header
	consumerKey := s.conf.ConsumerKey
	userToken := s.conf.AccessToken
	nonce := generateNonce(32)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signatureMethod := "HMAC-SHA1"
	version := "1.0"

	// collect the values for the signature field
	reqMethod := req.Method
	reqURL := req.URL.Scheme + "://" + req.URL.Host + req.URL.EscapedPath()

	var params url.Values
	if req.Body != nil {
		err := req.ParseForm()
		if err != nil {
			return errors.Wrap(err, "failed parsing request parameters")
		}
		params = req.Form
	} else {
		params = req.URL.Query()
	}

	params.Add("oauth_consumer_key", consumerKey)
	params.Add("oauth_token", userToken)
	params.Add("oauth_nonce", nonce)
	params.Add("oauth_timestamp", timestamp)
	params.Add("oauth_signature_method", signatureMethod)
	params.Add("oauth_version", version)

	// build the signature field
	signature := strings.Builder{}
	signature.WriteString(reqMethod)
	signature.WriteByte('&')
	signature.WriteString(url.QueryEscape(reqURL))
	signature.WriteByte('&')

	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	paramString := strings.Builder{}
	for _, key := range keys {
		values := params[key]
		paramString.WriteString(key)
		paramString.WriteByte('=')
		paramString.WriteString(url.QueryEscape(values[0]))
		paramString.WriteByte('&')
	}
	signature.WriteString(url.QueryEscape(strings.TrimSuffix(paramString.String(), "&")))
	signingKey := url.QueryEscape(s.conf.ConsumerKeySecret) + "&" + url.QueryEscape(s.conf.AccessTokenSecret)
	hash := hmac.New(sha1.New, []byte(signingKey))
	hash.Write([]byte(signature.String()))
	hashedSignature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	// build the authorization header
	authHeader := strings.Builder{}
	authHeader.WriteString(fmt.Sprintf("OAuth oauth_consumer_key=\"%s\", ", consumerKey))
	authHeader.WriteString(fmt.Sprintf("oauth_nonce=\"%s\", ", nonce))
	authHeader.WriteString(fmt.Sprintf("oauth_signature=\"%s\", ", url.QueryEscape(hashedSignature)))
	authHeader.WriteString(fmt.Sprintf("oauth_signature_method=\"%s\", ", signatureMethod))
	authHeader.WriteString(fmt.Sprintf("oauth_timestamp=\"%s\", ", timestamp))
	authHeader.WriteString(fmt.Sprintf("oauth_token=\"%s\", ", userToken))
	authHeader.WriteString(fmt.Sprintf("oauth_version=\"%s\"", version))

	req.Header.Set("Authorization", authHeader.String())
	return nil
}

func (s *speaker) TriggerCRC(webhookID string) error {
	// send a request to the twitter API to manually trigger a challenge-response check
	triggerCRCPath := fmt.Sprintf("/account_activity/all/%s/webhooks/%s.json", s.conf.TwitterEnvName, webhookID)
	req, err := http.NewRequest("PUT", apiPrefix+triggerCRCPath, nil)
	if err != nil {
		return errors.Wrap(err, "failed building trigger CRC request")
	}

	err = s.authorizeRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed authorizing trigger CRC request")
	}

	res, err := s.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed requesting trigger CRC")
	}

	var twitterRes entities.TwitterResponse
	err = json.NewDecoder(res.Body).Decode(&twitterRes)
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "failed decoding trigger CRC response")
	}

	err = twitterRes.GetErrors()
	if err != nil {
		return errors.Wrap(err, "trigger CRC response errors")
	}
	return nil
}

func (s *speaker) GetWebhook() (string, error) {
	getWebhookPath := fmt.Sprintf("/account_activity/all/%s/webhooks.json", s.conf.TwitterEnvName)

	req, err := http.NewRequest("GET", apiPrefix+getWebhookPath, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed building get webhooks request")
	}

	err = s.authorizeRequest(req)
	if err != nil {
		return "", errors.Wrap(err, "failed authorizing get webhooks request")
	}

	res, err := s.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed requesting get webhook")
	}

	var webhooks []interface{}
	err = json.NewDecoder(res.Body).Decode(&webhooks)
	if err != nil {
		return "", errors.Wrap(err, "failed decoding get webhooks response")
	}
	if len(webhooks) > 0 {
		webhook := webhooks[0].(map[string]interface{})
		return webhook["id"].(string), nil
	}
	return "", nil
}

func (s *speaker) RegisterWebhook() (string, error) {
	// send a request to the twitter API to register the configured URL as a webhook
	registerWebhookPath := fmt.Sprintf("/account_activity/all/%s/webhooks.json", s.conf.TwitterEnvName)
	webhookURL := "https://" + s.conf.Domains[0] + s.conf.Endpoint

	req, err := http.NewRequest("POST", apiPrefix+registerWebhookPath, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed building webhooks registration request")
	}
	query := req.URL.Query()
	query.Add("url", webhookURL)
	req.URL.RawQuery = query.Encode()

	err = s.authorizeRequest(req)
	if err != nil {
		return "", errors.Wrap(err, "failed authorizing webhooks registration request")
	}

	res, err := s.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed requesting webhooks registration")
	}

	var twitterRes entities.TwitterResponse
	err = json.NewDecoder(res.Body).Decode(&twitterRes)
	if err != nil {
		return "", errors.Wrap(err, "failed decoding register webhook response")
	}

	err = twitterRes.GetErrors()
	if err != nil {
		return "", errors.Wrap(err, "register webhook response errors")
	}
	return twitterRes.ID, nil
}

func (s *speaker) SendDM(userID string, msg string) error {
	// escape common control characters
	replacer := strings.NewReplacer("\n", `\n`, "\r", `\r`, "\t", `\t`)
	msg = replacer.Replace(msg)

	sendDMPath := "/direct_messages/events/new.json"
	eventFmt := `{"event": { "type": "message_create", 
		"message_create": {
			"target": {"recipient_id": "%s"},
			"message_data":{"text": "%s"}}}}`
	eventString := fmt.Sprintf(eventFmt, userID, msg)

	req, err := http.NewRequest("POST", apiPrefix+sendDMPath, strings.NewReader(eventString))
	if err != nil {
		return errors.Wrap(err, "failed building post direct message request")
	}

	err = s.authorizeRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed authorizing post direct message request")
	}

	res, err := s.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed posting direct message")
	}

	var twitterRes entities.TwitterResponse
	err = json.NewDecoder(res.Body).Decode(&twitterRes)
	if err != nil {
		return errors.Wrap(err, "failed decoding post direct message response")
	}

	err = twitterRes.GetErrors()
	if err != nil {
		return errors.Wrap(err, "post direct message response errors")
	}
	return nil
}

func (s *speaker) SubscribeUser() error {
	subscribeUserPath := fmt.Sprintf("/%s/subscriptions.json", s.conf.TwitterEnvName)
	req, err := http.NewRequest("POST", apiPrefix+subscribeUserPath, nil)
	if err != nil {
		return errors.Wrap(err, "failed building user subscription request")
	}

	err = s.authorizeRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed authorizing user subscription request")
	}

	res, err := s.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed requesting user subscription")
	}

	var twitterRes entities.TwitterResponse
	err = json.NewDecoder(res.Body).Decode(&twitterRes)
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "failed decoding trigger CRC response")
	}

	err = twitterRes.GetErrors()
	if err != nil {
		return errors.Wrap(err, "subscribe user response errors")
	}
	return nil
}
