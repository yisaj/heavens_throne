package entities

import (
	"database/sql"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

type Player struct {
	ID             int32
	TwitterID      string `db:"twitter_id"`
	ReceiveUpdates bool   `db:"receive_updates"`
	Active         bool
	Dead           bool
	MartialOrder   string `db:"martial_order"`
	Location       int32
	Class          string
	Experience     int16
	Rank           int16
}

func (p Player) FormatClass() string {
	classTranslation := map[string]string{
		"recruit":       "Recruit",
		"infantry":      "Infantry",
		"spear":         "Spear",
		"glaivemaster":  "Glaivemaster",
		"sword":         "Sword",
		"legionary":     "Legionary",
		"cavalry":       "Cavalry",
		"heavycavalry":  "Heavy Cavalry",
		"monsterknight": "Monster Knight",
		"lightcavalry":  "Light Cavalry",
		"horsearcher":   "Horse Archer",
		"ranger":        "Ranger",
		"archer":        "Archer",
		"mage":          "Mage",
		"medic":         "Medic",
		"healer":        "Healer",
	}

	rankTranslation := map[int16]string{
		1: "I",
		2: "II",
		3: "III",
		4: "IV",
		5: "V",
	}

	return fmt.Sprintf("%s %s", classTranslation[p.Class], rankTranslation[p.Rank])
}

type Location struct {
	ID    int32
	Name  string
	Owner sql.NullString
}

type TwitterResponse struct {
	Errors []TwitterError
	ID     string
}

func (tr TwitterResponse) GetErrors() error {
	if len(tr.Errors) > 0 {
		var err error = tr.Errors[0]
		for _, twitterErr := range tr.Errors[1:] {
			err = multierror.Append(err, twitterErr)
		}
		return err
	}
	return nil
}

type TwitterError struct {
	Message string
	Code    int32
}

func (te TwitterError) Error() string {
	return fmt.Sprintf("Twitter Err %d: %s", te.Code, te.Message)
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
