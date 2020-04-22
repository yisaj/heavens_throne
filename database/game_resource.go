package database

import (
	"context"

	"github.com/yisaj/heavens_throne/entities"

	"github.com/pkg/errors"
)

// GameResource contains database methods for game data
type GameResource interface {
	GetDay(ctx context.Context) (int32, error)
	IncrementDay(ctx context.Context) error
	CreateCombatRecord(ctx context.Context, locationID int32, event *entities.CombatEvent) error
}

func (c *connection) GetDay(ctx context.Context) (int32, error) {
	query := `SELECT count FROM calendar`

	var day int32
	err := c.db.GetContext(ctx, &day, query)
	if err != nil {
		return -1, errors.Wrap(err, "failed getting current day")
	}

	return day, nil
}

func (c *connection) IncrementDay(ctx context.Context) error {
	query := `UPDATE calendar SET count = count + 1`

	_, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "failed incrementing the day count")
	}

	return nil
}

var combatTypeStrings = map[entities.CombatEventType]string{
	entities.Attack:        "attack",
	entities.CounterAttack: "counterattack",
	entities.Revive:        "Revive",
}

var combatResultStrings = map[entities.CombatResult]string{
	entities.Success:  "success",
	entities.Failure:  "failure",
	entities.NoTarget: "notarget",
}

func (c *connection) CreateCombatRecord(ctx context.Context, locationID int32, event *entities.CombatEvent) error {
	query := `INSERT INTO combat_record (day, location, type, attacker, defender, attacker_class, defender_class, result)
		SELECT count, $1, $2, attacker.id, defender.id, attacker.class, defender.class, $3
		FROM calendar, player AS attacker, player AS defender WHERE attacker.twitter_id=$4 AND defender.twitter_id=$5`

	attacker := event.Attacker
	defender := event.Defender

	_, err := c.db.ExecContext(ctx, query, locationID, combatTypeStrings[event.EventType],
		combatResultStrings[event.Result], attacker.TwitterID, defender.TwitterID)
	if err != nil {
		return errors.Wrap(err, "failed creating combat record")
	}
	return nil
}
