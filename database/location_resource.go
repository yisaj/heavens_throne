package database

import (
	"context"

	"github.com/yisaj/heavens_throne/entities"

	"github.com/pkg/errors"
)

type LocationResource interface {
	GetLocation(ctx context.Context, locationID int32) (*entities.Location, error)
	GetAdjacentLocations(ctx context.Context, locationID int32) ([]int32, error)
	GetTempleLocation(ctx context.Context, order string) (int32, error)
	GetCurrentLogistics(ctx context.Context, order string) ([]entities.Logistic, error)
	GetNextLogistics(ctx context.Context, order string) ([]entities.Logistic, error)
	GetArrivingLogistics(ctx context.Context, locationID int32) ([]entities.Logistic, error)
	GetLeavingLogistics(ctx context.Context, locationID int32) ([]entities.Logistic, error)
	SetLocationOwner(ctx context.Context, locationID int32, owner string) error
}

func (c *connection) GetLocation(ctx context.Context, locationID int32) (*entities.Location, error) {
	query := `SELECT * FROM location WHERE id=$1`

	var location entities.Location
	err := c.db.GetContext(ctx, &location, query, locationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting location")
	}
	return &location, nil
}

func (c *connection) GetAdjacentLocations(ctx context.Context, locationID int32) ([]int32, error) {
	query := `SELECT adjacent FROM adjacent_location WHERE location=$1`

	var adjacentLocations []int32
	err := c.db.SelectContext(ctx, &adjacentLocations, query, locationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting adajcent locations")
	}
	return adjacentLocations, nil
}

func (c *connection) GetTempleLocation(ctx context.Context, order string) (int32, error) {
	query := `SELECT location FROM temple WHERE martial_order=$1`

	var location int32
	err := c.db.GetContext(ctx, &location, query, order)
	if err != nil {
		return -1, errors.Wrap(err, "failed getting temple location")
	}
	return location, nil
}

func (c *connection) GetCurrentLogistics(ctx context.Context, order string) ([]entities.Logistic, error) {
	query := `SELECT location.name, COUNT(*) FROM player
    	INNER JOIN location ON player.location=location.id WHERE player.martial_order=$1 GROUP BY location.name`

	var logistics []entities.Logistic
	err := c.db.SelectContext(ctx, &logistics, query, order)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting current logistics")
	}
	return logistics, nil
}

func (c *connection) GetNextLogistics(ctx context.Context, order string) ([]entities.Logistic, error) {
	query := `SELECT location.name, COUNT(*) FROM player
    	INNER JOIN location ON player.next_location=location.id WHERE player.martial_order=$1 GROUP BY location.name`

	var logistics []entities.Logistic
	err := c.db.SelectContext(ctx, &logistics, query, order)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting next logistics")
	}
	return logistics, nil
}

func (c *connection) GetArrivingLogistics(ctx context.Context, locationID int32) ([]entities.Logistic, error) {
	query := `SELECT prev_location.name, COUNT(*) FROM player
		INNER JOIN location AS next_location ON player.next_location=next_location.id
		INNER JOIN location AS prev_location ON player.location=prev_location.id
		WHERE next_location.id=$1
		GROUP BY prev_location.name`

	var logistics []entities.Logistic
	err := c.db.SelectContext(ctx, &logistics, query, locationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting arriving logistics")
	}
	return logistics, nil
}

func (c *connection) GetLeavingLogistics(ctx context.Context, locationID int32) ([]entities.Logistic, error) {
	query := `SELECT next_location.name, COUNT(*) FROM player
		INNER JOIN location AS next_location ON player.next_location=next_location.id
		INNER JOIN location AS prev_location ON player.location=prev_location.id
		WHERE prev_location.id=$1
		GROUP BY next_location.name`

	var logistics []entities.Logistic
	err := c.db.SelectContext(ctx, &logistics, query, locationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting arriving logistics")
	}
	return logistics, nil
}

func (c *connection) SetLocationOwner(ctx context.Context, locationID int32, owner string) error {
	query := `UPDATE location SET owner=$1 WHERE id=$2`

	_, err := c.db.ExecContext(ctx, query, owner, locationID)
	if err != nil {
		return errors.Wrap(err, "failed setting location owner")
	}
	return nil
}
