package commandparser

import (
	"context"
	"errors"
	"fmt"
	"github.com/blissfulreboot/slagbot/internal/pluginloader"
	"github.com/blissfulreboot/slagbot/internal/slackconnection"
	"github.com/blissfulreboot/slagbot/pkg/interfaces"
	"github.com/blissfulreboot/slagbot/pkg/types"
	"regexp"
	"strings"
	"sync"
)

type CommandHandler struct {
	incomingMsgChannel chan slackconnection.SlackMessage
	outgoingMsgChannel chan types.OutgoingSlackMessage
	plugins            []*pluginloader.ReadyPlugin
	logger             interfaces.LoggerInterface
}

func NewCommandHandler(incoming chan slackconnection.SlackMessage, outgoing chan types.OutgoingSlackMessage,
	plugins []*pluginloader.ReadyPlugin, logger interfaces.LoggerInterface) *CommandHandler {
	return &CommandHandler{
		plugins:            plugins,
		incomingMsgChannel: incoming,
		outgoingMsgChannel: outgoing,
		logger:             logger,
	}
}

func (ch *CommandHandler) StartCommandHandlingLoop(wg *sync.WaitGroup, ctx context.Context) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		ch.logger.Debug("Starting StartCommandHandlingLoop")
		for {
			select {
			case msg := <-ch.incomingMsgChannel:
				ch.logger.Debugf("StartCommandHandlingLoop received message: %+v", msg)
				parseErr := ch.handleMessage(msg)
				if parseErr != nil {
					ch.logger.Error("Failed to parse the command.")
					ch.outgoingMsgChannel <- types.OutgoingSlackMessage{
						Channel:   msg.Channel,
						UserEmail: "",
						Message:   "Failed to parse the command",
					}
				}

			case <-ctx.Done():
				ch.logger.Debug("Context done in StartCommandHandlingLoop")
				return
			}
		}
	}()
	ch.logger.Info("StartCommandHandlingLoop done")
}

func (ch *CommandHandler) parseArguments(message string, params []types.Parameter) (types.Arguments, error) {
	var err error
	args := make(types.Arguments)

	for _, param := range params {
		messageContainsKeyword := strings.Contains(message, param.Keyword)

		// Check if the parameter type is a flag since that requires some special handling
		if param.Type == types.Flag {
			args[param.Keyword] = false
			if messageContainsKeyword {
				args[param.Keyword] = true
			}
			continue
		}

		// Only with flag it is allowed for the Keyword to be absent
		if !messageContainsKeyword {
			return nil, errors.New(fmt.Sprintf("could not find Keyword '%s' when parsing arguments", param.Keyword))
		}
		// Compile the regexp for getting the argument
		var re *regexp.Regexp
		switch param.Type {
		case types.Before:
			re, err = regexp.Compile(fmt.Sprintf("\\s(\\S+)\\s%s", param.Keyword))
		case types.After:
			re, err = regexp.Compile(fmt.Sprintf("%s\\s(\\S+)", param.Keyword))
		default:
			return nil, errors.New(fmt.Sprintf("unsupported parameter type '%s'", param.Type))
		}
		// Handling for regexp compile errors
		if err != nil {
			return nil, err
		}

		results := re.FindStringSubmatch(message)
		if len(results) == 0 {
			return nil, errors.New(fmt.Sprintf("could not find match/value for parameter %s (param type %s)", param.Keyword, param.Type))
		}
		// If there is a match, then there should be at least two items in the slice. This should not be possible, but
		// perhaps there is some edge case?
		if len(results) == 1 {
			return nil, errors.New("only one item in the result slice, how is this possible")
		}
		args[param.Keyword] = results[1]
	}
	return args, nil
}

func (ch *CommandHandler) handleMessage(message slackconnection.SlackMessage) error {
	var msgCommand string
	var args types.Arguments
	for _, plug := range ch.plugins {
		ch.logger.Debugf("Cheking plugin %s for matching commands", plug.File)
		for _, cmd := range plug.Commands {
			if !strings.Contains(message.Text, cmd.Keyword) {
				continue
			}
			msgCommand = cmd.Keyword
			var err error
			args, err = ch.parseArguments(message.Text, cmd.Params)
			if err != nil {
				return err
			}
			plug.CommandChannel <- types.ParsedCommand{
				Channel:   message.Channel,
				Command:   msgCommand,
				Arguments: args,
			}
			return nil
		}
	}
	ch.logger.Debugf("Message handled: %+v", message)
	ch.logger.Debug("No command match found.")
	return errors.New("no command found")
}
