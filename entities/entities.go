package entities

import (
	"database/sql"
	"fmt"
)

// Player defines the player object, mirroring the database
type Player struct {
	ID             int32
	TwitterID      string `db:"twitter_id"`
	ReceiveUpdates bool   `db:"receive_updates"`
	Active         bool
	Dead           bool
	MartialOrder   string `db:"martial_order"`
	Location       int32
	NextLocation   int32 `db:"next_location"`
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
