package tbls

import (
	"context"
	"os"

	"github.com/k1LoW/tbls-ask/openai"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/datasource"
)

var (
	model  = "gpt-4-1106-preview"
	answer string
)

func Ask(query string, path string) string {
	if os.Getenv("OPENAI_API_KEY") == "" {
		return "OPENAI_API is not set"
	}
	ctx := context.Background()
	o := openai.New(os.Getenv("OPENAI_API_KEY"), model)
	dsn := config.DSN{URL: path}
	s, err := datasource.Analyze(dsn)
	if err != nil {
		return "Failed to analyze schema"
	}
	answer, err = o.Ask(ctx, query, s)
	if err != nil {
		return "Failed to ask"
	}
	return answer
}
