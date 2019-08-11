package twitlisten

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/yisaj/heavens_throne/config"

	"github.com/sirupsen/logrus"
)

type handler struct {
	mux    *http.ServeMux
	logger *logrus.Logger
}

func NewHandler(conf *config.Config, logger *logrus.Logger) http.Handler {
	h := &handler{
		http.NewServeMux(),
		logger,
	}

	h.mux.HandleFunc(conf.Endpoint, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "", "GET":
			h.handleCRC(w, r, conf.ConsumerSecret)
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
	responseFmt := `{"response_token":"sha=%s"}`
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(fmt.Sprintf(responseFmt, responseToken)))
	if err != nil {
		h.logger.WithError(err).Error("failed writing to crc response")
	}
}

func (h *handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("request %s %s", r.Method, r.URL.Path)
	h.mux.ServeHTTP(w, r)
}
