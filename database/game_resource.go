package database

import (
	"context"

	"github.com/pkg/errors"
)

type GameResource interface {
	IncrementDay(ctx context.Context) error
}

func (c *connection) IncrementDay(ctx context.Context) error {
	query := `UPDATE player SET location = next_location`

	_, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "failed incrementing the day")
	}
	return nil
}