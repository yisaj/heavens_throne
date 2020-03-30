package entities

import (
	"database/sql"
	"fmt"
)

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

// TODO: rename horsearcher to courser
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
		"horsearcher":   "Horse Archer",
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

type Stats struct {
	Potency int
	Defense int
	Speed   int
	Aggro   int
}

// TODO: consider the stats impact of rank
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

func (p *Player) GetStats() Stats {
	return classBaseStats[p.Class]
}

type Location struct {
	ID       int32
	Name     string
	Owner    sql.NullString
	Occupier sql.NullString
}

type Logistic struct {
	LocationName string `db:"name"`
	Count        int32
}
