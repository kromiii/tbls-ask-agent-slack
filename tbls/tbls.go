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

var o = openai.New(os.Getenv("OPENAI_API_KEY"), model)
var analyze = datasource.Analyze
var ask = o.Ask

func Ask(query string, path string) string {
	if os.Getenv("OPENAI_API_KEY") == "" {
		return "OPENAI_API is not set"
	}
	ctx := context.Background()
	dsn := config.DSN{URL: path}
	s, err := analyze(dsn)
	if err != nil {
		return "Failed to analyze schema"
	}
	answer, err = ask(ctx, query, s)
	if err != nil {
		return "Failed to ask"
	}
	return answer
}
