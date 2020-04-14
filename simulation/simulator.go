package simulation

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/yisaj/heavens_throne/database"
	"github.com/yisaj/heavens_throne/entities"

	"github.com/bsm/bst"
	"github.com/pkg/errors"
)

const (
	speedStdDev       float64 = 10
	attackStdDev      float64 = 40
	spearAttackBonus  int     = 10
	spearDefenseBonus int     = 10
	experienceStdDev  float64 = 5
	killExperience    float64 = 30
	deathExperience   float64 = 50
	battleExperience  float64 = 20
)

// SimLock provides mutual exclusion in the database between the simulator and
// twitlisten player input
type SimLock struct {
	held     bool
	holdLock sync.Mutex
	longLock sync.RWMutex
}

// WLock gets the write lock for when the simulator starts running
func (sl *SimLock) WLock() {
	sl.holdLock.Lock()
	sl.held = true
	sl.longLock.Lock()
	sl.holdLock.Unlock()
}

// WUnlock releases the write lock for when the simulator finishes running
func (sl *SimLock) WUnlock() {
	sl.holdLock.Lock()
	sl.held = false
	sl.longLock.Unlock()
	sl.holdLock.Unlock()
}

// Check gets a read lock if the simulator is not running, returning true
// otherwise. For twitlisten player input affecting the database
func (sl *SimLock) Check() bool {
	sl.holdLock.Lock()
	if !sl.held {
		sl.longLock.RLock()
	}
	sl.holdLock.Unlock()
	return sl.held
}

