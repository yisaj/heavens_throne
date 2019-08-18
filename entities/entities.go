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

type TwitterResponse struct {
	Errors []TwitterError
	ID     string
}

type TwitterError struct {
	Message string
	Code    int32
}

func (te TwitterError) Error() string {
	return fmt.Sprintf("Twitter Err %d: %s", te.Code, te.Message)
}
