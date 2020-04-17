package input

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/yisaj/heavens_throne/database"
	"github.com/yisaj/heavens_throne/simulation"
	"github.com/yisaj/heavens_throne/twitspeak"

	"github.com/pkg/errors"
)

var (
	locationIDs = map[string]int32{
		"heavensthrone": 0, "heavens": 0, "heaven": 0, "throne": 0, "throneofheaven": 0, "power": 0, "madness": 0,
		"sulferpoint": 1, "sulfer": 1, "point": 1,
		"herebe": 2, "here": 2, "be": 2,
		"nowhere":        3,
		"stcecilsbridge": 4, "stcecils": 4, "cecilsbridge": 4, "st": 4, "cecils": 4, "bridge": 4,
		"saintcecilsbridge": 4, "saintcecils": 4, "saint": 4, "stcecil": 4, "saintcecil": 4, "cecil": 4,
		"eyeofgideon": 5, "eye": 5, "gideon": 5, "gideonseye": 5, "eyeof": 5, "ofgideon": 5, "gideons": 5,
		"newdelphia": 6, "new": 6, "delphia": 6,
		"fog":      7,
		"wormland": 8, "worm": 8,
		"passageofsmoke": 9, "passage": 9, "passageof": 9, "ofsmoke": 9, "smoke": 9, "smokepassage": 9,
		"theashsea": 10, "ashsea": 10, "ash": 10, "sea": 10,
		"asteria":      11,
		"york":         12,
		"hideousmarsh": 13, "hideous": 13, "marsh": 13,
		"necropolis": 14, "necro": 14, "polis": 14,
		"crawlerpits": 15, "crawler": 15, "pits": 15,
		"obsidianlake": 16, "obsidian": 16, "lake": 16,
		"grisag":    17,
		"fucoterre": 18, "fuco": 18, "terre": 18,
		"campgray": 19, "gray": 19,
		"campwatkins": 20, "watkins": 20,
		"gallows":   21,
		"mercycove": 22, "mercy": 22, "cove": 22,
		"giantsbluff": 23, "giants": 23, "giant": 23, "bluff": 23,
		"hbeach": 24, "hollowbeach": 24, "hollow": 24, "beach": 24,
		"duncantalley": 25, "duncan": 25, "talley": 25,
		"mangrove":   26,
		"lighthouse": 27, "light": 27, "house": 27,
		"apostlevalley": 28, "apostle": 28, "valley": 28,
		"poppyfields": 29, "poppy": 29, "fields": 29, "field": 29,
		"agathinias": 30,
		"ithmont":    31,
		"whitecrypt": 32, "white": 32, "crypt": 32,
		"visygi":     33,
		"outerrealm": 34, "outer": 34, "realm": 34,
		"memoria": 35,
		"hemwood": 36, "hem": 36, "wood": 36,
		"rivercrossing": 37, "river": 37, "crossing": 37,
		"foolsway": 38, "fools": 38, "fool": 38, "way": 38,
		"landfall": 39, "fall": 39,
		"bouchardsisland": 40, "bouchards": 40, "bouchard": 40, "island": 40,
	}
	maxClassRanks = map[string]int16{
		"recruit":  1,
		"infantry": 3, "cavalry": 3, "ranger": 3,
		"spear": 5, "sword": 5, "heavycavalry": 5, "lightcavalry": 5, "archer": 5, "medic": 5,
		"glaivemaster": 1, "legionary": 1, "monsterknight": 1, "horsearcher": 1, "mage": 1, "healer": 1,
	}
	classAdvances = map[string][]string{
		"recruit":  {"infantry", "cavalry", "ranger"},
		"infantry": {"spear", "sword"}, "cavalry": {"heavycavalry", "lightcavalry"}, "ranger": {"archer", "medic"},
		"spear": {"glaivemaster"}, "sword": {"legionary"},
		"heavycavalry": {"monsterknight"}, "lightcavalry": {"horsearcher"},
		"archer": {"mage"}, "medic": {"healer"},
		"glaivemaster": {}, "legionary": {}, "monsterknight": {}, "horsearcher": {}, "mage": {}, "healer": {},
	}
	classDescriptions = map[string]string{
		"infantry":      "INFANTRY: ",
		"cavalry":       "CAVALRY: ",
		"ranger":        "RANGER: ",
		"spear":         "SPEAR: ",
		"sword":         "SWORD ",
		"heavycavalry":  "HEAVY CAVALRY: ",
		"lightcavalry":  "LIGHT CAVALRY ",
		"archer":        "ARCHER: ",
		"medic":         "MEDIC: ",
		"glaivemaster":  "GLAIVEMASTER: ",
		"legionary":     "LEGIONARY: ",
		"monsterknight": "MONSTER KNIGHT: ",
		"horsearcher":   "COURSER: ",
		"mage":          "MAGE: ",
		"healer":        "HEALER: ",
	}
)

