package simulation

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strconv"

	"github.com/pkg/errors"
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
	speaker  twitspeak.TwitterSpeaker
	resource database.Resource
}

// NewStoryTeller constructs a new storyteller
func NewStoryTeller(speaker twitspeak.TwitterSpeaker, resource database.Resource) StoryTeller {
	return &canary{
		speaker,
		resource,
	}
}

func (c *canary) SendNoFightUpdate(players []*entities.Player) error {
	return nil
}

func (c *canary) SendCombatUpdates(combatEvents []*CombatEvent) error {
	return nil
}

func (c *canary) PostMainThread(battleEvents []*BattleEvent) error {
	return nil
}

func (c *canary) rasterizeMapSVG() error {
	command := exec.Command("inkscape", "map.svg", "-e", "map.png")
	err := command.Run()
	if err != nil {
		return errors.Wrap(err, "failed running inkscape command")
	}
	return nil
}

func (c *canary) generateMapSVG() error {
	templateFile, err := os.Open("maptemplate.svg")
	if err != nil {
		return errors.Wrap(err, "failed opening map template file")
	}
	defer templateFile.Close()

	mapFile, err := os.OpenFile("map.svg", os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return errors.Wrap(err, "failed opening map head file")
	}
	defer mapFile.Close()

	scanner := bufio.NewScanner(templateFile)
	for scanner.Scan() {
		line := scanner.Text()
		if line[0] == '*' {
			locationID, err := strconv.ParseInt(line[5:7], 16, 32)
			if err != nil {
				return errors.Wrap(err, "failed converting location id")
			}

			location, err := c.resource.GetLocation(context.TODO(), int32(locationID))
			if err != nil {
				return errors.Wrap(err, "failed getting location for map generation")
			}

			mapFile.WriteString("url(#")

			if location.Owner.Valid {
				// write owner color
				switch location.Owner.String {
				case "Staghorn Sect":
					mapFile.WriteString("orange")
				case "Order Gorgona":
					mapFile.WriteString("purple")
				case "The Baaturate":
					mapFile.WriteString("green")
				}
			} else {
				// write gray
				mapFile.WriteString("gray")
			}

			if location.Occupier.Valid {
				// write dot color
				switch location.Occupier.String {
				case "Staghorn Sect":
					mapFile.WriteString("dotorange)")
				case "Order Gorgona":
					mapFile.WriteString("dotpurple)")
				case "The Baaturate":
					mapFile.WriteString("dotgreen)")
				}
			}

			mapFile.WriteString(line[7:])

		} else {
			mapFile.WriteString(line)
		}
	}

	return nil
}
