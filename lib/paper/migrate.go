package paper

import (
	"context"
	"fmt"
	"main/lib/logger"
	"strings"
	"sync"
)

var (
	schemaInitialized bool
	schemaInitOnce    sync.Once
	schemaInitErr     error
)

// InitSchema initializes the database schema (tables, indexes, extensions)
// This should be called once when the database is first set up
// Uses sync.Once to ensure it only runs once per process
func InitSchema(ctx context.Context) error {
	schemaInitOnce.Do(func() {
		db := GetDB()
		if db == nil {
			schemaInitErr = fmt.Errorf("database not initialized")
			return
		}

		// Quick check if schema already exists (faster than running all CREATE IF NOT EXISTS)
		var exists int
		err := db.QueryRowContext(ctx, `
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'result_embeddings' 
			LIMIT 1
		`).Scan(&exists)
		
		if err == nil {
			schemaInitialized = true
			return
		}

		// Execute schema statements individually
		statements := []string{
			"CREATE EXTENSION IF NOT EXISTS vector",
			`CREATE TABLE IF NOT EXISTS query_embeddings (
				id SERIAL PRIMARY KEY,
				query_hash TEXT UNIQUE NOT NULL,
				query_text TEXT,
				embedding VECTOR(512) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			`CREATE TABLE IF NOT EXISTS result_embeddings (
				id SERIAL PRIMARY KEY,
				paper_id TEXT NOT NULL UNIQUE,
				embedding VECTOR(512) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			"CREATE INDEX IF NOT EXISTS query_embeddings_embedding_idx ON query_embeddings USING hnsw (embedding vector_ip_ops)",
			"CREATE INDEX IF NOT EXISTS result_embeddings_embedding_idx ON result_embeddings USING hnsw (embedding vector_ip_ops)",
			"CREATE INDEX IF NOT EXISTS result_embeddings_paper_id_idx ON result_embeddings (paper_id)",
		}

		for _, stmt := range statements {
			_, err := db.ExecContext(ctx, stmt)
			if err != nil {
				stmtPreview := strings.TrimSpace(stmt)
				if len(stmtPreview) > 100 {
					stmtPreview = stmtPreview[:100]
				}
				logger.Error("Failed to execute schema statement", err, map[string]interface{}{
					"statement": stmtPreview,
				})
				schemaInitErr = fmt.Errorf("failed to execute schema: %w", err)
				return
			}
		}

		schemaInitialized = true
	})

	return schemaInitErr
}


