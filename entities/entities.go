package entities

import "fmt"

type Player struct {
	ID             int32
	TwitterID      string `db:"twitter_id"`
	ReceiveUpdates bool   `db:"receive_updates"`
	Active         bool
	Dead           bool
	MartialOrder   string `db:"martial_order"`
	Location       string
	Class          string
	Experience     int16
	Rank           int16
}

type TwitterErrors []struct {
	message string
	code    int32
}

func (te TwitterErrors) Error() string {
	body := ""
	for _, err := range te {
		body += fmt.Sprintf("code: %d, msg: %s\n", err.code, err.message)
	}
	return body
}
