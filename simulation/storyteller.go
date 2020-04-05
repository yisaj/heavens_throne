package simulation

import (
	"github.com/yisaj/heavens_throne/database"
	"github.com/yisaj/heavens_throne/entities"
	"github.com/yisaj/heavens_throne/twitspeak"
)

// StoryTeller contains the logic to generate combat/battle reports and send them
// to the player
type StoryTeller interface {
	SendNoFightUpdate(players []*entities.Player) error
	SendCombatUpdates(combatEvents []*CombatEvent) error
	PostMainThread(battleEvents []*BattleEvent) error
}

// A canary needs to be able to generate and send battle reports
type canary struct {
	speaker      twitspeak.TwitterSpeaker
	resource     database.Resource
	battleTweets []string
}

// NewStoryTeller constructs a new storyteller
func NewStoryTeller(speaker twitspeak.TwitterSpeaker, resource database.Resource) StoryTeller {
	return &canary{
		speaker,
		resource,
		make([]string, 0),
	}
}

func (c *canary) SendBattleUpdates(battleEvent *BattleEvent, combatEvents []*CombatEvent) error {
	return nil
}

func (c *canary) PostMainThread() error {
	return nil
}
