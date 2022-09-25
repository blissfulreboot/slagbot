package configuration

import (
	"fmt"
	"gitlab.com/blissfulreboot/golang/conffee"
	"os"
)

type Configuration struct {
	LogLevel               string
	LogEncoding            string
	PluginDir              string
	PluginExtension        string
	PluginExitGraceSeconds uint
	SlackAppToken          string `conffee:"required=true"`
	SlackBotToken          string `conffee:"required=true"`
}

func ReadConfiguration() (*Configuration, error) {
	conf := Configuration{
		LogLevel:               "info",
		LogEncoding:            "console",
		PluginDir:              "./",
		PluginExtension:        ".plugin",
		PluginExitGraceSeconds: 5,
		SlackAppToken:          "",
		SlackBotToken:          "",
	}
	err := conffee.ReadConfiguration("./slagbot.conf", &conf, false, true)
	if err != nil {
		return nil, err
	}

	if !(conf.LogEncoding == "json" || conf.LogEncoding == "console") {
		fmt.Println("Log encoding must be either 'json' or 'console' if defined. Default is 'console' if left undefined.")
		os.Exit(1)
	}
	return &conf, nil
}
