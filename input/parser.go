package input

import (
	"context"
	"strings"
)

type DMParser interface {
	ParseDM(ctx context.Context, recipientID string, msg string) error
}

type parser struct {
	inputHandler InputHandler
}

func NewDMParser(inputHandler InputHandler) DMParser {
	return &parser{
		inputHandler,
	}
}

func (p *parser) ParseDM(ctx context.Context, recipientID string, msg string) error {
	// look for command and tokenize the message
	bangString := msg[strings.IndexByte(msg, '!'):] + " "
	tokenizedCommand := strings.SplitN(bangString, " ", 2)
	command, argument := tokenizedCommand[0], tokenizedCommand[1]

	_, _ = command, argument

	return p.inputHandler.Echo(ctx, recipientID, msg)

	/*
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
	*/
}
