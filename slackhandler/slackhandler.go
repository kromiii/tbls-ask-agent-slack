package slackhandler

import (
	"errors"
	"regexp"

	"github.com/kromiii/tbls-ask-server/tbls"
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
		h.Api.PostMessage(
			ev.Channel,
			slack.MsgOptionText("ちょっと待ってね（考え中）", false),
			slack.MsgOptionTS(ev.TimeStamp),
		)

		q := regexp.MustCompile(`<@U[0-9A-Za-z]+>`).ReplaceAllString(ev.Text, "")
		a := tbls.Ask(q)

		_, _, err := h.Api.PostMessage(
			ev.Channel,
			slack.MsgOptionText(a, false),
			slack.MsgOptionTS(ev.TimeStamp),
		)
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown event")
	}
	return nil
}