// Handler contains methods to handle each of the possible player inputs
type Handler interface {
	Help(ctx context.Context, recipientID string) error
	Status(ctx context.Context, recipientID string) error
	Logistics(ctx context.Context, recipientID string, locationString string) error
	Join(ctx context.Context, recipientID string, order string) error
	Move(ctx context.Context, recipientID string, location string) error
	Advance(ctx context.Context, recipientID string, class string) error
	Quit(ctx context.Context, recipientID string) error
	ToggleUpdates(ctx context.Context, recipientID string) error
	InvalidCommand(ctx context.Context, recipientID string) error
	Echo(ctx context.Context, recipientID string, msg string) error
	Simulate(ctx context.Context, recipientID string) error
	Tweet(ctx context.Context, recipientID string, msg string) error
	Reply(ctx context.Context, recipientID string, argument string) error
	ImageTweet(ctx context.Context, recipientID string, filename string) error
}

// A player input handler has to be able to access database resources and respond
// to the player via a twitter speaker
type handler struct {
	resource  database.Resource
	speaker   twitspeak.TwitterSpeaker
	simulator simulation.Simulator
}

// newInputHandler constructs a handler to handle player input
func newInputHandler(resource database.Resource, speaker twitspeak.TwitterSpeaker, simulator simulation.Simulator) Handler {
	return &handler{
		resource,
		speaker,
		simulator,
	}
}

// Help sends the player a list of commands and info about the game
func (h *handler) Help(ctx context.Context, recipientID string) error {
	const newPlayerHelp = `
	Type !join [order] to join.
`
	const activePlayerHelp = `
	!status to see your status.
`

	player, err := h.resource.GetPlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed parsing DM")
	}

	if player == nil {
		err = h.speaker.SendDM(recipientID, newPlayerHelp)
	} else {
		err = h.speaker.SendDM(recipientID, activePlayerHelp)
	}

	if err != nil {
		return errors.Wrap(err, "failed sending help message")
	}
	return nil
}

func (h *handler) Status(ctx context.Context, recipientID string) error {
	// TODO WRITE: write a real status message
	const statusFormat = `
Order: %s
Class: %s
Experience: %d
Location: %s
Next Location: %s
`
	const advanceFormat = `
You have an !advance available
`

	player, err := h.resource.GetPlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed parsing DM")
	}
	if player == nil {
		return nil
	}

	// TODO ENGINEER: include status for dead players
	if player.IsAlive() {
		location, err := h.resource.GetLocation(ctx, player.Location.Int32)
		if err != nil || location == nil {
			return errors.Wrap(err, "failed sending player status")
		}
		nextLocation, err := h.resource.GetLocation(ctx, player.NextLocation.Int32)
		if err != nil || nextLocation == nil {
			return errors.Wrap(err, "failed sending player status")
		}

		msg := fmt.Sprintf(statusFormat, player.MartialOrder, player.FormatClass(), player.Experience, location.Name, nextLocation.Name)
		if player.Experience >= 100 {
			msg += fmt.Sprintf(advanceFormat)
		}

		err = h.speaker.SendDM(recipientID, msg)
		if err != nil {
			return errors.Wrap(err, "failed sending help message")
		}
	}

	return nil
}

