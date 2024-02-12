package tbls

import (
	"context"
	"os"

	"github.com/k1LoW/tbls-ask/openai"
	"github.com/k1LoW/tbls/datasource"
)

var (
	model  = "gpt-4-1106-preview"
	answer string
)

func Ask(query string) string {
	if os.Getenv("OPENAI_API_KEY") == "" {
		return "OPENAI_API is not set"
	}
	if os.Getenv("TBLS_SCHEMA") == "" {
		return "TBLS_SCHEMA is not set"
	}
	ctx := context.Background()
	o := openai.New(os.Getenv("OPENAI_API_KEY"), model)
	s, err := datasource.AnalyzeJSONStringOrFile(os.Getenv("TBLS_SCHEMA"))
	if err != nil {
		return "Failed to analyze schema"
	}
	answer, err = o.Ask(ctx, query, s)
	if err != nil {
		return "Failed to ask"
	}
	return answer
}
