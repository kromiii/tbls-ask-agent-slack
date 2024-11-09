package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/slack-go/slack"

	"github.com/k1LoW/tbls-ask/chat"
	"github.com/k1LoW/tbls-ask/prompt"
	"github.com/k1LoW/tbls-ask/schema"
)

func Ask(messages []slack.Message, path string, botUserID string, model string) string {
	ctx := context.Background()
	schema, err := schema.Load(path, schema.Options{})
	if err != nil {
		return fmt.Sprintf("Failed to load schema: %v", err)
	}
	schemaPrompt, err := prompt.Generate(schema)
	if err != nil {
		return fmt.Sprintf("Failed to generate schema prompt: %v", err)
	}

	service, err := chat.NewService(model)
	if err != nil {
		return fmt.Sprintf("Failed to create chat service: %v", err)
	}


	m := []chat.Message{
		{
			Role:    "system",
			Content: "You are a database expert. You are given a database schema and a question. Answer the question based on the schema.",
		},
		{
			Role:    "user",
			Content: schemaPrompt,
		},
	}
	
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
		m = append(m, chat.Message{
			Role:    role,
			Content: message.Text,
		})
	}

	answer, err := service.Ask(ctx, m, false)
	if err != nil {
		return fmt.Sprintf("Failed to ask: %v", err)
	}
	return answer
}
