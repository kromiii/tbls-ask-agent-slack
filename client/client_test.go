package client

import (
	"os"
	"strings"
	"testing"

	"github.com/slack-go/slack"
)

func TestAsk(t *testing.T) {
	messages := []slack.Message{
		{
			Msg: slack.Msg{
				User: "U123456",
				Text: "What tables are in the database?",
			},
		},
	}

	name := "test_db"
	path := "testdata/schema.json"
	botUserID := "UBOTID123"
	model := "gpt-3.5-turbo"

	_ = os.Setenv("OPENAI_API_KEY", "your-api-key")

	result := Ask(messages, name, path, botUserID, model)

	if result == "" {
		t.Error("Expected non-empty result, got empty string")
	}

	if result == "No messages found" {
		t.Error("Unexpected 'No messages found' result")
	}

	if result == "Failed to load schema: " {
		t.Error("Failed to load schema")
	}
}

func TestAskWithNoMessages(t *testing.T) {
	messages := []slack.Message{}
	name := "test_db"
	path := "testdata/schema.json"
	botUserID := "UBOTID123"
	model := "gpt-3.5-turbo"

	result := Ask(messages, name, path, botUserID, model)

	if result != "No messages found" {
		t.Errorf("Expected 'No messages found', got: %s", result)
	}
}

func TestAskWithInvalidSchema(t *testing.T) {
	messages := []slack.Message{
		{
			Msg: slack.Msg{
				User: "U123456",
				Text: "What tables are in the database?",
			},
		},
	}
	name := "test_db"
	path := "nonexistent/schema.json"
	botUserID := "UBOTID123"
	model := "gpt-3.5-turbo"

	result := Ask(messages, name, path, botUserID, model)

	expectedPrefix := "Failed to load schema: failed to analyze schema: "
	if !strings.HasPrefix(result, expectedPrefix) {
		t.Errorf("Expected result to start with '%s', got: %s", expectedPrefix, result)
	}
}
