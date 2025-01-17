package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"

	"github.com/k1LoW/tbls-ask/chat"
	"github.com/k1LoW/tbls-ask/prompt"
	"github.com/k1LoW/tbls-ask/schema"
)

func Ask(messages []slack.Message, name string, path string, botUserID string, model string) string {
	ctx := context.Background()

	if len(messages) == 0 {
		return "No messages found"
	}

	schema, err := schema.Load(path, schema.Options{})
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
			Content: "You are a database expert. You are given a database schema with chat histories. Answer the users' question based on the following schema.",
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
		if strings.Contains(message.Text, "Please select the target schema") {
			continue
		}
		var role string
		if message.User == botUserID {
			role = "assistant"
		} else {
			role = "user"
			message.Text = strings.ReplaceAll(message.Text, "<@"+botUserID+">", "@bot")
		}
		m = append(m, chat.Message{
			Role:    role,
			Content: message.Text,
		})
	}

	if os.Getenv("DEBUG_MODE") == "true" {
		log.Println("=== Debug: Prompt contents ===")
		for _, msg := range m {
			log.Printf("Role: %s\nContent: %s\n", msg.Role, msg.Content)
		}
		log.Println("============================")
	}

	answer, err := service.Ask(ctx, m, false)
	if err != nil {
		return fmt.Sprintf("Failed to ask: %v", err)
	}
	return answer
}
