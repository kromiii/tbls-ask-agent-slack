package openai

import (
    "context"
    "errors"

    "github.com/sashabaranov/go-openai"
)

type ChatMessage struct {
    Role    string
    Content string
}

type Client struct {
    apiClient *openai.Client
}

func NewClient(apiKey string) *Client {
    return &Client{
        apiClient: openai.NewClient(apiKey),
    }
}

func (c *Client) ChatCompletion(messages []ChatMessage) (string, error) {
    if len(messages) == 0 {
        return "", errors.New("messages array is empty")
    }

    openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
    for i, msg := range messages {
        openaiMessages[i] = openai.ChatCompletionMessage{
            Role:    msg.Role,
            Content: msg.Content,
        }
    }

    resp, err := c.apiClient.CreateChatCompletion(
        context.Background(),
        openai.ChatCompletionRequest{
            Model:    openai.GPT3Dot5Turbo,
            Messages: openaiMessages,
        },
    )

    if err != nil {
        return "", err
    }

    if len(resp.Choices) == 0 {
        return "", errors.New("no response from the model")
    }

    return resp.Choices[0].Message.Content, nil
}