// Logistics sends the player information about where allied units are and where
// they're going
// TODO ENGINEER: this shit is broke yo. displayed logistics are reversed, seemingly?
func (h *handler) Logistics(ctx context.Context, recipientID string, locationString string) error {
	const notFound = `
That's not a place that I know of.
`
	const allHeader = `
Here's all the logistics
----------------------------
`
	const arrivingHeader = `
These are the arriving logistics

`
	const leavingHeader = `
These are the leaving logistics

`

	player, err := h.resource.GetPlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed parsing DM")
	}
	if player == nil {
		return nil
	}

	if locationString == "" {
		currentLogistics, err := h.resource.GetCurrentLogistics(ctx, player.MartialOrder)
		if err != nil {
			return errors.Wrap(err, "failed getting logistics")
		}
		nextLogistics, err := h.resource.GetNextLogistics(ctx, player.MartialOrder)
		if err != nil {
			return errors.Wrap(err, "failed getting logistics")
		}

		sort.Slice(currentLogistics, func(i int, j int) bool {
			return currentLogistics[i].LocationName < currentLogistics[j].LocationName
		})
		sort.Slice(nextLogistics, func(i int, j int) bool {
			return nextLogistics[i].LocationName < nextLogistics[j].LocationName
		})

		var msg strings.Builder
		msg.WriteString(allHeader)
		i, j := 0, 0
		for i < len(currentLogistics) && j < len(nextLogistics) {
			current, next := currentLogistics[i], nextLogistics[j]
			if current.LocationName < next.LocationName {
				msg.WriteString(fmt.Sprintf("%s:  0 -> %d (%+d)\n", current.LocationName, current.Count, current.Count))
				i++
			} else if current.LocationName > next.LocationName {
				msg.WriteString(fmt.Sprintf("%s: %d -> 0 (%+d)\n", next.LocationName, next.Count, -next.Count))
				j++
			} else {
				msg.WriteString(fmt.Sprintf("%s: %d -> %d (%+d)\n", current.LocationName, current.Count, next.Count, next.Count-current.Count))
				i++
				j++
			}
		}
		for i < len(currentLogistics) {
			current := currentLogistics[i]
			msg.WriteString(fmt.Sprintf("%s: 0 -> %d (%+d)\n", current.LocationName, current.Count, current.Count))
			i++
		}
		for j < len(nextLogistics) {
			next := nextLogistics[j]
			msg.WriteString(fmt.Sprintf("%s: %d -> 0 (%+d)\n", next.LocationName, next.Count, -next.Count))
			j++
		}

		err = h.speaker.SendDM(recipientID, msg.String())
		if err != nil {
			return errors.Wrap(err, "failed getting logistics")
		}
	} else {
		reg, err := regexp.Compile("[^a-zA-Z0-9]+")
		if err != nil {
			return errors.Wrap(err, "failed building regexp")
		}
		locationString = reg.ReplaceAllString(locationString, "")

		id, err := strconv.Atoi(locationString)
		locationID := int32(id)
		if err != nil {
			id, ok := locationIDs[locationString]
			if ok {
				locationID = id
			} else {
				err = h.speaker.SendDM(recipientID, notFound)
				if err != nil {
					return errors.Wrap(err, "failed sending location not found message")
				}
				return nil
			}
		}

		arrivingLogistics, err := h.resource.GetArrivingLogistics(ctx, locationID)
		if err != nil {
			return errors.Wrap(err, "failed getting location logistics")
		}
		leavingLogistics, err := h.resource.GetLeavingLogistics(ctx, locationID)
		if err != nil {
			return errors.Wrap(err, "failed getting location logistics")
		}

		var msg strings.Builder
		msg.WriteString(arrivingHeader)
		for _, logistic := range arrivingLogistics {
			msg.WriteString(fmt.Sprintf("%s (+%d)\n", logistic.LocationName, logistic.Count))
		}
		msg.WriteString(leavingHeader)
		for _, logistic := range leavingLogistics {
			msg.WriteString(fmt.Sprintf("%s (-%d)\n", logistic.LocationName, logistic.Count))
		}

		err = h.speaker.SendDM(recipientID, msg.String())
		if err != nil {
			return errors.Wrap(err, "failed getting location logistics")
		}
	}

	return nil
}

