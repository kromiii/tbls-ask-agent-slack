package slackhandler

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kromiii/tbls-ask-agent-slack/openai"
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
		channelInfo, err := h.Api.GetConversationInfo(&slack.GetConversationInfoInput{
			ChannelID: ev.Channel,
		})
		if err != nil {
			return fmt.Errorf("failed to get channel info: %w", err)
		}

		// Load schemas from config
		configBytes, err := fileLoader("./schemas/config.yml")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		var config Config
		err = yaml.Unmarshal(configBytes, &config)
		if err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}

		// Check if channel name contains any schema name
		var matchedSchema *Schema
		for _, schema := range config.Schemas {
			if strings.Contains(strings.ToLower(channelInfo.Name), strings.ToLower(schema.Name)) {
				matchedSchema = &schema
				break
			}
		}

		if matchedSchema != nil {
			// If a matching schema is found, use it directly
			response, err := h.Api.AuthTest()
			if err != nil {
				return fmt.Errorf("failed to get bot user ID: %w", err)
			}
			botUserID := response.UserID

			var messages []slack.Message
			var threadTS string
			if ev.ThreadTimeStamp != "" {
				messages, _, _, err = h.Api.GetConversationReplies(
					&slack.GetConversationRepliesParameters{
						ChannelID: ev.Channel,
						Timestamp: ev.ThreadTimeStamp,
						Inclusive: true,
					},
				)
				if err != nil {
					return fmt.Errorf("failed to get conversation replies: %w", err)
				}
				threadTS = ev.ThreadTimeStamp
			} else {
				// If it's not in a thread, just use the current message
				messages = []slack.Message{
					{
						Msg: slack.Msg{
							User:    ev.User,
							Text:    ev.Text,
							Channel: ev.Channel,
						},
					},
				}
				threadTS = ev.TimeStamp
			}

			answer := openai.Ask(messages, matchedSchema.Path, botUserID)

			_, _, err = h.Api.PostMessage(
				ev.Channel,
				slack.MsgOptionText(answer, false),
				slack.MsgOptionTS(threadTS),
			)
			if err != nil {
				return fmt.Errorf("failed to post message: %w", err)
			}
		} else {
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
						Text: "Please select the target schema",
					},
					Accessory: &slack.Accessory{
						SelectElement: &slack.SelectBlockElement{
							ActionID: "select_schema",
							Type:     slack.OptTypeStatic,
							Placeholder: &slack.TextBlockObject{
								Type: slack.PlainTextType,
								Text: "Select schema",
							},
							Options: options,
						},
					},
				},
			), slack.MsgOptionTS(ev.TimeStamp))

			if err != nil {
				return err
			}
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

		response, err := h.Api.AuthTest()
		if err != nil {
			return err
		}
		botUserID := response.UserID

		answer := openai.Ask(messages, action.SelectedOption.Value, botUserID)

		_, _, err = h.Api.PostMessage(
			interaction.Channel.ID,
			slack.MsgOptionText(answer, false),
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
