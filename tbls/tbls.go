package tbls

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/datasource"
	"github.com/sashabaranov/go-openai"
	"github.com/slack-go/slack"
)

var (
	model = "gpt-4-turbo"
)

var analyze = datasource.Analyze
var botUserID = "U06JCJX67GC"

func Ask(messages []slack.Message, path string) string {
	if os.Getenv("OPENAI_API_KEY") == "" {
		return "OPENAI_API is not set"
	}
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()
	dsn := config.DSN{URL: path}
	_, err := analyze(dsn)
	if err != nil {
		log.Printf("Failed to analyze schema: %v", err)
		return "Failed to analyze schema"
	}

	m := []openai.ChatCompletionMessage{}
	for _, message := range messages {
		// skip messages which does not include the mention to the bot or the message not from bot user
		if message.User != botUserID && !strings.Contains(message.Text, "<@"+botUserID+">") {
			continue
		}
		// Role "user" is used for messages from the user
		// Role "assistant" is used for messages from the bot (assistant)
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
		Model:       model,
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
