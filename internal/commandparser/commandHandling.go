package commandparser

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"regexp"
	"slagbot/internal/pluginloader"
	slackbot2 "slagbot/internal/slackconnection"
	"slagbot/pkg/types"
	"strings"
	"sync"
)

type CommandHandler struct {
	incomingMsgChannel chan slackbot2.SlackMessage
	outgoingMsgChannel chan types.OutgoingSlackMessage
	plugins            []*pluginloader.ReadyPlugin
}

func NewCommandHandler(incoming chan slackbot2.SlackMessage, outgoing chan types.OutgoingSlackMessage, plugins []*pluginloader.ReadyPlugin) *CommandHandler {
	return &CommandHandler{
		plugins:            plugins,
		incomingMsgChannel: incoming,
		outgoingMsgChannel: outgoing,
	}
}

func (ch *CommandHandler) StartCommandHandlingLoop(wg *sync.WaitGroup, ctx context.Context) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Debugln("Starting StartCommandHandlingLoop")
		for {
			select {
			case msg := <-ch.incomingMsgChannel:
				log.Debugf("StartCommandHandlingLoop received message: %+v\n", msg)
				parseErr := ch.handleMessage(msg)
				if parseErr != nil {
					log.Errorln("Failed to parse the command.")
					ch.outgoingMsgChannel <- types.OutgoingSlackMessage{
						Channel:   &msg.Channel,
						UserEmail: nil,
						Message:   "Failed to parse the command",
					}
				}

			case <-ctx.Done():
				log.Debugln("Context done in StartCommandHandlingLoop")
				return
			}
		}
	}()
	log.Info("StartCommandHandlingLoop done")
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

func (ch *CommandHandler) handleMessage(message slackbot2.SlackMessage) error {
	var msgCommand string
	var args types.Arguments
	for _, plug := range ch.plugins {
		log.Debugf("Cheking plugin %s for matching commands", plug.File)
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
	log.Debugf("Message handled: %+v\n", message)
	log.Debugln("No command match found.")
	return errors.New("no command found")
}
