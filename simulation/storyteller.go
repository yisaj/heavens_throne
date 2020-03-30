package simulation

import (
	"github.com/yisaj/heavens_throne/database"
	"github.com/yisaj/heavens_throne/twitspeak"
)

type StoryTeller interface {
	SendBattleUpdates(battleEvent *BattleEvent, combatEvents []*CombatEvent) error
	PostMainThread() error
}

type canary struct {
	speaker      twitspeak.TwitterSpeaker
	resource     database.Resource
	battleTweets []string
}

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
