package input

import (
	"context"
	"github.com/pkg/errors"
	"github.com/yisaj/heavens_throne/database"
	"strings"

	"github.com/sirupsen/logrus"
)

type DMParser interface {
	ParseDM(ctx context.Context, recipientID string, msg string) error
}

type parser struct {
	inputHandler Handler
	resource     database.Resource
	logger       *logrus.Logger
}

func NewDMParser(inputHandler Handler, resource database.Resource, logger *logrus.Logger) DMParser {
	return &parser{
		inputHandler,
		resource,
		logger,
	}
}

func (p *parser) ParseDM(ctx context.Context, recipientID string, msg string) error {
	// look for command and tokenize the message
	bangIndex := strings.IndexByte(msg, '!')
	if bangIndex == -1 {
		return nil
	}
	bangString := msg[strings.IndexByte(msg, '!'):] + " "
	tokenizedCommand := strings.SplitN(bangString, " ", 2)
	command, argument := tokenizedCommand[0], tokenizedCommand[1]

	player, err := p.resource.GetPlayer(ctx, recipientID)
	if err != nil {
		return errors.Wrap(err, "failed parsing DM")
	}

	p.logger.Infof("got command: `%s`, argument: `%s` from ", command, argument)

	switch command {
	case "!help":
		return p.inputHandler.Help(ctx, player, recipientID)
	case "!status":
		return p.inputHandler.Status(ctx, player, recipientID)
	case "!logistics":
		return p.inputHandler.Logistics(ctx, player, recipientID, argument)
	case "!join":
		return p.inputHandler.Join(ctx, player, recipientID, argument)
	case "!move":
		return p.inputHandler.Move(ctx, player, recipientID, argument)
	case "!advance":
		return p.inputHandler.Advance(ctx, player, recipientID, argument)
	case "!quit":
		return p.inputHandler.Quit(ctx, player, recipientID)
	case "!toggleupdates":
		return p.inputHandler.ToggleUpdates(ctx, player, recipientID)
	default:
		return p.inputHandler.InvalidCommand(ctx, player, recipientID)
	}
}
