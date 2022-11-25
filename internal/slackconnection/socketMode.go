package slackconnection

import (
	"context"
	"github.com/blissfulreboot/slagbot/pkg/interfaces"
	"github.com/blissfulreboot/slagbot/pkg/types"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
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
	logger                 interfaces.LoggerInterface
}

func NewBot(appToken string, botToken string, logger interfaces.LoggerInterface) (*Bot, error) {
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
		slack.OptionAppLevelToken(appToken),
	)
	// Get the slackconnection's id
	response, authTestErr := api.AuthTest()
	if authTestErr != nil {
		return nil, authTestErr
	}
	slackbotSelfId := response.UserID

	logger.Info("Slackbot's UserID ", slackbotSelfId)

	client := socketmode.New(
		api,
	)

	return &Bot{
		IncomingMessageChannel: make(chan SlackMessage),
		OutgoingMessageChannel: make(chan types.OutgoingSlackMessage),
		client:                 client,
		slackbotSelfId:         slackbotSelfId,
		logger:                 logger,
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
	b.logger.Info("Connecting to Slack with Socket Mode...")
}

func (b *Bot) middlewareConnectionError(evt *socketmode.Event, client *socketmode.Client) {
	b.logger.Info("Connection failed. Retrying later...")
}

func (b *Bot) middlewareConnected(evt *socketmode.Event, client *socketmode.Client) {
	b.logger.Info("Connected to Slack with Socket Mode.")
}

func (b *Bot) incomingMessageHandler(evt *socketmode.Event, client *socketmode.Client) {
	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		b.logger.Debugf("> Ignored %+v", evt)
		return
	}
	client.Ack(*evt.Request)

	var slackMessage SlackMessage

	switch eventData := eventsAPIEvent.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		b.logger.Debugf("AppMentionEvent: %+v", eventData)
		slackMessage = SlackMessage{
			User:    eventData.User,
			Text:    eventData.Text,
			Channel: eventData.Channel,
		}
	case *slackevents.MessageEvent:
		b.logger.Debugf("MessageEvent: %+v", eventData)
		slackMessage = SlackMessage{
			User:    eventData.User,
			Text:    eventData.Text,
			Channel: eventData.Channel,
		}
	default:
		b.logger.Error("Unknown message event")
		b.logger.Debugf("Data: %+v", eventData)
	}

	if slackMessage.User == b.slackbotSelfId {
		b.logger.Debugf("Ignoring own message: %+v", slackMessage)
		return
	}

	b.logger.Debugf("slackMessage: %+v", slackMessage)

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
				if msg.Channel != "" {
					channelId = msg.Channel
				} else if msg.UserEmail != "" {
					b.logger.Debug(msg.UserEmail)
					user, getUserErr := b.client.GetUserByEmail(msg.UserEmail)
					if getUserErr != nil {
						b.logger.Errorf("User with email %s not found", msg.UserEmail)
						b.logger.Debug(getUserErr)
						continue
					}
					channelId = user.ID
				} else {
					b.logger.Error("User email and channel id cannot both be nil. Message was not sent.")
					b.logger.Debugf("Message: %s", msg.Message)
				}
				_, _, err := b.client.Client.PostMessage(channelId, slack.MsgOptionText(msg.Message, false))
				if err != nil {
					b.logger.Errorf("failed posting message: %v", err)
					b.logger.Debugf("Message: %s, Channel: %s", msg.Message, channelId)
				}
			case <-ctx.Done():
				b.logger.Debug("Context done in startOutgoingMessageHandler")
				return
			}
		}
	}()
	b.logger.Debug("startOutgoingMessageHandler Done")
}
