package input

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
)

type DMParser interface {
	ParseDM(ctx context.Context, recipientID string, msg string) error
}

type parser struct {
	inputHandler InputHandler
	logger       *logrus.Logger
}

func NewDMParser(inputHandler InputHandler, logger *logrus.Logger) DMParser {
	return &parser{
		inputHandler,
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

	p.logger.Infof("got command: `%s`, argument: `%s`", command, argument)

	switch command {
	case "!help":
		return p.inputHandler.Help(ctx, recipientID)
	case "!status":
		return p.inputHandler.Status(ctx, recipientID)
	case "!logistics":
		return p.inputHandler.Logistics(ctx, recipientID)
	case "!join":
		return p.inputHandler.Join(ctx, recipientID, argument)
	case "!move":
		return p.inputHandler.Move(ctx, recipientID, argument)
	case "!advance":
		return p.inputHandler.Advance(ctx, recipientID, argument)
	case "!quit":
		return p.inputHandler.Quit(ctx, recipientID)
	case "!toggleupdates":
		return p.inputHandler.ToggleUpdates(ctx, recipientID)
	default:
		return p.inputHandler.InvalidCommand(ctx, recipientID)
	}
}
