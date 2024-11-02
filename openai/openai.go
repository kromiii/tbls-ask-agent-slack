package openai

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/slack-go/slack"

	"github.com/k1LoW/tbls-ask/analyzer"
)

func Ask(messages []slack.Message, path string, botUserID string) string {
	if os.Getenv("OPENAI_API_KEY") == "" {
		return "OPENAI_API is not set"
	}
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()

	var a analyzer.Analyzer
	err := a.AnalyzeSchema(path, nil, nil, nil)
	if err != nil {
		return "Failed to analyze schema"
	}
	p, err := a.GeneratePrompt("Answer to the users' question based on the following chat history", false)
	if err != nil {
		return "Failed to generate prompt"
	}

	m := []openai.ChatCompletionMessage{}
	m = append(m, openai.ChatCompletionMessage{
		Role:    "user",
		Content: p,
	})
	for _, message := range messages {
		// skip messages which does not include the mention to the bot or the message not from bot user
		if message.User != botUserID && !strings.Contains(message.Text, "<@"+botUserID+">") {
			continue
		}
		var role string
		if message.User == botUserID {
			role = "assistant"
		} else {
			role = "user"
		}
		m = append(m, openai.ChatCompletionMessage{
			Role:    role,
			Content: message.Text,
		})
	}

	res, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       "gpt-4o",
		Temperature: 0.2, // https://community.openai.com/t/cheat-sheet-mastering-temperature-and-top-p-in-chatgpt-api-a-few-tips-and-tricks-on-controlling-the-creativity-deterministic-output-of-prompt-responses/172683
		Messages:    m,
	})
	if err != nil {
		log.Printf("Failed to ask: %v", err)
		return "Failed to ask"
	}
	answer := res.Choices[0].Message.Content
	return answer
}
