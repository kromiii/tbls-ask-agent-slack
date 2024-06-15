package tbls

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
	"github.com/sashabaranov/go-openai"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

type MockClient struct {
	*openai.Client
}

func (m *MockClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	return &openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Content: "mock answer",
				},
			},
		},
	}, nil
}

func TestAsk(t *testing.T) {
	t.Run("OPENAI_API_KEY not set", func(t *testing.T) {
		os.Setenv("OPENAI_API_KEY", "")
		assert.Equal(t, "OPENAI_API is not set", Ask([]slack.Message{}, "path"))
	})

	t.Run("Failed to analyze schema", func(t *testing.T) {
		os.Setenv("OPENAI_API_KEY", "mock_key")
		analyze = func(dsn config.DSN) (*schema.Schema, error) {
			return nil, errors.New("mock error")
		}
		assert.Equal(t, "Failed to analyze schema", Ask([]slack.Message{}, "path"))
	})
}
