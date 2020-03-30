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
	"github.com/yisaj/heavens_throne/input"
	"github.com/yisaj/heavens_throne/simulation"
	"github.com/yisaj/heavens_throne/twitspeak"

	"github.com/sirupsen/logrus"
)

type handler struct {
	mux        *http.ServeMux
	logger     *logrus.Logger
	WebhooksID string
	dmParser   input.DMParser
	speaker twitspeak.TwitterSpeaker
	simlock *simulation.SimLock
}

func NewHandler(conf *config.Config, logger *logrus.Logger, dmParser input.DMParser, speaker twitspeak.TwitterSpeaker, simlock *simulation.SimLock) http.Handler {
	h := &handler{
		http.NewServeMux(),
		logger,
		"",
		dmParser,
		speaker,
		simlock,
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

type Event struct {
	ForUserID           string `json:"for_user_id"`
	DirectMessageEvents []struct {
		MessageCreate struct {
			SenderID    string `json:"sender_id"`
			MessageData struct {
				Text string
			} `json:"message_data"`
		} `json:"message_create"`
	} `json:"direct_message_events"`
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
	const busySimulating = `
I'm busy simulating right now.'
`

	var event Event
	err := json.NewDecoder(r.Body).Decode(&event)
	if err == nil {
		for _, messageEvent := range event.DirectMessageEvents {
			recipientID := messageEvent.MessageCreate.SenderID
			if recipientID == event.ForUserID {
				continue
			}

			simulating := h.simlock.Check()
			if simulating {
				err = h.speaker.SendDM(recipientID, busySimulating)
				h.simlock.RUnlock()
				continue
			}

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
