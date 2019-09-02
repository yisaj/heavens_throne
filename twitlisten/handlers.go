package twitlisten

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/yisaj/heavens_throne/config"
	"github.com/yisaj/heavens_throne/entities"
	"github.com/yisaj/heavens_throne/input"

	"github.com/sirupsen/logrus"
)

type handler struct {
	mux        *http.ServeMux
	logger     *logrus.Logger
	WebhooksID string
	dmParser   input.DMParser
}

func NewHandler(conf *config.Config, logger *logrus.Logger, dmParser input.DMParser) http.Handler {
	h := &handler{
		http.NewServeMux(),
		logger,
		"",
		dmParser,
	}

	h.mux.HandleFunc(conf.Endpoint, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "", "GET":
			h.handleCRC(w, r, conf.ConsumerKeySecret)
		case "POST":
			h.handleEvent(w, r)
		default:
			w.WriteHeader(400)
		}
	})

	return h
}

func (h *handler) handleCRC(w http.ResponseWriter, r *http.Request, secret string) {
	// get crc_token parameter
	tokens, ok := r.URL.Query()["crc_token"]
	if !ok {
		// no crc_token found
		h.logger.Error("got a crc request with no crc_token")
		w.WriteHeader(400)
		return
	}
	crcToken := tokens[0]

	// hash and encode the crc_token
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(crcToken))
	responseToken := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	// respond to challenge
	responseFmt := `{"response_token":"sha256=%s"}`
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(fmt.Sprintf(responseFmt, responseToken)))
	if err != nil {
		h.logger.WithError(err).Error("failed writing to crc response")
	}
}

func (h *handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	var event entities.Event
	err := json.NewDecoder(r.Body).Decode(&event)
	if err == nil {
		for _, messageEvent := range event.DirectMessageEvents {
			recipientID := messageEvent.MessageCreate.SenderID
			msg := strings.ToLower(messageEvent.MessageCreate.MessageData.Text)
			err = h.dmParser.ParseDM(r.Context(), recipientID, msg)
			if err != nil {
				h.logger.WithError(err).Error("failed parsing direct message")
			}
		}
	}

	w.WriteHeader(200)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("request %s %s", r.Method, r.URL.Path)
	h.mux.ServeHTTP(w, r)
}
