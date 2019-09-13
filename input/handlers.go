package input

import (
	"context"
	"fmt"
	"strings"

	"github.com/yisaj/heavens_throne/database"
	"github.com/yisaj/heavens_throne/twitspeak"

	"github.com/pkg/errors"
)

type InputHandler interface {
	Help(ctx context.Context, recipientID string) error
	Status(ctx context.Context, recipientID string) error
	Logistics(ctx context.Context, recipientID string) error
	Join(ctx context.Context, recipientID string, order string) error
	Move(ctx context.Context, recipientID string, location string) error
	Advance(ctx context.Context, recipientID string, class string) error
	Quit(ctx context.Context, recipientID string) error
	ToggleUpdates(ctx context.Context, recipientID string) error
	InvalidCommand(ctx context.Context, recipientID string) error
	Echo(ctx context.Context, recipientID string, msg string) error
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

func (h *handler) Help(ctx context.Context, recipientID string) error {
	const newPlayerHelp = `
	Type !join [order] to join.
`
	const activePlayerHelp = `
	!status to see your status.
`

	player, err := h.resource.GetPlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed checking player activeness")
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
	// TODO: write a real status message
	// TODO: handle available advances text
	const statusFormat = `
Order: %s
Class: %s
Experience: %d
Location: %s
`

	player, err := h.resource.GetPlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed sending player status")
	}
	if player != nil {
		location, err := h.resource.GetLocation(ctx, player.Location)
		if err != nil || location == nil {
			return errors.Wrap(err, "failed sending player status")
		}

		msg := fmt.Sprintf(statusFormat, player.MartialOrder, player.FormatClass(), player.Experience, location.Name)
		err = h.speaker.SendDM(recipientID, msg)
	}

	return nil
}

func (h *handler) Logistics(ctx context.Context, recipientID string) error {
	return nil
}

func (h *handler) Join(ctx context.Context, recipientID string, order string) error {
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

	player, err := h.resource.GetPlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed checking player during join")
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

	err = h.speaker.SendDM(recipientID, fmt.Sprintf(joinFormat, player.MartialOrder, player.FormatClass(), player.Location))
	if err != nil {
		return errors.Wrap(err, "failed to send join message")
	}
	return nil
}

func (h *handler) Move(ctx context.Context, recipientID string, location string) error {
	return nil
}

func (h *handler) Advance(ctx context.Context, recipientID string, class string) error {
	return nil
}

// TODO: remember to deactivate instead of deleting
func (h *handler) Quit(ctx context.Context, recipientID string) error {
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

func (h *handler) ToggleUpdates(ctx context.Context, recipientID string) error {
	return nil
}

func (h *handler) InvalidCommand(ctx context.Context, recipientID string) error {
	return nil
}

func (h *handler) Echo(ctx context.Context, recipientID string, msg string) error {
	return h.speaker.SendDM(recipientID, "Just got the message: "+msg)
}
