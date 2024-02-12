package slackhandler

import (
	"errors"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type SlackHandler struct {
	Api *slack.Client
}

func (h *SlackHandler) HandleCallBackEvent(event slackevents.EventsAPIEvent) error {
	innerEvent := event.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		_, _, err := h.Api.PostMessage(ev.Channel, slack.MsgOptionBlocks(
			slack.SectionBlock{
				Type: slack.MBTSection,
				Text: &slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: "どの都市の天気を調べますか？",
				},
				Accessory: &slack.Accessory{
					SelectElement: &slack.SelectBlockElement{
						ActionID: "select_city",
						Type:     slack.OptTypeStatic,
						Placeholder: &slack.TextBlockObject{
							Type: slack.PlainTextType,
							Text: "都市を選択",
						},
						Options: []*slack.OptionBlockObject{
							{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "東京"}, Value: "東京"},
							{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "ソウル"}, Value: "ソウル"},
							{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "北京"}, Value: "北京"},
							{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "シドニー"}, Value: "シドニー"},
							{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "パリ"}, Value: "パリ"},
							{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "ロンドン"}, Value: "ロンドン"},
							{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "ベルリン"}, Value: "ベルリン"},
							{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "ニューヨーク"}, Value: "ニューヨーク"},
							{Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "ロサンゼルス"}, Value: "ロサンゼルス"},
						},
					},
				},
			},
		))
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
	case "select_city":
		weather, err := weather.GetWeather(action.SelectedOption.Value)
		if err != nil {
			return err
		}
		_, _, _, err = h.Api.SendMessage(
			"",
			slack.MsgOptionReplaceOriginal(interaction.ResponseURL),
			slack.MsgOptionBlocks(
				slack.SectionBlock{
					Type: slack.MBTSection,
					Text: &slack.TextBlockObject{
						Type: slack.MarkdownType,
						Text: "どの都市の天気を調べますか？: " + action.SelectedOption.Text.Text,
					},
				},
				slack.SectionBlock{
					Type: slack.MBTSection,
					Text: &slack.TextBlockObject{
						Type: slack.MarkdownType,
						Text: "```\n" + weather + "```",
					},
				},
			),
		)
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown action")
	}
	return nil
}
