package paper

import (
	"context"
	"database/sql"
	"fmt"
	"main/lib/logger"
	"os"
	"strings"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	globalDB     *sql.DB
	dbOnce       sync.Once
	dbInitErr    error
	dbInitialized bool
)

// IsVectorDBEnabled checks if vector DB is enabled via environment variable
func IsVectorDBEnabled() bool {
	enabled := os.Getenv("VECTOR_DB_ENABLED")
	if enabled == "false" || enabled == "0" {
		return false
	}
	return true // Default to enabled
}

// InitDB initializes the database connection
func InitDB() error {
	dbOnce.Do(func() {
		// Check if vector DB is enabled
		if !IsVectorDBEnabled() {
			dbInitErr = fmt.Errorf("vector DB disabled via VECTOR_DB_ENABLED environment variable")
			return
		}
		
		// Get connection string - always use VECTOR_DB_ prefix
		connStr := os.Getenv("VECTOR_DB_DATABASE_URL")
		
		// If no full connection string, try to construct from individual parts
		if connStr == "" {
			// Try unpooled first (for serverless), then regular
			host := os.Getenv("VECTOR_DB_PGHOST_UNPOOLED")
			if host == "" {
				host = os.Getenv("VECTOR_DB_PGHOST")
			}
			
			port := os.Getenv("VECTOR_DB_PGPORT")
			user := os.Getenv("VECTOR_DB_PGUSER")
			password := os.Getenv("VECTOR_DB_PGPASSWORD")
			database := os.Getenv("VECTOR_DB_PGDATABASE")
			
			if host != "" && user != "" && password != "" && database != "" {
				if port == "" {
					port = "5432"
				}
				connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require", user, password, host, port, database)
			}
		}
		if connStr == "" {
			dbInitErr = fmt.Errorf("VECTOR_DB_DATABASE_URL or VECTOR_DB_PG* environment variables are not set")
			return
		}

		// Configure connection pool for serverless (avoid prepared statement conflicts)
		// Add connection parameters to disable prepared statements for serverless compatibility
		if !strings.Contains(connStr, "prefer_simple_protocol") {
			if strings.Contains(connStr, "?") {
				connStr += "&prefer_simple_protocol=1"
			} else {
				connStr += "?prefer_simple_protocol=1"
			}
		}

		db, err := sql.Open("pgx", connStr)
		if err != nil {
			dbInitErr = fmt.Errorf("failed to open database connection: %w", err)
			logger.Error("Failed to open database connection", err, nil)
			return
		}

		// Configure connection pool settings for serverless
		db.SetMaxOpenConns(1)  // Single connection for serverless
		db.SetMaxIdleConns(1)  // Single idle connection
		db.SetConnMaxLifetime(0) // Don't close connections based on time

		// Test connection
		if err := db.Ping(); err != nil {
			dbInitErr = fmt.Errorf("failed to ping database: %w", err)
			logger.Error("Failed to ping database", err, nil)
			_ = db.Close()
			return
		}

		globalDB = db
		dbInitialized = true
		
		ctx := context.Background()
		_ = InitSchema(ctx)
	})

	return dbInitErr
}

// GetDB returns the global database connection
// Returns nil if database is not initialized or unavailable
func GetDB() *sql.DB {
	if !dbInitialized {
		return nil
	}
	return globalDB
}

// CloseDB closes the database connection
func CloseDB() error {
	if globalDB != nil {
		return globalDB.Close()
	}
	return nil
}

// IsDBEnabled returns true if database is available
func IsDBEnabled() bool {
	return dbInitialized && globalDB != nil
}

