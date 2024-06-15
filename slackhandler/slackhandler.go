package slackhandler

import (
	"errors"
	"log"
	"os"

	"github.com/kromiii/tbls-ask-agent-slack/tbls"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"gopkg.in/yaml.v2"
)

type SlackHandler struct {
	Api *slack.Client
}

type Schema struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type Config struct {
	Schemas []Schema `yaml:"schemas"`
}

var fileLoader = os.ReadFile

func (h *SlackHandler) HandleCallBackEvent(event slackevents.EventsAPIEvent) error {
	innerEvent := event.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		data, err := fileLoader("./schemas/config.yml")
		if err != nil {
			return err
		}

		var config Config
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return err
		}

		var options []*slack.OptionBlockObject
		for _, schema := range config.Schemas {
			options = append(options, &slack.OptionBlockObject{
				Text:  &slack.TextBlockObject{Type: slack.PlainTextType, Text: schema.Name},
				Value: schema.Path,
			})
		}

		_, _, err = h.Api.PostMessage(ev.Channel, slack.MsgOptionBlocks(
			slack.SectionBlock{
				Type: slack.MBTSection,
				Text: &slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: "対象のスキーマを選んでください",
				},
				Accessory: &slack.Accessory{
					SelectElement: &slack.SelectBlockElement{
						ActionID: "select_schema",
						Type:     slack.OptTypeStatic,
						Placeholder: &slack.TextBlockObject{
							Type: slack.PlainTextType,
							Text: "スキーマを選択",
						},
						Options: options,
					},
				},
			},
		), slack.MsgOptionTS(ev.TimeStamp))

		if err != nil {
			return err
		}
	default:
		return errors.New("unknown event")
	}
	return nil
}

func (h *SlackHandler) HandleInteractionCallback(interaction slack.InteractionCallback) error {
	if len(interaction.ActionCallback.BlockActions) != 1 {
		return errors.New("invalid request")
	}

	action := interaction.ActionCallback.BlockActions[0]
	switch action.ActionID {
	case "select_schema":
		threadTimestamp := interaction.Message.ThreadTimestamp
		messages, _, _, err := h.Api.GetConversationReplies(
			&slack.GetConversationRepliesParameters{
				ChannelID: interaction.Channel.ID,
				Timestamp: threadTimestamp,
				Inclusive: true,
			},
		)
		if err != nil {
			return err
		}

		// q := regexp.MustCompile(`<@U[0-9A-Za-z]+>`).ReplaceAllString(messages[0].Text, "")
		a := tbls.Ask(messages, action.SelectedOption.Value)
		_, _, err = h.Api.PostMessage(
			interaction.Channel.ID,
			slack.MsgOptionText(a, false),
			slack.MsgOptionTS(interaction.Message.Timestamp),
		)
		if err != nil {
			log.Printf("Failed to post message: %v", err)
		}
	default:
		return errors.New("unknown action")
	}
	return nil
}
