package client

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/slack-go/slack"

	"github.com/k1LoW/tbls-ask/chat"
	"github.com/k1LoW/tbls-ask/prompt"
	"github.com/k1LoW/tbls-ask/schema"
	"github.com/kromiii/tbls-ask-agent-slack/search"

	_ "github.com/mattn/go-sqlite3"
)

func Ask(messages []slack.Message, name string, path string, botUserID string, model string) string {
	ctx := context.Background()

	// messages の最後の要素を query (string) として取り出す
	query := messages[len(messages)-1].Text
	// vector search
	db, err := sql.Open("sqlite3", "vectors.db")
	if err != nil {
		return "Failed to open database: " + err.Error()
	}
	defer db.Close()

	// TableSearcherの初期化
	searcher := search.NewTableSearcher(db, os.Getenv("OPENAI_API_KEY"))

	// 検索の実行
	results, err := searcher.SearchTables(
		context.Background(),
		name,
		query, // 検索クエリ
		5,     // 上位5件の結果を取得
		0.7,   // 類似度スコアが0.7以上のものを取得
	)
	if err != nil {
		return "Failed to search tables: " + err.Error()
	}

	includes := make([]string, len(results))
	for i, result := range results {
		includes[i] = result.TableName
	}

	// includes の中身を出力しておく
	fmt.Println(includes)

	schema, err := schema.Load(path, schema.Options{
		Includes: includes,
		Distance: 2,
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
