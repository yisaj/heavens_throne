package database

import (
	"context"
	"database/sql"

	"github.com/yisaj/heavens_throne/entities"

	"github.com/pkg/errors"
)

// PlayerResource contains database methods for player data
type PlayerResource interface {
	CreatePlayer(ctx context.Context, twitterID string, martialOrder string, location int32) (*entities.Player, error)
	GetPlayer(ctx context.Context, twitterID string) (*entities.Player, error)
	DeactivatePlayer(ctx context.Context, twitterID string) error
	ClearPlayers(ctx context.Context) error
	DeletePlayer(ctx context.Context, twitterID string) error
	MovePlayer(ctx context.Context, twitterID string, destination int32) error
	TogglePlayerUpdates(ctx context.Context, twitterID string) (bool, error)
	AdvancePlayer(ctx context.Context, twitterID string, class string, rank int16) error
	GetAllPlayers(ctx context.Context) ([]entities.Player, error)
	KillPlayer(ctx context.Context, twitterID string) error
}

func (c *connection) CreatePlayer(ctx context.Context, twitterID string, martialOrder string, location int32) (*entities.Player, error) {
	query := `INSERT INTO player (twitter_id, martial_order, location, next_location) VALUES ($1, $2, $3, $3) RETURNING *`

	var player entities.Player
	err := c.db.GetContext(ctx, &player, query, twitterID, martialOrder, location)

	if err != nil {
		return nil, errors.Wrap(err, "failed player creation")
	}
	return &player, nil
}

// TODO: populate location information
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

func (c *connection) MovePlayer(ctx context.Context, twitterID string, destination int32) error {
	query := `UPDATE player SET next_location=$1 WHERE twitter_id=$2`

	_, err := c.db.ExecContext(ctx, query, destination, twitterID)
	if err != nil {
		return errors.Wrap(err, "failed updating player location")
	}
	return nil
}

func (c *connection) TogglePlayerUpdates(ctx context.Context, twitterID string) (bool, error) {
	query := `UPDATE player SET receive_updates = NOT receive_updates RETURNING receive_updates`

	var receiveUpdates bool
	err := c.db.GetContext(ctx, &receiveUpdates, query)
	if err != nil {
		return false, errors.Wrap(err, "failed toggling player updates setting")
	}
	return receiveUpdates, nil
}

func (c *connection) AdvancePlayer(ctx context.Context, twitterID string, class string, rank int16) error {
	query := `UPDATE player SET class=$1, rank=$2, experience=experience - 100 WHERE twitter_id=$3 RETURNING *`

	_, err := c.db.ExecContext(ctx, query, class, rank, twitterID)
	if err != nil {
		return errors.Wrap(err, "failed advancing player class and rank")
	}
	return nil
}

func (c *connection) GetAllPlayers(ctx context.Context) ([]entities.Player, error) {
	query := `SELECT * FROM player`

	var players []entities.Player
	err := c.db.SelectContext(ctx, &players, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting all players")
	}
	return players, nil
}

func (c *connection) KillPlayer(ctx context.Context, twitterID string) error {
	query := `UPDATE player SET location=
    	(SELECT location FROM temple WHERE location.martial_order=player.martial_order)
		WHERE twitter_id=$1`

	_, err := c.db.ExecContext(ctx, query, twitterID)
	if err != nil {
		return errors.Wrap(err, "failed killing player")
	}
	return nil
}