// Join adds a new player to the game under the chosen order
func (h *handler) Join(ctx context.Context, recipientID string, order string) error {
	// TODO WRITE: write a real join message
	const joinFormat = `
ORDER: %s
CLASS: %s
LOCATION: %d
`
	const invalidOrder = `
Invalid order. Please select from 'staghorn', 'gorgona', or 'baaturate'.
`
	const alreadyPlaying = `
You're already playing.
`
	const deactivatedPlayer = `
The Gate is closed to you. At least for this cycle.
`

	player, err := h.resource.GetPlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed parsing DM")
	}

	if player != nil {
		if player.Active {
			err = h.speaker.SendDM(recipientID, alreadyPlaying)
			if err != nil {
				return errors.Wrap(err, "failed to send already playing message")
			}
		} else {
			err = h.speaker.SendDM(recipientID, deactivatedPlayer)
			if err != nil {
				return errors.Wrap(err, "failed to send deactivated player message")
			}
		}
		return nil
	}

	var orderName string
	if strings.Contains(order, "staghorn") {
		orderName = "Staghorn Sect"
	} else if strings.Contains(order, "gorgona") {
		orderName = "Order Gorgona"
	} else if strings.Contains(order, "baaturate") {
		orderName = "The Baaturate"
	} else {
		err := h.speaker.SendDM(recipientID, invalidOrder)
		if err != nil {
			return errors.Wrap(err, "failed to send invalid order message")
		}
		return nil
	}

	locationID, err := h.resource.GetTempleLocation(ctx, orderName)
	if err != nil {
		return errors.Wrap(err, "failed getting starting location")
	}

	player, err = h.resource.CreatePlayer(ctx, recipientID, orderName, locationID)
	if err != nil {
		return errors.Wrap(err, "failed joining new player")
	}

	err = h.speaker.SendDM(recipientID, fmt.Sprintf(joinFormat, player.MartialOrder, player.FormatClass(), player.Location.Int32))
	if err != nil {
		return errors.Wrap(err, "failed to send join message")
	}

	return nil
}

// Move tries to set the player's next location to the given location
func (h *handler) Move(ctx context.Context, recipientID string, locationString string) error {
	const notFound = `
That's not a place that I know of.
`
	const notAdjacent = `
That's not an adjacent location.'
`
	const moving = `
You are now moving to %s.
`
	const dead = `
You are too dead to move anywhere.	
`
	player, err := h.resource.GetPlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed parsing DM")
	}
	if player == nil {
		return nil
	}

	if !player.IsAlive() {
		err = h.speaker.SendDM(recipientID, dead)
		if err != nil {
			return errors.Wrap(err, "failed sending player move on dead message")
		}
		return nil
	}

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return errors.Wrap(err, "failed building regexp")
	}
	locationString = reg.ReplaceAllString(locationString, "")

	id, err := strconv.Atoi(locationString)
	locationID := int32(id)
	if err != nil {
		id, ok := locationIDs[locationString]
		if ok {
			locationID = id
		} else {
			err = h.speaker.SendDM(recipientID, notFound)
			if err != nil {
				return errors.Wrap(err, "failed sending location not found message")
			}
			return nil
		}
	}

	if locationID != player.Location.Int32 {
		adjacentLocations, err := h.resource.GetAdjacentLocations(ctx, player.Location.Int32)
		if err != nil {
			return errors.Wrap(err, "failed moving player")
		}

		// TODO ENGINEER: Revert to false after testing
		found := true
		for _, adjacentLocation := range adjacentLocations {
			if adjacentLocation == locationID {
				found = true
				break
			}
		}
		if !found {
			err = h.speaker.SendDM(recipientID, notAdjacent)
			if err != nil {
				return errors.Wrap(err, "failed sending not adjacent message")
			}
			return nil
		}
	}

	err = h.resource.UpdatePlayerDestination(ctx, recipientID, locationID)
	if err != nil {
		return errors.Wrap(err, "failed moving player")
	}
	location, err := h.resource.GetLocation(ctx, locationID)
	if err != nil {
		return errors.Wrap(err, "failed moving player")
	}
	err = h.speaker.SendDM(recipientID, fmt.Sprintf(moving, location.Name))
	if err != nil {
		return errors.Wrap(err, "failed sending moved player message")
	}

	return nil
}

