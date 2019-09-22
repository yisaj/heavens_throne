package input

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/yisaj/heavens_throne/database"
	"github.com/yisaj/heavens_throne/entities"
	"github.com/yisaj/heavens_throne/twitspeak"

	"github.com/pkg/errors"
)

var (
	locationIDs = map[string]int32{
		"heavensthrone": 0, "heavens": 0, "heaven": 0, "throne": 0, "throneofheaven": 0, "power": 0, "madness": 0,
		"sulferpoint": 1, "sulfer": 1, "point": 1,
		"herebe": 2, "here": 2, "be": 2,
		"noname": 3, "no": 3, "name": 3,
		"stcecilsbridge": 4, "stcecils": 4, "cecilsbridge": 4, "st": 4, "cecils": 4, "bridge": 4,
		"saintcecilsbridge": 4, "saintcecils": 4, "saint": 4, "stcecil": 4, "saintcecil": 4, "cecil": 4,
		"eyeofgideon": 5, "eye": 5, "gideon": 5, "gideonseye": 5, "eyeof": 5, "ofgideon": 5, "gideons": 5,
		"newdelphia": 6, "new": 6, "delphia": 6,
		"fog":      7,
		"wormland": 8, "worm": 8,
		"passageofsmoke": 9, "passage": 9, "passageof": 9, "ofsmoke": 9, "smoke": 9, "smokepassage": 9,
		"theashsea": 10, "ashsea": 10, "ash": 10, "sea": 10,
		"asteria":      11,
		"yerk":         12,
		"hideousmarsh": 13, "hideous": 13, "marsh": 13,
		"necropolis": 14, "necro": 14, "polis": 14,
		"crawlerpits": 15, "crawler": 15, "pits": 15,
		"obsidianlake": 16, "obsidian": 16, "lake": 16,
		"grisag":     17,
		"terreignot": 18, "terre": 18, "ignot": 18,
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
)

type InputHandler interface {
	Help(ctx context.Context, player *entities.Player, recipientID string) error
	Status(ctx context.Context, player *entities.Player, recipientID string) error
	Logistics(ctx context.Context, player *entities.Player, recipientID string) error
	Join(ctx context.Context, player *entities.Player, recipientID string, order string) error
	Move(ctx context.Context, player *entities.Player, recipientID string, location string) error
	Advance(ctx context.Context, player *entities.Player, recipientID string, class string) error
	Quit(ctx context.Context, player *entities.Player, recipientID string) error
	ToggleUpdates(ctx context.Context, player *entities.Player, recipientID string) error
	InvalidCommand(ctx context.Context, player *entities.Player, recipientID string) error
	Echo(ctx context.Context, player *entities.Player, recipientID string, msg string) error
}

type handler struct {
	resource database.Resource
	speaker  twitspeak.TwitterSpeaker
}

func NewInputHandler(resource database.Resource, speaker twitspeak.TwitterSpeaker) InputHandler {
	return &handler{
		resource,
		speaker,
	}
}

func (h *handler) Help(ctx context.Context, player *entities.Player, recipientID string) error {
	const newPlayerHelp = `
	Type !join [order] to join.
`
	const activePlayerHelp = `
	!status to see your status.
`

	var err error
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

func (h *handler) Status(ctx context.Context, player *entities.Player, recipientID string) error {
	// TODO: write a real status message
	// TODO: handle available advances text
	const statusFormat = `
Order: %s
Class: %s
Experience: %d
Location: %s
Next Location: %s
`

	if player != nil {
		location, err := h.resource.GetLocation(ctx, player.Location)
		if err != nil || location == nil {
			return errors.Wrap(err, "failed sending player status")
		}
		nextLocation, err := h.resource.GetLocation(ctx, player.NextLocation)
		if err != nil || nextLocation == nil {
			return errors.Wrap(err, "failed sending player status")
		}

		msg := fmt.Sprintf(statusFormat, player.MartialOrder, player.FormatClass(), player.Experience, location.Name, nextLocation.Name)
		err = h.speaker.SendDM(recipientID, msg)
	}

	return nil
}

func (h *handler) Logistics(ctx context.Context, player *entities.Player, recipientID string) error {
	const logisticsHeader = `
Here's all the logistics
----------------------------
`

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
	msg.WriteString(logisticsHeader)
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
	return nil
}

func (h *handler) Join(ctx context.Context, player *entities.Player, recipientID string, order string) error {
	// TODO: write a real join message
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
	var err error
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

	err = h.speaker.SendDM(recipientID, fmt.Sprintf(joinFormat, player.MartialOrder, player.FormatClass(), player.Location))
	if err != nil {
		return errors.Wrap(err, "failed to send join message")
	}
	return nil
}

func (h *handler) Move(ctx context.Context, player *entities.Player, recipientID string, locationString string) error {
	const notFound = `
That's not a place that I know of.
`
	const notAdjacent = `
That's not an adjacent location.'
`
	const moving = `
You are now moving to %s.
`

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

	if locationID != player.Location {
		adjacentLocations, err := h.resource.GetAdjacentLocations(ctx, player.Location)
		if err != nil {
			return errors.Wrap(err, "failed moving player")
		}

		found := false
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

	err = h.resource.MovePlayer(ctx, recipientID, locationID)
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

func (h *handler) Advance(ctx context.Context, player *entities.Player, recipientID string, class string) error {
	return nil
}

// TODO: remember to deactivate instead of deleting
func (h *handler) Quit(ctx context.Context, player *entities.Player, recipientID string) error {
	quitMsg := `
Heaven's Gate closes behind you.
`

	err := h.resource.DeletePlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed quitting game")
	}

	err = h.speaker.SendDM(recipientID, quitMsg)
	if err != nil {
		return errors.Wrap(err, "failed to send quit message")
	}
	return nil
}

func (h *handler) ToggleUpdates(ctx context.Context, player *entities.Player, recipientID string) error {
	const noUpdates = `
You will no longer receive daily personal battle reports.
`
	const yesUpdates = `
You will now receive daily personal battle reports.
`

	receiveUpdates, err := h.resource.TogglePlayerUpdates(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed toggling updates")
	}

	if receiveUpdates {
		err = h.speaker.SendDM(recipientID, yesUpdates)
	} else {
		err = h.speaker.SendDM(recipientID, noUpdates)
	}
	if err != nil {
		return errors.Wrap(err, "failed sending toggle updates message")
	}
	return nil
}

func (h *handler) InvalidCommand(ctx context.Context, player *entities.Player, recipientID string) error {
	const invalid = `
That's not something I understand. Try seeking !help.
`

	err := h.speaker.SendDM(recipientID, invalid)
	if err != nil {
		return errors.Wrap(err, "failed sending invalid command message")
	}
	return nil
}

func (h *handler) Echo(ctx context.Context, player *entities.Player, recipientID string, msg string) error {
	err := h.speaker.SendDM(recipientID, "Just got the message: "+msg)
	if err != nil {
		return errors.Wrap(err, "failed sending echo message")
	}
	return nil
}
