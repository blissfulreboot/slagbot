package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"slagbot/pkg/plugins"
	"slagbot/pkg/types"
	"strings"
)

var commandChannel chan types.ParsedCommand
var slackMessageChannel chan<- types.OutgoingSlackMessage
var pluginLogger plugins.PluginLogger
var ctx context.Context
var cancel context.CancelFunc

func Run(cmdChannel chan types.ParsedCommand, slackMsgChannel chan<- types.OutgoingSlackMessage, logger plugins.PluginLogger) {
	commandChannel = cmdChannel
	slackMessageChannel = slackMsgChannel
	pluginLogger = logger
	logger.Info("Run called")
	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case cmd := <-cmdChannel:
				logger.Info("Received: ", cmd.Command)
				logger.Info("Channel: ", cmd.Channel)
				for key, arg := range cmd.Arguments {
					logger.Info(key, ": ", arg)
				}

				switch cmd.Command {
				case "blissfulreboot":
					slackMsgChannel <- types.OutgoingSlackMessage{
						Channel:   &cmd.Channel,
						UserEmail: nil,
						Message:   "Thanks!",
					}
				case "on the channel":
					email, ok := cmd.Arguments["is very nice to"]
					if !ok {
						slackMsgChannel <- types.OutgoingSlackMessage{
							Channel:   &cmd.Channel,
							UserEmail: nil,
							Message:   "I think I did not understand this completely",
						}
						continue
					}
					emailString, ok := email.(string)
					if !ok {
						log.Info("Email is not a string")
						continue
					}
					firstParts := strings.Split(emailString, "|")
					finalParts := strings.Split(firstParts[0], ":")

					slackMsgChannel <- types.OutgoingSlackMessage{
						Channel:   nil,
						UserEmail: &finalParts[len(finalParts)-1],
						Message:   "Which is great, I think!",
					}
				}

			case <-ctx.Done():
				log.Info("Context done in plugin Run")
				return
			}
		}
	}()
}

func Stop() {
	pluginLogger.Info("Stop called")
	cancel()
}

func GetCommands() []types.Command {
	return []types.Command{{
		Keyword:     "blissfulreboot",
		Description: "foobar",
		Params: []types.Parameter{
			{
				Keyword:     "is nice to",
				Description: "foobar",
				Type:        "after",
			},
		},
	}, {
		Keyword:     "on the channel",
		Description: "foobar",
		Params: []types.Parameter{
			{
				Keyword:     "is very nice to",
				Description: "foobar",
				Type:        "before",
			},
		},
	}}
}
