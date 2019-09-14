package database

import (
	"context"
	"database/sql"

	"github.com/yisaj/heavens_throne/entities"

	"github.com/pkg/errors"
)

type PlayerResource interface {
	CreatePlayer(ctx context.Context, twitterID string, martialOrder string, location int32) (*entities.Player, error)
	GetPlayer(ctx context.Context, twitterID string) (*entities.Player, error)
	DeactivatePlayer(ctx context.Context, twitterID string) error
	ClearPlayers(ctx context.Context) error
	DeletePlayer(ctx context.Context, twitterID string) error
	MovePlayer(ctx context.Context, twitterID string, destination int32) (*entities.Location, error)
}

func (c *connection) CreatePlayer(ctx context.Context, twitterID string, martialOrder string, location int32) (*entities.Player, error) {
	query := `INSERT INTO player (twitter_id, martial_order, location) VALUES ($1, $2, $3) RETURNING *`

	var player entities.Player
	err := c.db.GetContext(ctx, &player, query, twitterID, martialOrder, location)

	if err != nil {
		return nil, errors.Wrap(err, "failed player creation")
	}
	return &player, nil
}

func (c *connection) GetPlayer(ctx context.Context, twitterID string) (*entities.Player, error) {
	query := `SELECT * FROM player WHERE twitter_id=$1 AND active=true`

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

func (c *connection) DeactivatePlayer(ctx context.Context, twitterID string) error {
	query := `UPDATE player SET active=false WHERE twitter_id=$1`

	_, err := c.db.ExecContext(ctx, query, twitterID)
	if err != nil {
		return errors.Wrap(err, "failed deactivating player")
	}
	return nil
}

func (c *connection) ClearPlayers(ctx context.Context) error {
	query := `TRUNCATE player`

	_, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "failed clearing players")
	}
	return nil
}

func (c *connection) DeletePlayer(ctx context.Context, twitterID string) error {
	query := `DELETE FROM player WHERE twitter_id=$1`

	_, err := c.db.ExecContext(ctx, query, twitterID)
	if err != nil {
		return errors.Wrap(err, "failed deleting player")
	}
	return nil
}

func (c *connection) MovePlayer(ctx context.Context, twitterID string, destination int32) (*entities.Location, error) {
	query := `UPDATE player SET location=$1 WHERE twitter_id=$2 RETURNING *`

	var location entities.Location
	err := c.db.GetContext(ctx, &location, query, destination, twitterID)
	if err != nil {
		return nil, errors.Wrap(err, "failed updating player location")
	}
	return &location, nil
}
