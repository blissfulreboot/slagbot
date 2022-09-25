package pluginloader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"slagbot/pkg/interfaces"
	"slagbot/pkg/types"
	"sync"
	"time"
)

type pluginGetCommandFunc func() []types.Command
type pluginRunFunc func(chan types.ParsedCommand, chan<- types.OutgoingSlackMessage, interfaces.LoggerInterface)
type pluginStopFunc func()

type ReadyPlugin struct {
	File           string
	getCommands    func() []types.Command
	run            func(chan types.ParsedCommand, chan<- types.OutgoingSlackMessage, interfaces.LoggerInterface)
	stop           func()
	Commands       []types.Command
	CommandChannel chan types.ParsedCommand
}

func preparePlugin(file string, plugin *plugin.Plugin) (*ReadyPlugin, error) {
	// Lookup the required symbols
	gcSymbol, gcSymbolLookupErr := plugin.Lookup("GetCommands")
	if gcSymbolLookupErr != nil {
		return nil, gcSymbolLookupErr
	}
	runSymbol, runSymbolLookupErr := plugin.Lookup("Run")
	if runSymbolLookupErr != nil {
		return nil, runSymbolLookupErr
	}
	stopSymbol, stopSymbolLookuplErr := plugin.Lookup("Stop")
	if stopSymbolLookuplErr != nil {
		return nil, stopSymbolLookuplErr
	}

	// Check that the symbols are functions
	gcFunc, gcSymbolAssertionOk := gcSymbol.(func() []types.Command)
	if gcSymbolAssertionOk != true {
		return nil, errors.New("the getCommands symbol is not a function")
	}
	runFunc, runSymbolAssertionOk := runSymbol.(func(chan types.ParsedCommand, chan<- types.OutgoingSlackMessage, interfaces.LoggerInterface))
	if runSymbolAssertionOk != true {
		return nil, errors.New("the run symbol is not a function")
	}
	stopFunc, stopSymbolAssertionOk := stopSymbol.(func())
	if stopSymbolAssertionOk != true {
		return nil, errors.New("the stop symbol is not a function")
	}

	commands := gcFunc()

	readyPlugin := ReadyPlugin{
		File:           file,
		getCommands:    gcFunc,
		run:            runFunc,
		stop:           stopFunc,
		Commands:       commands,
		CommandChannel: make(chan types.ParsedCommand),
	}
	return &readyPlugin, nil
}

func LoadPlugins(plugindir string, pluginExtension string, pluginGracePeriodSeconds uint, logger interfaces.LoggerInterface,
	slackMessageChannel chan<- types.OutgoingSlackMessage, wg *sync.WaitGroup, ctx context.Context) ([]*ReadyPlugin, error) {

	files, err := os.ReadDir(plugindir)
	if err != nil {
		return nil, err
	}
	var pluginFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == pluginExtension {
			pluginFiles = append(pluginFiles, file.Name())
		}
	}

	var loadedPlugins []*ReadyPlugin
	for _, file := range pluginFiles {
		logger.Infof("Attempting to load plugin %s", file)
		plug, pluginError := plugin.Open(file)
		if pluginError != nil {
			logger.Error(fmt.Sprintf("Could not load plug %s", file))
			logger.Debug(pluginError)
			continue
		}
		logger.Infof("Plugin %s loaded. Preparing it...", file)
		readyPlugin, initErr := preparePlugin(file, plug)
		if initErr != nil {
			return nil, initErr
		}
		logger.Infof("Plugin %s prepared. Calling the run function", file)
		go readyPlugin.run(readyPlugin.CommandChannel, slackMessageChannel, logger)

		loadedPlugins = append(loadedPlugins, readyPlugin)
	}

	// Handler for external stop signal
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				logger.Debug("Context done in LoadPlugins")
				for _, plug := range loadedPlugins {
					go plug.stop()
				}
				logger.Infof("stop called for all logging, waiting for %ds before continuing to allow graceful exit.", pluginGracePeriodSeconds)
				time.Sleep(time.Duration(pluginGracePeriodSeconds) * time.Second)
				return
			}
		}
		logger.Info("Done")
	}()

	return loadedPlugins, nil
}
