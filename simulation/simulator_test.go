package simulation

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/bsm/bst"
	"github.com/yisaj/heavens_throne/entities"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UTC().UnixNano())
	code := m.Run()
	os.Exit(code)
}

func initializePlayers() map[string][]entities.Player {
	players := make(map[string][]entities.Player)

	numSoldiers := 3
	var classes = []string{
		"recruit", "infantry", "cavalry", "ranger", "spear", "sword",
		"heavycavalry", "lightcavalry", "archer", "medic",
		"glaivemaster", "legionary", "monsterknight", "horsearcher",
		"mage", "healer",
	}
	var orders = []string{
		"Staghorn Sect", "Order Gorgona", "The Baaturate",
	}

	totalPlayers := 0
	for _, order := range orders {
		for _, class := range classes {
			for i := 0; i < numSoldiers; i++ {
				players[order] = append(players[order], entities.Player{
					ID:           int32(totalPlayers),
					Class:        class,
					Rank:         1,
					MartialOrder: order,
				})

				totalPlayers++
			}
		}
	}

	return players
}

func TestCalculateAttackOrder(t *testing.T) {
	players := initializePlayers()
	sim := NewNormalSimulator(nil, nil, nil)
	attackOrder := sim.calculateAttackOrder(players)

	for it := attackOrder.Iterator(); it.Next(); {
		found := false
		orderedPlayer := it.Value().(*entities.Player)
		for _, player := range players[orderedPlayer.MartialOrder] {
			if &player == orderedPlayer {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing player: %d\n", orderedPlayer.ID)
			break
		}
	}

	for iter := attackOrder.Iterator(); iter.Next(); {
		t.Logf("%f - %s\n", iter.Key().(bst.Float64), iter.Value().(*entities.Player).Class)
	}
}

func TestAttackTarget(t *testing.T) {
	attacker := entities.Player{
		ID:           0,
		Class:        "sword",
		Rank:         1,
		MartialOrder: "Order Gorgona",
	}

	defender := entities.Player{
		ID:           1,
		Class:        "archer",
		Rank:         1,
		MartialOrder: "The Baaturate",
	}

	sim := NewNormalSimulator(nil, nil, nil)
	event := sim.attackTarget(&attacker, &defender, 0)
	t.Logf("%+v\n", event)
}

func TestBattleSimulation(t *testing.T) {
	players := initializePlayers()
	simulator := NewNormalSimulator(nil, nil, nil)
	survivors, fatalities, combatEvents, err := simulator.SimulateBattle(0, players)
	if err != nil {
		t.Error(err)
	}

	for _, survivor := range survivors {
		t.Logf("%+v\n", survivor)
	}
	for _, fatality := range fatalities {
		t.Logf("%+v\n", fatality)
	}
	for _, event := range combatEvents {
		t.Logf("%+v\n", event)
	}
}
