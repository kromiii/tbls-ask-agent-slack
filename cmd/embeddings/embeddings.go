package embeddings

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/k1LoW/tbls-ask/schema"
	tblsschema "github.com/k1LoW/tbls/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Schemas []struct {
		Name string `yaml:"name"`
		Path string `yaml:"path"`
	} `yaml:"schemas"`
}

type TableVector struct {
	SchemaName string
	TableName  string
	Vector     []float32
}

func Run() {
	config, err := readConfig("./schemas/config.yml")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	db, err := initializeDB()
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}
	defer db.Close()

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
	db, err := sql.Open("sqlite3", "vectors-db/vectors.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`DROP TABLE IF EXISTS table_vectors`)
	if err != nil {
		return nil, err
	}

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
	s, err := schema.Load(schemaPath, schema.Options{})
	if err != nil {
		return fmt.Errorf("failed to load schema: %v", err)
	}

	var descriptions []string
	var tableNames []string

	for _, table := range s.Tables {
		description := createTableDescription(table)
		descriptions = append(descriptions, description)
		tableNames = append(tableNames, table.Name)
	}

	vectors, err := getEmbeddings(descriptions, client)
	if err != nil {
		return err
	}

	for i, vector := range vectors {
		if err := storeVector(db, schemaName, tableNames[i], vector); err != nil {
			return err
		}
	}

	return nil
}

func createTableDescription(table *tblsschema.Table) string {
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

func getEmbeddings(texts []string, client *openai.Client) ([][]float32, error) {
	resp, err := client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: texts,
			Model: openai.AdaEmbeddingV2,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings received")
	}

	var embeddings [][]float32
	for _, data := range resp.Data {
		embeddings = append(embeddings, data.Embedding)
	}

	return embeddings, nil
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
