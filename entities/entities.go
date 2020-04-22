package entities

import (
	"database/sql"
	"fmt"
)

// TODO ENGINEER: review database structures and optimize

// Player defines the player object, mirroring the database
type Player struct {
	ID             int32
	TwitterID      string `db:"twitter_id"`
	ReceiveUpdates bool   `db:"receive_updates"`
	Active         bool
	MartialOrder   string `db:"martial_order"`
	Location       sql.NullInt32
	NextLocation   sql.NullInt32 `db:"next_location"`
	Class          string
	Experience     int16
	Rank           int16
}

// FormatClass outputs a pretty formatted string of the player's class and rank
func (p *Player) FormatClass() string {
	classTranslation := map[string]string{
		"recruit":       "Initiate",
		"infantry":      "Infantry",
		"spear":         "Spear",
		"glaivemaster":  "Glaivemaster",
		"sword":         "Sword",
		"legionary":     "Legionary",
		"cavalry":       "Cavalry",
		"heavycavalry":  "Heavy Cavalry",
		"monsterknight": "Monster Knight",
		"lightcavalry":  "Light Cavalry",
		"horsearcher":   "Courser",
		"ranger":        "Ranger",
		"archer":        "Archer",
		"mage":          "Mage",
		"medic":         "Medic",
		"healer":        "Healer",
	}

	rankTranslation := map[int16]string{
		1: "I",
		2: "II",
		3: "III",
		4: "IV",
		5: "V",
	}

	return fmt.Sprintf("%s %s", classTranslation[p.Class], rankTranslation[p.Rank])
}

// Stats defines the stats for a given class and rank
type Stats struct {
	Potency int
	Defense int
	Speed   int
	Aggro   int
}

// TODO DESIGN: consider the stats impact of rank
var (
	classBaseStats = map[string]Stats{
		"recruit":       Stats{10, 10, 10, 10},
		"infantry":      Stats{60, 60, 40, 60},
		"cavalry":       Stats{40, 40, 60, 50},
		"ranger":        Stats{50, 50, 50, 40},
		"spear":         Stats{70, 70, 50, 70},
		"sword":         Stats{70, 70, 50, 70},
		"heavycavalry":  Stats{50, 50, 70, 60},
		"lightcavalry":  Stats{50, 50, 70, 60},
		"archer":        Stats{60, 60, 60, 50},
		"medic":         Stats{60, 60, 60, 50},
		"legionary":     Stats{80, 80, 60, 80},
		"glaivemaster":  Stats{80, 80, 60, 80},
		"monsterknight": Stats{60, 60, 80, 70},
		"horsearcher":   Stats{60, 60, 80, 70},
		"mage":          Stats{70, 70, 70, 60},
		"healer":        Stats{70, 70, 70, 60},
	}
)

// GetStats returns the stats for a given player's class and rank
func (p *Player) GetStats() Stats {
	return classBaseStats[p.Class]
}

// IsRanged returns whether the player is a ranged class
func (p *Player) IsRanged() bool {
	return p.Class == "archer" || p.Class == "mage"
}

func (p *Player) IsAlive() bool {
	return p.Location.Valid
}

// Location defines the location object, mirroring the database
type Location struct {
	ID       int32
	Name     string
	Owner    sql.NullString
	Occupier sql.NullString
}

// Logistic defines the logistic object, which provides unit counts relative to
// a location. mirrors the database
type Logistic struct {
	LocationName string `db:"name"`
	Count        int32
}

// CombatEventType denotes the actions that can be taken during combat
type CombatEventType int

// All the combat event types
const (
	Attack CombatEventType = iota
	CounterAttack
	Revive
)

// CombatResult denotes the possible outcomes of combat
type CombatResult int

// TODO ENGINEER: this naming might be too general and conflict with other stuff eventually
// All the combat results
const (
	Success CombatResult = iota
	Failure
	NoTarget
)

// CombatEvent details what happened in a particular instance of combat
type CombatEvent struct {
	Attacker  *Player
	Defender  *Player
	EventType CombatEventType
	Result    CombatResult
}
