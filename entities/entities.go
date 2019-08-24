package entities

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

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

func (tr TwitterResponse) GetErrors() error {
	if len(tr.Errors) > 0 {
		var err error = tr.Errors[0]
		for _, twitterErr := range tr.Errors[1:] {
			err = multierror.Append(err, twitterErr)
		}
		return errors.Wrap(err, "register webhook response errors")
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