// RUnlock releases a read lock for when a twitlisten player input finishes with
// the database
func (sl *SimLock) RUnlock() {
	sl.longLock.RUnlock()
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

// All the combat results
const (
	Success CombatResult = iota
	Failure
	NoTarget
)

// CombatEvent details what happened in a particular instance of combat
type CombatEvent struct {
	Attacker  *entities.Player
	Defender  *entities.Player
	EventType CombatEventType
	Result    CombatResult
}

// Simulator is the base interface for all game simulators
type Simulator interface {
	Simulate() error
}

// NormalSimulator is the first, most natural implementation of a simulator
type NormalSimulator struct {
	logger      *logrus.Logger
	resource    database.Resource
	storyteller StoryTeller
	lock        *SimLock
}

// NewNormalSimulator constructs a NormalSimulator
func NewNormalSimulator(logger *logrus.Logger, resource database.Resource, storyteller StoryTeller, lock *SimLock) NormalSimulator {
	return NormalSimulator{
		logger,
		resource,
		storyteller,
		lock,
	}
}

type LocationEventType int

const (
	Battle LocationEventType = iota
	NoContest
)

// LocationEvent denotes what happened during an entire battle at a location
type LocationEvent struct {
	eventType      LocationEventType
	survivors      map[string][]*entities.Player
	fatalities     map[string][]*entities.Player
	locationBefore entities.Location
	locationAfter  entities.Location
}

// Simulate simulates a day and makes the appropriate changes to the database
func (ns *NormalSimulator) Simulate() error {
	ns.lock.WLock()

	// increment the day and move all players
	err := ns.resource.IncrementDay(context.TODO())
	if err != nil {
		return errors.Wrap(err, "failed simulation")
	}

	// get all players
	players, err := ns.resource.GetAllPlayers(context.TODO())
	if err != nil {
		return errors.Wrap(err, "failed simulation")
	}

	ns.logger.Debugf("SIMMM %s", players[0].TwitterID)
	ns.logger.Debugf("SIMMM %s", players[1].TwitterID)

	// process all players into a map grouped by location
	playersByLocationAndOrder := make(map[int32]map[string][]entities.Player)
	for _, player := range players {
		ns.logger.Debugf("BETA: %+v", player)
		if playersByLocationAndOrder[player.Location] == nil {
			playersByLocationAndOrder[player.Location] = make(map[string][]entities.Player)
		}
		playersByLocationAndOrder[player.Location][player.MartialOrder] = append(playersByLocationAndOrder[player.Location][player.MartialOrder], player)
		ns.logger.Debugf("DELTA: %+v", playersByLocationAndOrder)
	}

	ns.logger.Debugf("ALPHA %+v", playersByLocationAndOrder)

	locationEvents := make([]LocationEvent, 0, len(playersByLocationAndOrder))
	// for each location simulate a battle
	for locationID, locationPlayers := range playersByLocationAndOrder {
		// Count how many armies are present
		numArmies := 0
		var occupier string
		for order, players := range locationPlayers {
			if len(players) > 0 {
				numArmies++
				occupier = order
			}
		}

		var locationEvent LocationEvent
		if numArmies < 2 {
			// no battle occurs
			err := ns.storyteller.SendNoReports(locationPlayers[occupier])
			if err != nil {
				return errors.Wrap(err, "failed simulation")
			}

			locationEvent = LocationEvent{
				eventType: NoContest,
			}

		} else {
			// battle occurs
			survivors, fatalities, combatEvents, err := ns.SimulateBattle(locationID, locationPlayers)
			if err != nil {
				return errors.Wrap(err, "failed simulation")
			}

			// kill all dead players in the database
			for _, dead := range fatalities {
				for _, fatality := range dead {
					err := ns.resource.KillPlayer(context.TODO(), fatality.TwitterID)
					if err != nil {
						return errors.Wrap(err, "failed simulation")
					}
				}
			}

			// dole out player experience
			for _, event := range combatEvents {
				ns.giveCombatExperience(&event)
			}

			locationEvent = LocationEvent{
				eventType:  Battle,
				survivors:  survivors,
				fatalities: fatalities,
			}

			// send out storyteller individual tweets and save battle tweets
			err = ns.storyteller.SendCombatReports(combatEvents)
			if err != nil {
				return errors.Wrap(err, "failed simulation")
			}

			// calculate the occupier
			var max int
			for order, array := range survivors {
				if len(array) > max {
					occupier = order
					max = len(array)
				}
			}
		}

		// check if ownership of the location has changed
		location, err := ns.resource.GetLocation(context.TODO(), locationID)
		if err != nil {
			return errors.Wrap(err, "failed simulation")
		}

		deltaLocation := *location
		if location.Occupier.String == occupier {
			// location ownership should change
			deltaLocation.Owner.String = occupier
			err = ns.resource.SetLocationOwner(context.TODO(), location.ID, occupier)
			if err != nil {
				return errors.Wrap(err, "failed simulation")
			}
		}
		deltaLocation.Occupier.String = occupier

		// Change the location owner to the new owner
		if location.Owner.String != deltaLocation.Owner.String {
			err = ns.resource.SetLocationOwner(context.TODO(), locationID, occupier)
			if err != nil {
				return errors.Wrap(err, "failed simulation")
			}
		}

		// change the location occupier if necessary
		if location.Occupier.String != deltaLocation.Occupier.String {
			err = ns.resource.SetLocationOccupier(context.TODO(), locationID, occupier)
			if err != nil {
				return errors.Wrap(err, "failed simulation")
			}
		}

		locationEvent.locationBefore = *location
		locationEvent.locationAfter = deltaLocation
		locationEvents = append(locationEvents, locationEvent)
	}

	err = ns.resource.RevivePlayers(context.TODO())
	if err != nil {
		return errors.Wrap(err, "failed simulation")
	}

	// tweet new map state and battle tweets
	err = ns.storyteller.PostMainThread(locationEvents)
	if err != nil {
		return errors.Wrap(err, "failed simulation")
	}

	// TODO ENGINEER: check if game is over

	ns.lock.WUnlock()
	return nil
}

func (ns *NormalSimulator) giveCombatExperience(event *CombatEvent) {
	// TODO DESIGN: do this
	/*
		if event.Result == Success {
			if event.Attacker.ID == playerID {
				event.Attacker.Experience += int16(rand.NormFloat64()*experienceStdDev + killExperience)
			} else {
				event.Attacker.Experience += int16(rand.NormFloat64()*experienceStdDev + deathExperience)
			}
		} else {
			event.Attacker.Experience += int16(rand.NormFloat64()*experienceStdDev + battleExperience)
		}
	*/
}

// SimulateBattle simulates a battle at a single location
func (ns *NormalSimulator) SimulateBattle(location int32, players map[string][]entities.Player) (map[string][]*entities.Player, map[string][]*entities.Player, []CombatEvent, error) {
	deadPlayers := map[string]*bst.Map{
		"Staghorn Sect": bst.NewMap(10),
		"Order Gorgona": bst.NewMap(10),
		"The Baaturate": bst.NewMap(10),
	}

	// calculate attack order
	livingPlayers := ns.calculateAttackOrder(players)
	combatEvents := make([]CombatEvent, 0, livingPlayers.Len())

	// calculate total aggros
	totalAggros, medicPowers := ns.calculateTotalAggros(livingPlayers)

	// for each player take an action
	for iter := livingPlayers.Iterator(); iter.Next(); {
		player := iter.Value().(*entities.Player)
		playerStats := player.GetStats()
		playerInitiative := iter.Key().(bst.Float64)

		if player.Class == "healer" {
			// try to revive an ally
			reviveEvent := ns.reviveTarget(player, deadPlayers, livingPlayers)
			combatEvents = append(combatEvents, reviveEvent)
		}

		// attack a random enemy
		numAttacks := 1
		if player.Class == "mage" {
			numAttacks = 3
		}
		for i := 0; i < numAttacks; i++ {
			// calculate total enemy aggro
			totalEnemyAggro := ns.calculateEnemyAggro(player, totalAggros)
			ns.logger.Debugf("totalEnemyAggro: %d", totalEnemyAggro)
			// select target
			target, targetInitiative := ns.selectTarget(player, livingPlayers, totalEnemyAggro)
			targetStats := target.GetStats()

			if target == nil {
				attackEvent := CombatEvent{nil, player, Attack, NoTarget}
				combatEvents = append(combatEvents, attackEvent)
				continue
			}

			// decide what to do
			attackEvent := ns.attackTarget(player, target, medicPowers[target.MartialOrder])
			fmt.Printf("HRY %+v\n", attackEvent)
			combatEvents = append(combatEvents, attackEvent)
			if attackEvent.Result == Success {
				// move target to graveyard
				deadPlayers[target.MartialOrder].Add(bst.Float64(targetInitiative), target)
				livingPlayers.Delete(bst.Float64(targetInitiative))

				// make sure aggro and medic counts are correct
				if target.Class != "monsterknight" {
					totalAggros["standard"][target.MartialOrder] -= targetStats.Aggro
					totalAggros["horsearcher"][target.MartialOrder]--
				}
				totalAggros["ranged"][target.MartialOrder] -= targetStats.Aggro
				if target.Class == "medic" || target.Class == "healer" {
					medicPowers[target.MartialOrder] -= targetStats.Potency
				}

			} else {
				if target.Class == "glaivemaster" {
					counterAttackEvent := ns.counterAttackTarget(target, player, medicPowers[player.MartialOrder])
					combatEvents = append(combatEvents, counterAttackEvent)
					if counterAttackEvent.Result == Success {
						// move player to graveyard
						deadPlayers[player.MartialOrder].Add(bst.Float64(playerInitiative), player)
						livingPlayers.Delete(bst.Float64(playerInitiative))

						// make sure aggro and medic counts are correct
						if player.Class != "monsterknight" {
							totalAggros["standard"][player.MartialOrder] -= playerStats.Aggro
							totalAggros["horsearcher"][player.MartialOrder]--
						}
						totalAggros["ranged"][player.MartialOrder] -= playerStats.Aggro
						if player.Class == "medic" || player.Class == "healer" {
							medicPowers[player.MartialOrder] -= playerStats.Potency
						}
					}
				}
			}
		}
	}

	// serialize the dead players
	fatalities := make(map[string][]*entities.Player)
	for order, dead := range deadPlayers {
		fatalities[order] = make([]*entities.Player, 0, dead.Len())
		for iter := dead.Iterator(); iter.Next(); {
			player := iter.Value().(*entities.Player)
			ns.logger.Debugf("DEADER: %+v", *player)
			fatalities[order] = append(fatalities[order], player)
		}
	}

	// serialize the living players
	survivors := make(map[string][]*entities.Player)
	for iter := livingPlayers.Iterator(); iter.Next(); {
		player := iter.Value().(*entities.Player)
		survivors[player.MartialOrder] = append(survivors[player.MartialOrder], player)
	}

	return survivors, fatalities, combatEvents, nil
}

func (ns *NormalSimulator) selectTarget(player *entities.Player, livingPlayers *bst.Map, totalEnemyAggro int) (*entities.Player, bst.Float64) {
	aggroLeft := rand.Intn(totalEnemyAggro)
	var target *entities.Player
	var targetInitiative bst.Float64
	for iter := livingPlayers.Iterator(); iter.Next(); {
		target = iter.Value().(*entities.Player)
		targetInitiative = iter.Key().(bst.Float64)
		if target.MartialOrder == player.MartialOrder || (target.Class == "monsterknight" && !player.IsRanged()) {
			continue
		}

		if player.Class == "horsearcher" {
			aggroLeft--
		} else {
			aggroLeft -= target.GetStats().Aggro
		}
		if aggroLeft <= 0 {
			break
		}
	}
	return target, targetInitiative
}

func (ns *NormalSimulator) calculateTotalAggros(attackOrder *bst.Map) (map[string]map[string]int, map[string]int) {
	totalAggros := map[string]map[string]int{
		"standard": {
			"Staghorn Sect": 0,
			"Order Gorgona": 0,
			"The Baaturate": 0,
		},
		"ranged": {
			"Staghorn Sect": 0,
			"Order Gorgona": 0,
			"The Baaturate": 0,
		},
		"horsearcher": {
			"Staghorn Sect": 0,
			"Order Gorgona": 0,
			"The Baaturate": 0,
		},
	}
	medicPowers := map[string]int{
		"Staghorn Sect": 0,
		"Order Gorgona": 0,
		"The Baaturate": 0,
	}

	for iter := attackOrder.Iterator(); iter.Next(); {
		player := iter.Value().(*entities.Player)
		stats := player.GetStats()
		ns.logger.Debugf("PLAYER: %s %s", player.MartialOrder, player.TwitterID)

		// calculate total aggros
		if player.Class != "monsterknight" {
			totalAggros["standard"][player.MartialOrder] += stats.Aggro
			totalAggros["horsearcher"][player.MartialOrder]++
		}
		totalAggros["ranged"][player.MartialOrder] += stats.Aggro

		// calculate medic totals
		if player.Class == "medic" || player.Class == "healer" {
			medicPowers[player.MartialOrder] += stats.Potency
		}
	}
	ns.logger.Debugf("aggros: %d %d %d", totalAggros["standard"]["The Baaturate"], totalAggros["standard"]["Staghorn Sect"], totalAggros["standard"]["Order Gorgona"])
	return totalAggros, medicPowers
}

func (ns *NormalSimulator) calculateEnemyAggro(player *entities.Player, totalAggros map[string]map[string]int) int {
	enemyAggro := 0
	if player.IsRanged() {
		for order, aggro := range totalAggros["ranged"] {
			if order != player.MartialOrder {
				enemyAggro += aggro
			}
		}
		return enemyAggro
	} else if player.Class == "horsearcher" {
		for order, aggro := range totalAggros["horsearcher"] {
			if order != player.MartialOrder {
				enemyAggro += aggro
			}
		}
		return enemyAggro
	}
	for order, aggro := range totalAggros["standard"] {
		if order != player.MartialOrder {
			enemyAggro += aggro
		}
	}
	return enemyAggro
}

// reviveTarget attempts to return an allied player from the dead and back to the battle
func (ns *NormalSimulator) reviveTarget(player *entities.Player, deadPlayers map[string]*bst.Map, attackOrder *bst.Map) CombatEvent {
	// try to revive an ally
	// TODO DESIGN: figure revive probability rates
	myDead := deadPlayers[player.MartialOrder]
	iter := myDead.Iterator()
	iter.Next()
	if myDead.Len() > 0 {
		target := iter.Value().(*entities.Player)
		initiative := iter.Key()
		if rand.Intn(2) > 0 {
			attackOrder.Add(initiative, target)
			myDead.Delete(initiative)
			return CombatEvent{target, player, Revive, Success}
		}
		return CombatEvent{target, player, Revive, Failure}
	}
	return CombatEvent{nil, player, Revive, NoTarget}
}

func (ns *NormalSimulator) calculateAttackOrder(players map[string][]entities.Player) *bst.Map {
	attackOrder := bst.NewMap(len(players))
	for order := range players {
		for _, player := range players[order] {
			playerStats := player.GetStats()
			ns.logger.Debugf("PLAYER CALC: %s", player.TwitterID)
			for {
				// initiative is negated, since the map sorts in ascending order
				initiative := bst.Float64(-(rand.NormFloat64()*speedStdDev + float64(playerStats.Speed)))
				if !attackOrder.Exists(initiative) {
					attackOrder.Add(initiative, &player)
					break
				}
			}
		}
	}
	return attackOrder
}

func (ns *NormalSimulator) counterAttackTarget(attacker *entities.Player, defender *entities.Player, medicBonus int) CombatEvent {
	event := ns.attackTarget(attacker, defender, medicBonus)
	event.EventType = CounterAttack
	return event
}

func (ns *NormalSimulator) attackTarget(attacker *entities.Player, defender *entities.Player, medicBonus int) CombatEvent {
	attackerStats := attacker.GetStats()
	defenderStats := defender.GetStats()

	attackerIsCavalry := attacker.Class == "cavalry" || attacker.Class == "lightcavalry" ||
		attacker.Class == "heavycavalry" || attacker.Class == "monsterknight" || attacker.Class == "horsearcher"
	attackerIsSpear := attacker.Class == "spear" || attacker.Class == "glaivemaster"

	defenderIsCavalry := defender.Class == "cavalry" || defender.Class == "lightcavalry" ||
		defender.Class == "heavycavalry" || defender.Class == "monsterknight" || defender.Class == "horsearcher"
	defenderIsSpear := defender.Class == "spear" || defender.Class == "glaivemaster"

	// calculate outcome
	attackPower := attackerStats.Potency
	defensePower := defenderStats.Defense

	// medic bonus
	// TODO DESIGN: determine scaling of medic bonus
	defense := float64(defensePower) + float64(medicBonus)/1000.*10
	// spear bonus
	if attackerIsCavalry && defenderIsSpear {
		defensePower += spearDefenseBonus
	} else if attackerIsSpear && defenderIsCavalry {
		attackPower += spearAttackBonus
	}
	// TODO DESIGN: implement defender's bonus

	attack := rand.NormFloat64()*attackStdDev + float64(attackPower)

	fmt.Printf("%f, %f", attack, defense)
	if attack > defense {
		return CombatEvent{attacker, defender, Attack, Success}
	}
	return CombatEvent{attacker, defender, Attack, Failure}
}
