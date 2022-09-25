# slagbot

# Running the bot

To see what CLI parameters the slagbot accepts, run the binary with `-h`: `slagbot -h`

# Compiling

Compressing the executables requires UPX (https://upx.github.io/). Notice that the plugins should not be compressed as it causes a segfault.

Build all (without compression): `make all`

Build all (executables compressed): `make all-with-upx`

Clean: `make clean`

# Develoging plugins

## Required symbols

All plugins MUST implement the following functions:

````go
func GetCommands() []types.Command {}
func Run(cmdChannel chan types.ParsedCommand, slackMsgChannel chan<- types.OutgoingSlackMessage, logger interfaces.LoggerInterface) {}
func Stop() {}
````

## Adding commands

Each command consists of **_Keyword_**, **_Description_** and **_Parameters_**. Keyword is used to identify which command is called. For example, if the Keyword is `blissfulreboot`, then the slackbot looks if the message contains that keyword. If there is a match, then the message is parsed and sent to the plugin that owns the command. The keyword can consist of multiple words. Each command can have multiple parameters. These have **_Keyword_**, **_Description_** and **_Type_**. The Keyword works much like the command keyword does, but a value or a flag can be stored. Type defines what is stored and from where. Valid values for the type are _before_, _after_ and _flag_. If the type is flag, then a boolean true is stored, otherwise the parser takes the previous or next word (limited by spaces), and stores it. All parameters are passed to the plugin in a map, where the key is the parameter's keyword.

**_Example_**:

````go
[]types.Command{{
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
````

## Developing plugins without actual Slack

For this, there is the `mock` client that provides the possibility to write "slack" messages to the bot so that it can parse them and send to plugins.
