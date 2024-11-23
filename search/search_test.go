package search

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock OpenAI client
type mockOpenAIClient struct {
	mock.Mock
}

func (m *mockOpenAIClient) CreateEmbeddings(ctx context.Context, request openai.EmbeddingRequestConverter) (openai.EmbeddingResponse, error) {
    args := m.Called(ctx, request)
    return args.Get(0).(openai.EmbeddingResponse), args.Error(1)
}

func TestSearchTables(t *testing.T) {
	// Set up mock DB
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Set up mock OpenAI client
	mockClient := new(mockOpenAIClient)

	// Create TableSearcher instance
	ts := &TableSearcher{
		db:     db,
		client: mockClient,
	}

	// Test case
	ctx := context.Background()
	schemaName := "public"
	query := "user information"
	limit := 2
	minScore := 0.5

	// Mock OpenAI client response
	mockClient.On("CreateEmbeddings", mock.Anything, mock.Anything).Return(openai.EmbeddingResponse{
		Data: []openai.Embedding{
			{Embedding: []float32{0.1, 0.2, 0.3}},
		},
	}, nil)

	// Mock DB response
	rows := sqlmock.NewRows([]string{"schema_name", "table_name", "vector"}).
		AddRow("public", "users", mustMarshal([]float32{0.2, 0.3, 0.4})).
		AddRow("public", "products", mustMarshal([]float32{0.3, 0.4, 0.5})).
		AddRow("public", "orders", mustMarshal([]float32{0.4, 0.5, 0.6}))

	dbMock.ExpectQuery("SELECT schema_name, table_name, vector FROM table_vectors").
		WithArgs(schemaName).
		WillReturnRows(rows)

	// Execute the function
	results, err := ts.SearchTables(ctx, schemaName, query, limit, minScore)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "users", results[0].TableName)
	assert.Equal(t, "products", results[1].TableName)
	assert.InDelta(t, 0.9925833209243807, results[0].Score, 0.0001)
	assert.InDelta(t, 0.9827076386442765, results[1].Score, 0.0001)

	// Verify expectations
	assert.NoError(t, dbMock.ExpectationsWereMet())
	mockClient.AssertExpectations(t)
}

func TestGetQueryEmbedding(t *testing.T) {
	mockClient := new(mockOpenAIClient)
	ts := &TableSearcher{client: mockClient}

	ctx := context.Background()
	query := "test query"

	expectedEmbedding := []float32{0.1, 0.2, 0.3}
	mockClient.On("CreateEmbeddings", mock.Anything, mock.Anything).Return(openai.EmbeddingResponse{
		Data: []openai.Embedding{
			{Embedding: expectedEmbedding},
		},
	}, nil)

	result, err := ts.getQueryEmbedding(ctx, query)

	assert.NoError(t, err)
	assert.Equal(t, expectedEmbedding, result)
	mockClient.AssertExpectations(t)
}

func TestGetAllTableVectors(t *testing.T) {
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	ts := &TableSearcher{db: db}

	schemaName := "public"
	rows := sqlmock.NewRows([]string{"schema_name", "table_name", "vector"}).
		AddRow("public", "users", mustMarshal([]float32{0.1, 0.2, 0.3})).
		AddRow("public", "products", mustMarshal([]float32{0.4, 0.5, 0.6}))

	dbMock.ExpectQuery("SELECT schema_name, table_name, vector FROM table_vectors").
		WithArgs(schemaName).
		WillReturnRows(rows)

	result, err := ts.getAllTableVectors(schemaName)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "users", result[0].TableName)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, result[0].Vector)
	assert.Equal(t, "products", result[1].TableName)
	assert.Equal(t, []float32{0.4, 0.5, 0.6}, result[1].Vector)

	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestCosineSimilarity(t *testing.T) {
	testCases := []struct {
		name     string
		a        []float32
		b        []float32
		expected float64
	}{
		{
			name:     "identical vectors",
			a:        []float32{1, 2, 3},
			b:        []float32{1, 2, 3},
			expected: 1,
		},
		{
			name:     "orthogonal vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{0, 1, 0},
			expected: 0,
		},
		{
			name:     "opposite vectors",
			a:        []float32{1, 2, 3},
			b:        []float32{-1, -2, -3},
			expected: -1,
		},
		{
			name:     "different length vectors",
			a:        []float32{1, 2, 3},
			b:        []float32{1, 2},
			expected: 0,
		},
		{
			name:     "zero vector",
			a:        []float32{0, 0, 0},
			b:        []float32{1, 2, 3},
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cosineSimilarity(tc.a, tc.b)
			assert.InDelta(t, tc.expected, result, 0.0001)
		})
	}
}

func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

