package main

import (
	"context"
	"fmt"
	"github.com/blissfulreboot/slagbot/internal/commandparser"
	"github.com/blissfulreboot/slagbot/internal/configuration"
	"github.com/blissfulreboot/slagbot/internal/pluginloader"
	"github.com/blissfulreboot/slagbot/internal/slackconnection"
	"github.com/blissfulreboot/slagbot/pkg/logging"
	"os"
	"os/signal"
	"sync"
)

func main() {
	conf, confErr := configuration.ReadConfiguration()
	if confErr != nil {
		fmt.Println(confErr)
		os.Exit(1)
	}

	logger := logging.NewLogger(conf.LogLevel, conf.LogEncoding)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slackbot, botCreateErr := slackconnection.NewBot(conf.SlackAppToken, conf.SlackBotToken, logger)
	if botCreateErr != nil {
		logger.Error(botCreateErr.Error())
		return
	}

	slackbot.Start(wg, ctx)

	logger.Debug("After slackconnection.Start")

	plugins, pluginLoaderErr := pluginloader.LoadPlugins(conf.PluginDir, conf.PluginExtension, conf.PluginExitGraceSeconds,
		logger, slackbot.OutgoingMessageChannel, wg, ctx)

	if pluginLoaderErr != nil {
		logger.Error(pluginLoaderErr.Error())
		return
	}
	logger.Debug("After utils.LoadPlugins")

	commandHandler := commandparser.NewCommandHandler(slackbot.IncomingMessageChannel, slackbot.OutgoingMessageChannel,
		plugins, logger)
	logger.Debug("After utils.NewCommandHandler")

	commandHandler.StartCommandHandlingLoop(wg, ctx)
	logger.Debug("After StartCommandHandlingLoop")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		wg.Add(1)
		defer wg.Done()

		<-c
		cancel()
	}()

	wg.Wait()

}
