package slackhandler

import (
	"errors"
	"os"

	// "regexp"

	// "github.com/kromiii/tbls-ask-server/tbls"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"gopkg.in/yaml.v2"
)

type SlackHandler struct {
	Api *slack.Client
}

type Config struct {
	Schemas []struct {
		Name string `yaml:"name"`
		Path string `yaml:"path"`
	} `yaml:"schemas"`
}

func (h *SlackHandler) HandleCallBackEvent(event slackevents.EventsAPIEvent) error {
	innerEvent := event.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		data, err := os.ReadFile("config.yml")
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

		// q := regexp.MustCompile(`<@U[0-9A-Za-z]+>`).ReplaceAllString(ev.Text, "")
		// a := tbls.Ask(q)

		// _, _, err := h.Api.PostMessage(
		// 	ev.Channel,
		// 	slack.MsgOptionText(a, false),
		// 	slack.MsgOptionTS(ev.TimeStamp),
		// )
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown event")
	}
	return nil
}
