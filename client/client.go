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

// SchemaInfo represents a schema with name and path
type SchemaInfo struct {
	Name string
	Path string
}

func Ask(messages []slack.Message, name string, path string, botUserID string, model string) string {
	return AskWithSchemas(messages, []*SchemaInfo{{Name: name, Path: path}}, botUserID, model)
}

// AskWithSchemas handles multiple schemas, providing all schema contexts with DB names
func AskWithSchemas(messages []slack.Message, schemas []*SchemaInfo, botUserID string, model string) string {
	if len(messages) == 0 {
		return "No messages found"
	}

	if len(schemas) == 0 {
		return "No schemas provided"
	}

	schemaPrompt, systemPrompt, err := buildSchemaPrompt(schemas)
	if err != nil {
		return err.Error()
	}

	service, err := chat.NewService(model)
	if err != nil {
		return fmt.Sprintf("Failed to create chat service: %v", err)
	}

	m := buildChatMessages(systemPrompt, schemaPrompt, messages, botUserID)

	debugLogMessages(m)

	ctx := context.Background()
	answer, err := service.Ask(ctx, m, false)
	if err != nil {
		return fmt.Sprintf("Failed to ask: %v", err)
	}
	return answer
}

func buildSchemaPrompt(schemas []*SchemaInfo) (schemaPrompt string, systemPrompt string, err error) {
	if len(schemas) == 1 {
		loadedSchema, err := schema.Load(schemas[0].Path, schema.Options{})
		if err != nil {
			return "", "", fmt.Errorf("Failed to load schema: %s", err.Error())
		}

		schemaPrompt, err := prompt.Generate(loadedSchema)
		if err != nil {
			return "", "", fmt.Errorf("Failed to generate schema prompt: %v", err)
		}

		return schemaPrompt, "You are a database expert. You are given a database schema with chat histories. Answer the users' question based on the following schema.", nil
	}

	var schemaPrompts []string
	for _, s := range schemas {
		loadedSchema, err := schema.Load(s.Path, schema.Options{})
		if err != nil {
			return "", "", fmt.Errorf("Failed to load schema %s: %v", s.Name, err)
		}

		sp, err := prompt.Generate(loadedSchema)
		if err != nil {
			return "", "", fmt.Errorf("Failed to generate schema prompt for %s: %v", s.Name, err)
		}

		schemaPrompts = append(schemaPrompts, fmt.Sprintf("=== Database: %s ===\n%s", s.Name, sp))
	}

	return strings.Join(schemaPrompts, "\n\n"),
		"You are a database expert. You are given multiple database schemas with chat histories. Each schema is labeled with its database name. Answer the users' question based on the following schemas.",
		nil
}

func buildChatMessages(systemPrompt, schemaPrompt string, messages []slack.Message, botUserID string) []chat.Message {
	m := []chat.Message{
		{Role: "user", Content: systemPrompt},
		{Role: "user", Content: schemaPrompt},
	}

	if customInstruction := os.Getenv("CUSTOM_INSTRUCTION"); customInstruction != "" {
		m = append(m, chat.Message{Role: "user", Content: customInstruction})
	}

	for _, message := range messages {
		if strings.Contains(message.Text, "Please select the target schema") || strings.Contains(message.Text, "Selected schema:") {
			continue
		}
		var role string
		if message.User == botUserID {
			role = "assistant"
		} else {
			role = "user"
			message.Text = strings.ReplaceAll(message.Text, "<@"+botUserID+">", "@bot")
		}
		m = append(m, chat.Message{Role: role, Content: message.Text})
	}

	return m
}

func debugLogMessages(m []chat.Message) {
	if os.Getenv("DEBUG_MODE") == "true" {
		log.Println("=== Debug: Prompt contents ===")
		for _, msg := range m {
			log.Printf("Role: %s\nContent: %s\n", msg.Role, msg.Content)
		}
		log.Println("============================")
	}
}
