package tbls

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/k1LoW/tbls-ask/openai"
	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
	"github.com/stretchr/testify/assert"
)

type MockOpenAI struct {
	*openai.OpenAI
}

func (m *MockOpenAI) Ask(ctx context.Context, query string, s *schema.Schema) (string, error) {
	return "mock answer", nil
}

type MockDataSource struct {
	schema.Schema
}

type Schema struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

func (m *MockDataSource) Analyze(dsn config.DSN) (*schema.Schema, error) {
	return &schema.Schema{}, nil
}

func TestAsk(t *testing.T) {
	t.Run("OPENAI_API_KEY not set", func(t *testing.T) {
		os.Setenv("OPENAI_API_KEY", "")
		assert.Equal(t, "OPENAI_API is not set", Ask("query", "path"))
	})

	t.Run("Failed to analyze schema", func(t *testing.T) {
		os.Setenv("OPENAI_API_KEY", "mock_key")
		analyze = func(dsn config.DSN) (*schema.Schema, error) {
			return nil, errors.New("mock error")
		}
		assert.Equal(t, "Failed to analyze schema", Ask("query", "path"))
	})

	t.Run("Failed to ask", func(t *testing.T) {
		analyze = func(dsn config.DSN) (*schema.Schema, error) {
			return &schema.Schema{}, nil
		}
		ask = func(ctx context.Context, query string, s *schema.Schema) (string, error) {
			return "", errors.New("mock error")
		}
		assert.Equal(t, "Failed to ask", Ask("query", "path"))
	})

	t.Run("Successful ask", func(t *testing.T) {
		analyze = func(dsn config.DSN) (*schema.Schema, error) {
			return &schema.Schema{}, nil
		}
		ask = func(ctx context.Context, query string, s *schema.Schema) (string, error) {
			return "mock answer", nil
		}
		assert.Equal(t, "mock answer", Ask("query", "path"))
	})
}
