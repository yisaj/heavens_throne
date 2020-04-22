package simulation

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yisaj/heavens_throne/database"
	"github.com/yisaj/heavens_throne/entities"
	"github.com/yisaj/heavens_throne/twitspeak"
)

// StoryTeller contains the logic to generate combat/battle reports and send them
// to the player
type StoryTeller interface {
	Tell() error
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

func (c *canary) Tell() error {
	day, err := c.resource.GetDay(context.TODO())
	if err != nil {
		return errors.Wrap(err, "failed telling story")
	}
	// generate and send DMs to players

	// generate and post the map
	err = c.generateMapSVG()
	if err != nil {
		return errors.Wrap(err, "failed telling story")
	}

	err = c.rasterizeMapSVG()
	if err != nil {
		return errors.Wrap(err, "failed telling story")
	}

	imageID, err := c.speaker.UploadPNG("map.png")
	if err != nil {
		return errors.Wrap(err, "failed telling story")
	}

	mapCaption, err := c.generateMapCaption(day)
	if err != nil {
		return errors.Wrap(err, "failed telling story")
	}
	tweetID, err := c.speaker.Tweet(mapCaption, "", imageID)
	if err != nil {
		return errors.Wrap(err, "failed telling story")
	}

	// generate and post battle reports
	battleLocations, err := c.resource.GetBattleLocations(context.TODO())

	for _, battleLocation := range battleLocations {
		// players before combat
		// class breakdown

		// players after combat
		// class breakdown

		// results
		// captured
		// occupied
		// stalemate
	}
	return nil
}

func (c *canary) generateMapCaption(day int32) (string, error) {
	// TODO ENGINEER: the map tweet should also give the day, victories, temple captures, highlights, etc

	// write the day header

	// write the story

	return "DAY " + strconv.Itoa(int(day)), nil
}

func (c *canary) SendNoReports(players []entities.Player) error {
	for _, player := range players {
		err := c.speaker.SendDM(player.TwitterID, generateNoReport(&player))
		if err != nil {
			return errors.Wrap(err, "failed to send no report")
		}
	}
	return nil
}

func (c *canary) SendCombatReports(combatEvents []entities.CombatEvent) error {
	for _, combatEvent := range combatEvents {
		err := c.speaker.SendDM(combatEvent.Attacker.TwitterID, generateCombatReport(&combatEvent))
		if err != nil {
			return errors.Wrap(err, "failed to send combat report")
		}
	}
	return nil
}

func generateNoReport(player *entities.Player) string {
	return "No fight"
}

func generateCombatReport(combatEvent *entities.CombatEvent) string {
	combatMsg := `
Your %s was %s.	
`
	var typeStr string
	var resultStr string

	switch combatEvent.EventType {
	case entities.Attack:
		typeStr = "Attack"
	case entities.CounterAttack:
		typeStr = "Counter Attack"
	case entities.Revive:
		typeStr = "Revive"
	}

	switch combatEvent.Result {
	case entities.Success:
		resultStr = "Successful"
	case entities.Failure:
		resultStr = "Unsuccessful"
	}

	return fmt.Sprintf(combatMsg, typeStr, resultStr)
}

func generateLocationReport(locationEvent *LocationEvent) string {
	battleMsg := `
location: %s, survivors: %d, fatalities: %d	
`

	return fmt.Sprintf(battleMsg, locationEvent.locationAfter.Name, len(locationEvent.survivors), len(locationEvent.fatalities))
}

func (c *canary) rasterizeMapSVG() error {
	command := exec.Command("inkscape", "map.svg", "-e", "map.png", "-w", "2000", "-h", "1930")
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

	mapFile, err := os.Create("map.svg")
	if err != nil {
		return errors.Wrap(err, "failed opening map output file")
	}
	defer mapFile.Close()

	reader := bufio.NewReader(templateFile)
	for line, _, err := reader.ReadLine(); err == nil; line, _, err = reader.ReadLine() {
		if len(line) > 0 && line[0] == '*' {
			logrus.WithFields(logrus.Fields{}).Debug(line)
			locationID, err := strconv.ParseInt(string(line[5:7]), 16, 32)

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
			} else {
				mapFile.WriteString("dotgray)")
			}

			mapFile.WriteString(string(line[7:]))

		} else {
			mapFile.WriteString(string(line))
		}
	}
	if err != io.EOF {
		return errors.Wrap(err, "failed reading maptemplate.svg")
	}

	return nil
}
