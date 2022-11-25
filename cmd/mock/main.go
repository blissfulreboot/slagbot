package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/blissfulreboot/slagbot/internal/commandparser"
	"github.com/blissfulreboot/slagbot/internal/configuration"
	"github.com/blissfulreboot/slagbot/internal/pluginloader"
	"github.com/blissfulreboot/slagbot/internal/slackconnection"
	"github.com/blissfulreboot/slagbot/pkg/logging"
	"github.com/blissfulreboot/slagbot/pkg/types"
	"os"
	"os/signal"
	"sync"
	"time"
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

	incomingMessagesChannel := make(chan slackconnection.SlackMessage)
	outgoingMessageChannel := make(chan types.OutgoingSlackMessage)

	plugins, pluginLoaderErr := pluginloader.LoadPlugins(conf.PluginDir, conf.PluginExtension, conf.PluginExitGraceSeconds,
		logger, outgoingMessageChannel, wg, ctx)

	if pluginLoaderErr != nil {
		logger.Error(pluginLoaderErr)
		return
	}
	logger.Debug("After utils.LoadPlugins")

	commandHandler := commandparser.NewCommandHandler(incomingMessagesChannel, outgoingMessageChannel, plugins, logger)
	logger.Debug("After utils.NewCommandHandler")

	commandHandler.StartCommandHandlingLoop(wg, ctx)
	logger.Debug("After StartCommandHandlingLoop")

	// Start outgoingMessageChannel reader
	go func() {
		fmt.Println("Starting to listen outgoing messages")
		wg.Add(1)
		defer wg.Done()
		for {
			select {
			case msg := <-outgoingMessageChannel:
				fmt.Println("Received message from a plugin")
				fmt.Printf("Content of the message: %+v\n\n", msg)
			case <-ctx.Done():
				fmt.Println("Closing outgoing message listener")
				return
			}
		}
	}()

	// Start incomingMessageChannel writer
	go func() {
		wg.Add(1)
		defer wg.Done()
		time.Sleep(3 * time.Second)
		fmt.Println("Write to simulate slack messages going to the bot (no need to address the bot).")
		textChannel := make(chan string)
		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			for {
				fmt.Print("input: ")
				if !scanner.Scan() {
					break
				}
				textChannel <- scanner.Text()
			}
			fmt.Println("Text scanner exited with error. Exiting.")
			os.Exit(1)
		}()
		for {
			select {
			case text := <-textChannel:
				msg := slackconnection.SlackMessage{
					User:    "MockUser",
					Text:    text,
					Channel: "MockChannel",
				}
				incomingMessagesChannel <- msg
			case <-ctx.Done():
				fmt.Println("Exiting input handler.")
				return
			}

		}
	}()

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
