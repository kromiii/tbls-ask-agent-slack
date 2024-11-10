package search

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/sashabaranov/go-openai"
)

// SearchResult represents a single table search result with similarity score
type SearchResult struct {
	SchemaName string
	TableName  string
	Score      float64
}

// TableSearcher handles the table search functionality
type TableSearcher struct {
	db     *sql.DB
	client *openai.Client
}

// NewTableSearcher creates a new instance of TableSearcher
func NewTableSearcher(db *sql.DB, openaiKey string) *TableSearcher {
	client := openai.NewClient(openaiKey)
	return &TableSearcher{
		db:     db,
		client: client,
	}
}

// SearchTables searches for relevant tables based on the query
// limit specifies the maximum number of results to return
// minScore specifies the minimum similarity score (0-1) for results
func (ts *TableSearcher) SearchTables(ctx context.Context, query string, limit int, minScore float64) ([]SearchResult, error) {
	// Get query embedding
	queryVector, err := ts.getQueryEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	// Get all table vectors from database
	tableVectors, err := ts.getAllTableVectors()
	if err != nil {
		return nil, fmt.Errorf("failed to get table vectors: %w", err)
	}

	// Calculate similarities and sort results
	var results []SearchResult
	for _, tv := range tableVectors {
		score := cosineSimilarity(queryVector, tv.Vector)
		if score >= minScore {
			results = append(results, SearchResult{
				SchemaName: tv.SchemaName,
				TableName:  tv.TableName,
				Score:      score,
			})
		}
	}

	// Sort results by score in descending order
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

type tableVector struct {
	SchemaName string
	TableName  string
	Vector     []float32
}

func (ts *TableSearcher) getQueryEmbedding(ctx context.Context, query string) ([]float32, error) {
	resp, err := ts.client.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Input: []string{query},
			Model: openai.AdaEmbeddingV2,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding received")
	}

	return resp.Data[0].Embedding, nil
}

func (ts *TableSearcher) getAllTableVectors() ([]tableVector, error) {
	rows, err := ts.db.Query(`
		SELECT schema_name, table_name, vector 
		FROM table_vectors
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vectors []tableVector
	for rows.Next() {
		var tv tableVector
		var vectorBytes []byte
		if err := rows.Scan(&tv.SchemaName, &tv.TableName, &vectorBytes); err != nil {
			return nil, err
		}
		
		if err := json.Unmarshal(vectorBytes, &tv.Vector); err != nil {
			return nil, err
		}
		vectors = append(vectors, tv)
	}

	return vectors, rows.Err()
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct float64
	var normA float64
	var normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
