package search

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/k1LoW/tbls/schema"
	"github.com/sashabaranov/go-openai"
)

const (
	RelevantTablesPromptTmpl = `You are an AI assistant specializing in database analysis. Given the following list of tables and their descriptions (if available):
%s
Please analyze the user's query:
"%s"
Identify and list only the table names that are most relevant to answering this query. Do not include any explanations or additional text. Only provide a comma-separated list of table names.
`
)

func ExtractRelevantTables(ctx context.Context, s *schema.Schema, query string) ([]string, error) {
	var tableInfo string
	tableNames := make([]string, 0, len(s.Tables))
	for _, t := range s.Tables {
		if t.Type == "VIEW" || t.Type == "MATERIALIZED VIEW" {
			continue
		}
		tableNames = append(tableNames, t.Name)
		if t.Comment != "" {
			tableInfo += fmt.Sprintf("%s: %s\n", t.Name, t.Comment)
		} else {
			tableInfo += fmt.Sprintf("%s\n", t.Name)
		}
	}

	prompt := fmt.Sprintf(RelevantTablesPromptTmpl, tableInfo, query)

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o",
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "user",
				Content: prompt,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	answer := resp.Choices[0].Message.Content

	if answer == "" {
		return nil, fmt.Errorf("failed to extract relevant tables: empty response")
	}

	relevantTables := strings.Split(answer, ",")
	for i, t := range relevantTables {
		relevantTables[i] = strings.TrimSpace(t)
		found := false
		for _, tableName := range tableNames {
			if tableName == relevantTables[i] {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("failed to extract relevant tables: %s not found", relevantTables[i])
		}
	}

	return relevantTables, nil
}
