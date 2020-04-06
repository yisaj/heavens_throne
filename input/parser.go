package input

import (
	"context"
	"strings"

	"github.com/yisaj/heavens_throne/database"
	"github.com/yisaj/heavens_throne/twitspeak"

	"github.com/sirupsen/logrus"
)

// DMParser contains the logic to parse a player DM and call the appropriate
// player input handler
type DMParser interface {
	ParseDM(ctx context.Context, recipientID string, msg string) error
}

// player input parsers need to be able to get player info from the database and
// call the appropriate handler
type parser struct {
	inputHandler Handler
	logger       *logrus.Logger
}

// NewDMParser constructs a new parser to parse player input
func NewDMParser(resource database.Resource, speaker twitspeak.TwitterSpeaker, logger *logrus.Logger) DMParser {
	return &parser{
		newInputHandler(resource, speaker),
		logger,
	}
}

// ParseDM takes a player DM and executes the appropriate logic
func (p *parser) ParseDM(ctx context.Context, recipientID string, msg string) error {
	// look for command and tokenize the message
	bangIndex := strings.IndexByte(msg, '!')
	if bangIndex == -1 {
		bangIndex = 0
	}
	bangString := msg[bangIndex:]
	tokenizedCommand := strings.SplitN(bangString, " ", 2)
	command, argument := tokenizedCommand[0], ""
	if len(tokenizedCommand) > 1 {
		argument = tokenizedCommand[1]
	}

	p.logger.Infof("got command: `%s`, argument: `%s` from `%s`", command, argument, recipientID)

	switch strings.ToLower(command) {
	case "!help", "help":
		return p.inputHandler.Help(ctx, recipientID)
	case "!status", "status":
		return p.inputHandler.Status(ctx, recipientID)
	case "!logistics", "logistics":
		return p.inputHandler.Logistics(ctx, recipientID, strings.ToLower(argument))
	case "!join", "join":
		return p.inputHandler.Join(ctx, recipientID, strings.ToLower(argument))
	case "!move", "move":
		return p.inputHandler.Move(ctx, recipientID, strings.ToLower(argument))
	case "!advance", "advance":
		return p.inputHandler.Advance(ctx, recipientID, strings.ToLower(argument))
	case "!quit", "quit":
		return p.inputHandler.Quit(ctx, recipientID)
	case "!toggleupdates", "toggleupdates":
		return p.inputHandler.ToggleUpdates(ctx, recipientID)
	case "!echo", "echo":
		return p.inputHandler.Echo(ctx, recipientID, argument)
	case "!simulate", "simulate":
		return p.inputHandler.Simulate(ctx, recipientID)
	case "!tweet", "tweet":
		return p.inputHandler.Tweet(ctx, recipientID, argument)
	case "!reply", "reply":
		return p.inputHandler.Reply(ctx, recipientID, argument)
	default:
		return p.inputHandler.InvalidCommand(ctx, recipientID)
	}
}
