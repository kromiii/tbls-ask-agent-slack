package slackhandler

import (
	"errors"
	"testing"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

type TestSlackHandler struct {
	Api MockSlackClient
}

func (h *TestSlackHandler) HandleCallBackEvent(event slackevents.EventsAPIEvent) error {
	// メソッドの実装をここに書く
	return nil
}

type MockSlackClient struct {
	slack.Client
}

func (m *MockSlackClient) PostMessage(channel string, options ...slack.MsgOption) (string, string, error) {
	return "", "", nil
}

func TestHandleCallBackEvent(t *testing.T) {
	t.Run("AppMentionEvent", func(t *testing.T) {
		h := &TestSlackHandler{
			Api: MockSlackClient{},
		}

		event := slackevents.EventsAPIEvent{
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.AppMentionEvent{},
			},
		}

		// Mock the ReadFile function to return a predefined config
		fileLoader = func(_ string) ([]byte, error) {
			config := Config{
				Schemas: []Schema{
					{Name: "Test", Path: "./config.yml.sample"},
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
