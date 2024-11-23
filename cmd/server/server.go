package server

import (
	"fmt"
	"log"
	"os"

	"github.com/kromiii/tbls-ask-agent-slack/slackhandler"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func Run() {
	oauthToken := os.Getenv("SLACK_OAUTH_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	api := slack.New(oauthToken, slack.OptionAppLevelToken(appToken))
	client := socketmode.New(api)

	slackHandler := slackhandler.SlackHandler{
		Api: api,
	}

	path := "./schemas/config.yml"

	go func() {
		for socketEvent := range client.Events {
			switch socketEvent.Type {
			case socketmode.EventTypeConnecting:
				fmt.Println("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				fmt.Println("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				fmt.Println("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				event, ok := socketEvent.Data.(slackevents.EventsAPIEvent)
				if !ok {
					continue
				}
				client.Ack(*socketEvent.Request)
				err := slackHandler.HandleCallBackEvent(event, path)
				if err != nil {
					log.Print(err)
				}
			case socketmode.EventTypeInteractive:
				interaction, ok := socketEvent.Data.(slack.InteractionCallback)
				if !ok {
					continue
				}
				client.Ack(*socketEvent.Request)
				err := slackHandler.HandleInteractionCallback(interaction)
				if err != nil {
					log.Print(err)
				}
			}
		}
	}()

	err := client.Run()
	if err != nil {
		log.Print(err)
	}
}
