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
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/hashicorp/go-multierror"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yisaj/heavens_throne/config"
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
	logger *logrus.Logger
}

// TwitterSpeaker contains the methods to send the twitter api HTTPS messages
type TwitterSpeaker interface {
	TriggerCRC(webhookID string) error
	GetWebhook() (string, error)
	RegisterWebhook() (string, error)
	SendDM(userID string, msg string) error
	SubscribeUser() error
	Tweet(msg string, target string) (string, error)
}

// twitterError is the standard error format for a twitter api error
type twitterError struct {
	Message string
	Code    int32
}

// Error fulfils the error interface for twitterError
func (te twitterError) Error() string {
	return fmt.Sprintf("Twitter Err %d: %s", te.Code, te.Message)
}

// mergeTwitterErrors parses out all the errors in a twitter api response
func mergeTwitterErrors(te []twitterError) error {
	if len(te) > 0 {
		var err error = te[0]
		for _, twitterErr := range te[1:] {
			err = multierror.Append(err, twitterErr)
		}
		return err
	}
	return nil
}

// NewSpeaker returns a new speaker to send messages to the twitter api with
func NewSpeaker(conf *config.Config, logger *logrus.Logger) TwitterSpeaker {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	return &speaker{
		client,
		conf,
		logger,
	}
}

func percentEscape(s string) string {
	result := make([]byte, 0, len(s)*3)
	for _, b := range []byte(s) {
		if isPercentEscapable(b) {
			result = append(result, '%')
			result = append(result, "0123456789ABCDEF"[b>>4])
			result = append(result, "0123456789ABCDEF"[b&15])
		} else {
			result = append(result, b)
		}
	}
	return string(result)
}

func isPercentEscapable(b byte) bool {
	return !('A' <= b && 'Z' >= b || 'a' <= b && 'z' >= b || '0' <= b && '9' >= b || '-' == b || '_' == b || '.' == b || '~' == b)
}

// generateNonce returns a one time psuedo random string
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

// authorizeRequest fills in an http request so that the twitter api will accept
// it as authorized
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

	params := make(map[string]string)
	// collect body parameters
	if req.Body != nil {
		err := req.ParseForm()
		if err != nil {
			return errors.Wrap(err, "failed parsing request parameters")
		}
		for key, values := range req.PostForm {
			if _, ok := params[key]; ok {
				return errors.New("authorization error: duplicate post form keys")
			}

			params[key] = values[0]
		}
	}

	// collect url parameters
	for key, values := range req.URL.Query() {
		if _, ok := params[key]; ok {
			return errors.New("authorization error: duplicate url keys")
		}

		params[key] = values[0]
	}

	params["oauth_consumer_key"] = consumerKey
	params["oauth_token"] = userToken
	params["oauth_signature_method"] = signatureMethod
	params["oauth_timestamp"] = timestamp
	params["oauth_nonce"] = nonce
	params["oauth_version"] = version

	// build the signature field
	signature := strings.Builder{}
	signature.WriteString(reqMethod)
	signature.WriteByte('&')
	signature.WriteString(percentEscape(reqURL))
	signature.WriteByte('&')

	sortedKeys := make([]string, 0, len(params))
	for key := range params {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	paramBuilder := strings.Builder{}
	for _, key := range sortedKeys {
		paramBuilder.WriteString(key)
		paramBuilder.WriteByte('=')
		paramBuilder.WriteString(percentEscape(params[key]))
		paramBuilder.WriteByte('&')
	}
	paramString := strings.TrimSuffix(paramBuilder.String(), "&")

	signature.WriteString(percentEscape(paramString))
	signingKey := percentEscape(s.conf.ConsumerKeySecret) + "&" + percentEscape(s.conf.AccessTokenSecret)
	hash := hmac.New(sha1.New, []byte(signingKey))
	hash.Write([]byte(signature.String()))
	hashedSignature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	// build the authorization header
	authHeader := strings.Builder{}
	authHeader.WriteString(fmt.Sprintf("OAuth oauth_consumer_key=\"%s\",", consumerKey))
	authHeader.WriteString(fmt.Sprintf("oauth_token=\"%s\",", userToken))
	authHeader.WriteString(fmt.Sprintf("oauth_signature_method=\"%s\",", signatureMethod))
	authHeader.WriteString(fmt.Sprintf("oauth_timestamp=\"%s\",", timestamp))
	authHeader.WriteString(fmt.Sprintf("oauth_nonce=\"%s\",", nonce))
	authHeader.WriteString(fmt.Sprintf("oauth_version=\"%s\",", version))
	authHeader.WriteString(fmt.Sprintf("oauth_signature=\"%s\"", percentEscape(hashedSignature)))
	req.Header.Set("Authorization", authHeader.String())
	return nil
}