// Advance attempts to level a player up to another rank or class
// TODO ENGINEER: think about if rank advance with a class name should error or just auto rank advance
func (h *handler) Advance(ctx context.Context, recipientID string, class string) error {
	const notExperienced = `
You don't have enough experience.
`
	const maxClass = `
You've already reached the peak.
`
	const advanceInfoHeader = `
Here are your advances:
`
	const rankAdvance = `
You advanced: %s -> %s.
`
	const classAdvance = `
You advanced: %s -> %s.
`
	const unknownClass = `
That's not a class I'm aware of.
`

	player, err := h.resource.GetPlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed parsing DM")
	}
	if player == nil {
		return nil
	}

	advances := classAdvances[player.Class]
	if len(advances) == 0 {
		err := h.speaker.SendDM(recipientID, maxClass)
		if err != nil {
			return errors.Wrap(err, "failed getting advance info")
		}
		return nil
	}

	if class == "" {
		// not enough experience
		if player.Experience < 100 {
			err := h.speaker.SendDM(recipientID, notExperienced)
			if err != nil {
				return errors.Wrap(err, "failed getting advance info")
			}
			return nil
		}
		// TODO ENGINEER: move can advance, next advance, max rank, etc logic to player object
		if player.Rank < maxClassRanks[player.Class] {
			// advance a rank
			oldRank := player.FormatClass()
			player.Rank++

			err := h.resource.AdvancePlayer(ctx, recipientID, player.Class, player.Rank)
			if err != nil {
				return errors.Wrap(err, "failed advancing player rank")
			}

			newRank := player.FormatClass()

			err = h.speaker.SendDM(recipientID, fmt.Sprintf(rankAdvance, oldRank, newRank))
			if err != nil {
				return errors.Wrap(err, "failed advancing player rank")
			}
		} else {
			// need to advance a class, which the user didn't provide
			var msg strings.Builder
			msg.WriteString(advanceInfoHeader)
			for _, advance := range advances {
				msg.WriteString(classDescriptions[advance])
			}

			err := h.speaker.SendDM(recipientID, msg.String())
			if err != nil {
				return errors.Wrap(err, "failed getting advance info")
			}
		}
	} else {
		reg, err := regexp.Compile("[^a-zA-Z]+")
		if err != nil {
			return errors.Wrap(err, "failed building regexp")
		}
		classString := reg.ReplaceAllString(class, "")

		// look for the class advance by name
		for _, advance := range advances {
			if advance == classString {
				oldClass := player.FormatClass()
				player.Class = advance
				player.Rank = 1

				err = h.resource.AdvancePlayer(ctx, recipientID, player.Class, player.Rank)
				if err != nil {
					return errors.Wrap(err, "failed advancing player class")
				}

				newClass := player.FormatClass()

				err := h.speaker.SendDM(recipientID, fmt.Sprintf(classAdvance, oldClass, newClass))
				if err != nil {
					return errors.Wrap(err, "failed advancing player class")
				}
				return nil
			}
		}

		// unknown advance name
		err = h.speaker.SendDM(recipientID, unknownClass)
		if err != nil {
			return errors.Wrap(err, "failed advancing player class")
		}
	}

	return nil
}

// Quit deactivates a player's account
// TODO DESIGN: remember to deactivate instead of deleting (also think of rejoin logic)
func (h *handler) Quit(ctx context.Context, recipientID string) error {
	quitMsg := `
Heaven's Gate closes behind you.
`

	player, err := h.resource.GetPlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed parsing DM")
	}
	if player == nil {
		return nil
	}

	err = h.resource.DeletePlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed quitting game")
	}

	err = h.speaker.SendDM(recipientID, quitMsg)
	if err != nil {
		return errors.Wrap(err, "failed to send quit message")
	}
	return nil
}

