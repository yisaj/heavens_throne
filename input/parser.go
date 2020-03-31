package input

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/yisaj/heavens_throne/database"

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
	resource     database.Resource
	logger       *logrus.Logger
}

// NewDMParser constructs a new parser to parse player input
func NewDMParser(inputHandler Handler, resource database.Resource, logger *logrus.Logger) DMParser {
	return &parser{
		inputHandler,
		resource,
		logger,
	}
}

// ParseDM takes a player DM and executes the appropriate logic
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