// TriggerCRC sends a message to the twitter api requesting a challenge response
// check. used on application wakeup to connect to the twitter api
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

	type triggerCRCResponse struct {
		Errors []twitterError
	}
	var triggerCRCRes triggerCRCResponse
	err = json.NewDecoder(res.Body).Decode(&triggerCRCRes)
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "failed decoding trigger CRC response")
	}

	err = mergeTwitterErrors(triggerCRCRes.Errors)
	if err != nil {
		return errors.Wrap(err, "trigger CRC response errors")
	}
	return nil
}

// GetWebhook gets the webhook id from the twitter api
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

	if res.StatusCode != 200 {
		type getWebhookResponse struct {
			Errors []twitterError
		}
		var getWebhookRes getWebhookResponse
		err = json.NewDecoder(res.Body).Decode(&getWebhookRes)
		if err != nil {
			return "", errors.Wrap(err, "failed decoding get webhooks twitter errors")
		}
		return "", errors.Wrap(mergeTwitterErrors(getWebhookRes.Errors), "get webhooks errors")
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

// RegisterWebhook registers the app as a new webhook with twitter and returns the id
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

	type regWebhookResponse struct {
		ID     string
		Errors []twitterError
	}
	var regWebhookRes regWebhookResponse
	err = json.NewDecoder(res.Body).Decode(&regWebhookRes)
	if err != nil {
		return "", errors.Wrap(err, "failed decoding register webhook response")
	}

	err = mergeTwitterErrors(regWebhookRes.Errors)
	if err != nil {
		return "", errors.Wrap(err, "register webhook response errors")
	}
	return regWebhookRes.ID, nil
}

// SendDM sends a twitter direct message to a given user
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

	type dmResponse struct {
		Errors []twitterError
	}
	var dmRes dmResponse
	err = json.NewDecoder(res.Body).Decode(&dmRes)
	if err != nil {
		return errors.Wrap(err, "failed decoding post direct message response")
	}

	err = mergeTwitterErrors(dmRes.Errors)
	if err != nil {
		return errors.Wrap(err, "post direct message response errors")
	}
	return nil
}

// SubscribeUser subscribes to the heavens throne user account in order to receive
// user events
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

	type subscribeResponse struct {
		Errors []twitterError
	}
	var subscribeRes subscribeResponse
	err = json.NewDecoder(res.Body).Decode(&subscribeRes)
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "failed decoding trigger CRC response")
	}

	err = mergeTwitterErrors(subscribeRes.Errors)
	if err != nil {
		return errors.Wrap(err, "subscribe user response errors")
	}
	return nil
}

func (s *speaker) Tweet(msg string, target string) (string, error) {
	tweetPath := fmt.Sprintf("/statuses/update.json?status=%s", percentEscape(msg))
	if target != "" {
		tweetPath += fmt.Sprintf("&in_reply_to_status_id=%s", target)
	}

	req, err := http.NewRequest("POST", apiPrefix+tweetPath, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed building tweet request")
	}

	err = s.authorizeRequest(req)
	if err != nil {
		return "", errors.Wrap(err, "failed authorizing tweet request")
	}

	res, err := s.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed tweet request")
	}

	type tweetResponse struct {
		ID_str string
		Errors []twitterError
	}
	var tweetRes tweetResponse
	err = json.NewDecoder(res.Body).Decode(&tweetRes)
	if err != nil && err != io.EOF {
		return "", errors.Wrap(err, "failed decoding tweet response")
	}

	err = mergeTwitterErrors(tweetRes.Errors)
	if res.StatusCode != 200 {
		if err != nil {
			return "", errors.Wrap(err, "send tweet response errors")
		} else {
			return "", fmt.Errorf("send tweet twitter response with code: %d", res.StatusCode)
		}
	}

	return tweetRes.ID_str, nil
}
