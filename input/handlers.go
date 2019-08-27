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
	const helpMessage = `

`

	err := h.speaker.SendDM(recipientID, helpMessage)
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
		err = h.notPlaying(ctx, recipientID)
	} else {
		msg := fmt.Sprintf(statusFormat, player.FormatOrder(), player.FormatClass(), player.Experience, player.Location)
		err = h.speaker.SendDM(recipientID, msg)
	}
}

func (h *handler) Logistics(ctx context.Context, recipientID string) error {

}

func (h *handler) Join(ctx context.Context, recipientID string, order string) error {
	// TODO: write a real join message
	const joinFormat = `
ORDER: %s
CLASS: %s
LOCATION: %s
`
	// TODO: handle starting location of new users
	var orderName string
	if strings.Contains(order, "staghorn") {
		orderName = "staghorn"
	} else if strings.Contains(order, "gorgona") {
		orderName = "gorgona"
	} else if strings.Contains(order, "baaturate") {
		orderName = "baaturate"
	}

	player, err := h.resource.CreatePlayer(ctx, recipientID, orderName, "location")
	if err != nil {
		return errors.Wrap(err, "failed joining new player")
	}

	err = h.speaker.SendDM(recipientID, fmt.Sprintf(joinFormat, player.FormatOrder(), player.FormatClass(), player.Location))
	if err != nil {
		errors.Wrap(err, "failed to send join message")
	}
	return nil
}

func (h *handler) Move(ctx context.Context, recipientID string, location string) error {

}

func (h *handler) Advance(ctx context.Context, recipientID string, class string) error {

}

func (h *handler) Quit(ctx context.Context, recipientID string) error {

}

func (h *handler) ToggleUpdates(ctx context.Context, recipientID string) error {

}

func (h *handler) InvalidCommand(ctx context.Context, recipientID string) error {

}

func (h *handler) notPlaying(ctx context.Context, recipientID string) error {
	const notPlayingMessage = `

`

	err := h.speaker.SendDM(recipientID, notPlayingMessage)
	if err != nil {
		return errors.Wrap(err, "failed sending not playing message")
	}
	return nil
}
