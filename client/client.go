package client

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/slack-go/slack"

	"github.com/k1LoW/tbls-ask/chat"
	"github.com/k1LoW/tbls-ask/prompt"
	"github.com/k1LoW/tbls-ask/schema"
	"github.com/kromiii/tbls-ask-agent-slack/search"
)

const (
	distance = 2
)

func Ask(messages []slack.Message, name string, path string, botUserID string, model string, searchMode bool) string {
	ctx := context.Background()

	if len(messages) == 0 {
		return "No messages found"
	}

	var includes []string
	var selectedTablesFeedback string

	if searchMode {
		entireSchema, err := schema.Load(path, schema.Options{})
		if err != nil {
			return "Failed to load schema: " + err.Error()
		}

		query := messages[len(messages)-1].Text

		results, err := search.ExtractRelevantTables(
			context.Background(),
			entireSchema,
			query,
		)
		if err == nil {
			includes = make([]string, len(results))
			for i, result := range results {
				includes[i] = result
				selectedTablesFeedback = "Relevant tables: " + strings.Join(includes, ", ")
			}
		}
	}

	schema, err := schema.Load(path, schema.Options{
		Includes: includes,
		Distance: distance,
	})
	if err != nil {
		return "Failed to load schema: " + err.Error()
	}

	schemaPrompt, err := prompt.Generate(schema)
	if err != nil {
		return fmt.Sprintf("Failed to generate schema prompt: %v", err)
	}

	service, err := chat.NewService(model)
	if err != nil {
		return fmt.Sprintf("Failed to create chat service: %v", err)
	}

	customInstruction := os.Getenv("CUSTOM_INSTRUCTION")

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

	if customInstruction != "" {
		m = append(m, chat.Message{
			Role:    "system",
			Content: customInstruction,
		})
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
	if selectedTablesFeedback != "" {
		answer = selectedTablesFeedback + "\n\n" + answer
	}
	return answer
}
