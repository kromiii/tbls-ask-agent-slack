package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/k1LoW/tbls/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v2"
)

// Config represents the YAML configuration structure
type Config struct {
	Schemas []struct {
		Name string `yaml:"name"`
		Path string `yaml:"path"`
	} `yaml:"schemas"`
}

// TableVector represents the database structure for storing vectors
type TableVector struct {
	SchemaName string
	TableName  string
	Vector     []float32
}

func main() {
	// Step 1: Read configuration file
	config, err := readConfig("./schemas/config.yml")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	// Step 2: Initialize OpenAI client
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	// Step 3: Initialize SQLite database
	db, err := initializeDB()
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}
	defer db.Close()

	// Step 4: Process each schema
	for _, schema := range config.Schemas {
		err := processSchema(schema.Name, schema.Path, client, db)
		if err != nil {
			fmt.Printf("Error processing schema %s: %v\n", schema.Name, err)
			continue
		}
	}
}

func readConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func initializeDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "vectors.db")
	if err != nil {
		return nil, err
	}

	// Create table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS table_vectors (
			schema_name TEXT,
			table_name TEXT,
			vector BLOB,
			PRIMARY KEY (schema_name, table_name)
		)
	`)
	return db, err
}

func processSchema(schemaName, schemaPath string, client *openai.Client, db *sql.DB) error {
	// Fetch schema JSON
	schemaData, err := fetchSchemaJSON(schemaPath)
	if err != nil {
		return err
	}

	// Parse schema using tbls
	var s schema.Schema
	if err := json.Unmarshal(schemaData, &s); err != nil {
		return err
	}

	// Process each table
	for _, table := range s.Tables {
		// Create table description
		description := createTableDescription(table)

		// Get embedding from OpenAI
		vector, err := getEmbedding(description, client)
		if err != nil {
			return err
		}

		// Store in database
		if err := storeVector(db, schemaName, table.Name, vector); err != nil {
			return err
		}
	}

	return nil
}

func fetchSchemaJSON(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func createTableDescription(table *schema.Table) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Table Name: %s\n", table.Name))
	buf.WriteString(fmt.Sprintf("Description: %s\n\n", table.Comment))
	buf.WriteString("Columns:\n")

	for _, column := range table.Columns {
		buf.WriteString(fmt.Sprintf("- %s (%s)", column.Name, column.Type))
		if column.Comment != "" {
			buf.WriteString(fmt.Sprintf(": %s", column.Comment))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

func getEmbedding(text string, client *openai.Client) ([]float32, error) {
	resp, err := client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{text},
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

func storeVector(db *sql.DB, schemaName, tableName string, vector []float32) error {
	vectorBytes, err := json.Marshal(vector)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT OR REPLACE INTO table_vectors (schema_name, table_name, vector)
		VALUES (?, ?, ?)
	`, schemaName, tableName, vectorBytes)

	return err
}