// ToggleUpdates toggles daily combat and battle reports in DMs
func (h *handler) ToggleUpdates(ctx context.Context, recipientID string) error {
	const updatesOff = `
You will no longer receive daily personal battle reports.
`
	const updatesOn = `
You will now receive daily personal battle reports.
`

	player, err := h.resource.GetPlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed parsing DM")
	}
	if player == nil {
		return nil
	}

	receiveUpdates, err := h.resource.TogglePlayerUpdates(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed toggling updates")
	}

	if receiveUpdates {
		err = h.speaker.SendDM(recipientID, updatesOn)
	} else {
		err = h.speaker.SendDM(recipientID, updatesOff)
	}
	if err != nil {
		return errors.Wrap(err, "failed sending toggle updates message")
	}

	return nil
}

// InvalidCommand tells the player that their command wasn't recognized
func (h *handler) InvalidCommand(ctx context.Context, recipientID string) error {
	const invalid = `
That's not something I understand. Try seeking !help.
`

	player, err := h.resource.GetPlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed parsing DM")
	}
	if player == nil {
		return nil
	}

	err = h.speaker.SendDM(recipientID, invalid)
	if err != nil {
		return errors.Wrap(err, "failed sending invalid command message")
	}

	return nil
}

// Echo just sends a message to a player
func (h *handler) Echo(ctx context.Context, recipientID string, msg string) error {
	err := h.speaker.SendDM(recipientID, "Just got the message: "+msg)
	if err != nil {
		return errors.Wrap(err, "failed sending echo message")
	}
	return nil
}

func (h *handler) Simulate(ctx context.Context, recipientID string) error {
	err := h.simulator.Simulate()
	if err != nil {
		return errors.Wrap(err, "failed simulation")
	}

	err = h.speaker.SendDM(recipientID, "Attempting to simulate...")
	if err != nil {
		return errors.Wrap(err, "failed sending echo message")
	}
	return nil
}

func (h *handler) Tweet(ctx context.Context, recipientID string, msg string) error {
	tweetID, err := h.speaker.Tweet(msg, "", "")
	if err != nil {
		return errors.Wrap(err, "failed posting tweet by DM")
	}

	err = h.speaker.SendDM(recipientID, fmt.Sprintf("Sent tweet with ID: %s", tweetID))
	if err != nil {
		return errors.Wrap(err, "failed sending tweet post confirmation")
	}
	return nil
}

func (h *handler) Reply(ctx context.Context, recipientID string, argument string) error {
	args := strings.SplitN(argument, " ", 2)
	if len(args) < 2 {
		err := h.speaker.SendDM(recipientID, fmt.Sprintf("No tweet ID/message was supplied. Got: %s", argument))
		if err != nil {
			return errors.Wrap(err, "failed sending reply error message")
		}
		return nil
	}

	tweetID, err := h.speaker.Tweet(args[1], args[0], "")
	if err != nil {
		return errors.Wrap(err, "failed posting tweet reply")
	}

	err = h.speaker.SendDM(recipientID, fmt.Sprintf("Replied to tweet %s with %s", args[0], tweetID))
	if err != nil {
		return errors.Wrap(err, "failed sending tweet reply confirmation")
	}
	return nil
}

func (h *handler) ImageTweet(ctx context.Context, recipientID string, filename string) error {
	imageID, err := h.speaker.UploadPNG(filename)
	if err != nil {
		return errors.Wrap(err, "failed uploading image for tweet")
	}

	tweetID, err := h.speaker.Tweet("Image", "", imageID)
	if err != nil {
		return errors.Wrap(err, "failed tweeting image tweet")
	}

	err = h.speaker.SendDM(recipientID, fmt.Sprintf("Posted image tweet %s", tweetID))
	if err != nil {
		return errors.Wrap(err, "failed sending image tweet confirmation")
	}
	return nil
}
