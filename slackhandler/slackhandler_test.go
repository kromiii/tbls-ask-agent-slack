package slackhandler

import (
	"errors"
	"os"
	"testing"

	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestHandleCallBackEvent(t *testing.T) {
	t.Run("AppMentionEvent", func(t *testing.T) {
		h := &SlackHandler{
			Api: &slack.Client{},
		}

		event := slackevents.EventsAPIEvent{
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{},
			},
		}

		// Mock the ReadFile function to return a predefined config
		os.ReadFile = func(name string) ([]byte, error) {
			config := Config{
				Schemas: []Schema{
					{Name: "Test", Path: "/test/path"},
				},
			}
			data, _ := yaml.Marshal(config)
			return data, nil
		}

		err := h.HandleCallBackEvent(event)
		assert.Nil(t, err)
	})

	t.Run("UnknownEvent", func(t *testing.T) {
		h := &SlackHandler{
			Api: &slack.Client{},
		}

		event := slackevents.EventsAPIEvent{
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.MessageAction{},
			},
		}

		err := h.HandleCallBackEvent(event)
		assert.Equal(t, errors.New("unknown event"), err)
	})
}
