package tbls

import (
	"bytes"
	"context"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/k1LoW/tbls-ask/templates"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/datasource"
	"github.com/sashabaranov/go-openai"
	"github.com/slack-go/slack"
)

const (
	model      = "gpt-4-turbo"
	quoteStart = "```sql"
	quoteEnd   = "```"
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
	s, err := analyze(dsn)
	if err != nil {
		log.Printf("Failed to analyze schema: %v", err)
		return "Failed to analyze schema"
	}

	m := []openai.ChatCompletionMessage{}
	tpl, err := template.New("").Parse(DefaultPromtTmpl)
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		return "Failed to ask"
	}
	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, map[string]any{
		"DatabaseVersion": templates.DatabaseVersion(s),
		"QuoteStart":      quoteStart,
		"QuoteEnd":        quoteEnd,
		"DDL":             templates.GenerateDDLRoughly(s),
	}); err != nil {
		log.Printf("Failed to execute template: %v", err)
		return "Failed to ask"
	}
	m = append(m, openai.ChatCompletionMessage{
		Role:    "system",
		Content: buf.String(),
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
