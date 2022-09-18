package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"gitlab.com/blissfulreboot/golang/conffee"
	"os"
	"os/signal"
	"slagbot/internal/commandparser"
	"slagbot/internal/pluginloader"
	"slagbot/internal/slackconnection"
	"sync"
)

type Configuration struct {
	PluginDir              string
	PluginExtension        string
	PluginExitGraceSeconds uint
	SlackAppToken          string `conffee:"required=true"`
	SlackBotToken          string `conffee:"required=true"`
}

func main() {
	log.SetLevel(log.TraceLevel)

	conf := Configuration{
		PluginDir:              "./",
		PluginExtension:        ".plugin",
		PluginExitGraceSeconds: 5,
		SlackAppToken:          "",
		SlackBotToken:          "",
	}
	conffee.ReadConfiguration("./slagbot.conf", &conf, false, true)
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slackbot, botCreateErr := slackconnection.NewBot(conf.SlackAppToken, conf.SlackBotToken)
	if botCreateErr != nil {
		log.Errorln(botCreateErr)
		return
	}

	slackbot.Start(wg, ctx)

	log.Debugln("After slackconnection.Start")

	plugins, pluginLoaderErr := pluginloader.LoadPlugins(conf.PluginDir, conf.PluginExtension, conf.PluginExitGraceSeconds,
		slackbot.OutgoingMessageChannel, wg, ctx)

	if pluginLoaderErr != nil {
		log.Errorln(pluginLoaderErr)
		return
	}
	log.Debugln("After utils.LoadPlugins")

	commandHandler := commandparser.NewCommandHandler(slackbot.IncomingMessageChannel, slackbot.OutgoingMessageChannel, plugins)
	log.Debugln("After utils.NewCommandHandler")

	commandHandler.StartCommandHandlingLoop(wg, ctx)
	log.Debugln("After StartCommandHandlingLoop")

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
