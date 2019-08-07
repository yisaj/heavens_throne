package database

import (
	"context"

	"github.com/pkg/errors"
)

type PlayerResource interface {
	CreatePlayer(ctx context.Context, twitterID string, martialOrder string, location string) error
}

func (c *connection) CreatePlayer(ctx context.Context, twitterID string, martialOrder string, location string) error {
	query := `INSERT INTO player (twitter_id, martial_order, location) VALUES ($1, $2, $3)`
	_, err := c.db.ExecContext(ctx, query, twitterID, martialOrder, location)

	if err != nil {
		return errors.Wrap(err, "failed player creation")
	}
	return nil
}
