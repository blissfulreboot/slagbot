package slackconnection

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"slagbot/pkg/types"
	"strings"
	"sync"
)

/*
This code is based on the socketmode event handler example:
https://github.com/slack-go/slack/blob/master/examples/socketmode_handler/socketmode_handler.go
*/

type SlackMessage struct {
	User    string
	Text    string
	Channel string
}

type Bot struct {
	IncomingMessageChannel chan SlackMessage
	OutgoingMessageChannel chan types.OutgoingSlackMessage
	client                 *socketmode.Client
	slackbotSelfId         string
}

func NewBot(appToken string, botToken string) (*Bot, error) {
	if appToken == "" {
		panic("SLACK_APP_TOKEN must be set.\n")
	}

	if !strings.HasPrefix(appToken, "xapp-") {
		panic("SLACK_APP_TOKEN must have the prefix \"xapp-\".")
	}

	if botToken == "" {
		panic("SLACK_BOT_TOKEN must be set.\n")
	}

	if !strings.HasPrefix(botToken, "xoxb-") {
		panic("SLACK_BOT_TOKEN must have the prefix \"xoxb-\".")
	}

	api := slack.New(
		botToken,
		//slack.OptionDebug(true),
		//slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken),
	)
	// Get the slackconnection's id
	response, authTestErr := api.AuthTest()
	if authTestErr != nil {
		return nil, authTestErr
	}
	slackbotSelfId := response.UserID

	log.Info("Slackbot's UserID ", slackbotSelfId)

	client := socketmode.New(
		api,
		//socketmode.OptionDebug(true),
		//socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	return &Bot{
		IncomingMessageChannel: make(chan SlackMessage),
		OutgoingMessageChannel: make(chan types.OutgoingSlackMessage),
		client:                 client,
		slackbotSelfId:         slackbotSelfId,
	}, nil
}

func (b *Bot) Start(wg *sync.WaitGroup, ctx context.Context) {
	socketmodeHandler := socketmode.NewSocketmodeHandler(b.client)

	socketmodeHandler.Handle(socketmode.EventTypeConnecting, b.middlewareConnecting)
	socketmodeHandler.Handle(socketmode.EventTypeConnectionError, b.middlewareConnectionError)
	socketmodeHandler.Handle(socketmode.EventTypeConnected, b.middlewareConnected)

	//\\ EventTypeEventsAPI //\\
	// Handle all EventsAPI
	//socketmodeHandler.Handle(socketmode.EventTypeEventsAPI, middlewareEventsAPI)

	// Handle a specific event from EventsAPI
	socketmodeHandler.HandleEvents(slackevents.AppMention, b.incomingMessageHandler)
	socketmodeHandler.HandleEvents(slackevents.Message, b.incomingMessageHandler)

	// Channels for incoming and outgoing messages must be created before starting the handler loops

	go socketmodeHandler.RunEventLoop()

	b.startOutgoingMessageHandler(wg, ctx)
}

func (b *Bot) middlewareConnecting(evt *socketmode.Event, client *socketmode.Client) {
	fmt.Println("Connecting to Slack with Socket Mode...")
}

func (b *Bot) middlewareConnectionError(evt *socketmode.Event, client *socketmode.Client) {
	fmt.Println("Connection failed. Retrying later...")
}

func (b *Bot) middlewareConnected(evt *socketmode.Event, client *socketmode.Client) {
	fmt.Println("Connected to Slack with Socket Mode.")
}

func (b *Bot) incomingMessageHandler(evt *socketmode.Event, client *socketmode.Client) {
	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		log.Debugf("> Ignored %+v\n", evt)
		return
	}
	client.Ack(*evt.Request)

	var slackMessage SlackMessage

	switch eventData := eventsAPIEvent.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		log.Debugf("AppMentionEvent: %+v\n", eventData)
		slackMessage = SlackMessage{
			User:    eventData.User,
			Text:    eventData.Text,
			Channel: eventData.Channel,
		}
	case *slackevents.MessageEvent:
		log.Debugf("MessageEvent: %+v\n", eventData)
		slackMessage = SlackMessage{
			User:    eventData.User,
			Text:    eventData.Text,
			Channel: eventData.Channel,
		}
	default:
		log.Errorln("Unknown message event")
		log.Debugf("Data: %+v\n", eventData)
	}

	if slackMessage.User == b.slackbotSelfId {
		log.Debugf("Ignoring own message: %+v\n", slackMessage)
		return
	}

	log.Debugf("slackMessage: %+v\n", slackMessage)

	b.IncomingMessageChannel <- slackMessage

}

func (b *Bot) startOutgoingMessageHandler(wg *sync.WaitGroup, ctx context.Context) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case msg := <-b.OutgoingMessageChannel:
				var channelId string
				if msg.Channel != nil {
					channelId = *msg.Channel
				} else if msg.UserEmail != nil {
					log.Debug(*msg.UserEmail)
					user, getUserErr := b.client.GetUserByEmail(*msg.UserEmail)
					if getUserErr != nil {
						log.Errorf("User with email %s not found\n", *msg.UserEmail)
						log.Debugln(getUserErr)
						continue
					}
					channelId = user.ID
				} else {
					log.Errorln("User email and channel id cannot both be nil. Message was not sent.")
					log.Debugf("Message: %s\n", msg.Message)
				}
				_, _, err := b.client.Client.PostMessage(channelId, slack.MsgOptionText(msg.Message, false))
				if err != nil {
					log.Errorf("failed posting message: %v", err)
					log.Debugf("Message: %s, Channel: %s\n", msg.Message, channelId)
				}
			case <-ctx.Done():
				log.Debugln("Context done in startOutgoingMessageHandler")
				return
			}
		}
	}()
	log.Debugln("startOutgoingMessageHandler Done")
}
