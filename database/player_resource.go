package database

import (
	"context"
	"database/sql"

	"github.com/yisaj/heavens_throne/entities"

	"github.com/pkg/errors"
)

type PlayerResource interface {
	CreatePlayer(ctx context.Context, twitterID string, martialOrder string, location string) (*entities.Player, error)
	GetPlayer(ctx context.Context, twitterID string) (*entities.Player, error)
}

func (c *connection) CreatePlayer(ctx context.Context, twitterID string, martialOrder string, location string) (*entities.Player, error) {
	query := `INSERT INTO player (twitter_id, martial_order, location) VALUES ($1, $2, $3) RETURNING *`

	var player entities.Player
	err := c.db.GetContext(ctx, &player, query, twitterID, martialOrder, location)

	if err != nil {
		return nil, errors.Wrap(err, "failed player creation")
	}
	return &player, nil
}

func (c *connection) GetPlayer(ctx context.Context, twitterID string) (*entities.Player, error) {
	query := `SELECT * FROM player WHERE twitter_id=$1`

	var player entities.Player
	err := c.db.GetContext(ctx, &player, query, twitterID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed getting player")
	}
	return &player, nil
}
