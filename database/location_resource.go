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
