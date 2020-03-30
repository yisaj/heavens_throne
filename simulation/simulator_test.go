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

func initializePlayers() []*entities.Player {
	var players []*entities.Player

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
				players = append(players, &entities.Player{
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
	attackOrder := calculateAttackOrder(players)

	for it := attackOrder.Iterator(); it.Next(); {
		found := false
		orderedPlayer := it.Value().(*entities.Player)
		for _, player := range players {
			if player == orderedPlayer {
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

	event := attackTarget(&attacker, &defender, 0)
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
