package tbls

import (
	"bytes"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/datasource"
	"github.com/slack-go/slack"

	"github.com/kromiii/tbls-ask-agent-slack/openai"
)

const (
	model      = "gpt-4-turbo"
	quoteStart = "```sql"
	quoteEnd   = "```"
)

func Ask(messages []slack.Message, path string, botUserID string) string {
	dsn := config.DSN{URL: path}
	s, err := datasource.Analyze(dsn)
	if err != nil {
		log.Printf("Failed to analyze schema: %v", err)
		return "Failed to analyze schema"
	}

	tpl, err := template.New("").Parse(DefaultPromtTmpl)
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		return "Failed to ask"
	}

	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, map[string]any{
		"DatabaseVersion": DatabaseVersion(s),
		"QuoteStart":      quoteStart,
		"QuoteEnd":        quoteEnd,
		"DDL":             GenerateDDLRoughly(s),
	}); err != nil {
		log.Printf("Failed to execute template: %v", err)
		return "Failed to ask"
	}

	chatMessages := make([]openai.ChatMessage, 0)
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
		chatMessages = append(chatMessages, openai.ChatMessage{
			Role:    role,
			Content: message.Text,
		})
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	answer, err := client.ChatCompletion(chatMessages)
	if err != nil {
		log.Printf("Failed to ask: %v", err)
		return "Failed to ask"
	}
	return answer
}
