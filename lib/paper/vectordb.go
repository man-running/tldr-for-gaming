package paper

import (
	"context"
	"database/sql"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"main/lib/logger"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/takara-ai/serverlessVector"
)

const (
	// Default embedding dimension (512 for the TEI model)
	defaultDimension = 512
)

// VectorDBCache manages vector database for embeddings (CDN handles result caching)
type VectorDBCache struct {
	// In-memory fallback vector DB
	vectorDB *serverlessVector.VectorDB
	
	// Mutex for thread-safe operations
	mu sync.RWMutex
	
	// Dimension for embeddings
	dimension int
	
	// Database enabled flag
	dbEnabled bool
}

// NewVectorDBCache creates a new vector database cache instance
func NewVectorDBCache(dimension int) *VectorDBCache {
	if dimension <= 0 {
		dimension = defaultDimension
	}
	
	// Initialize database connection
	_ = InitDB()
	dbEnabled := IsDBEnabled()
	
	// Check if fallback is enabled (default: true)
	// Use same getEnv pattern as database.go if needed
	fallbackEnabled := os.Getenv("VECTOR_DB_FALLBACK")
	useFallback := fallbackEnabled != "false" && fallbackEnabled != "0"
	
	// Always initialize in-memory fallback if enabled
	var inMemoryDB *serverlessVector.VectorDB
	if useFallback {
		inMemoryDB = serverlessVector.NewVectorDB(dimension, serverlessVector.DotProduct)
	}
	
	return &VectorDBCache{
		vectorDB: inMemoryDB,
		dimension: dimension,
		dbEnabled: dbEnabled,
	}
}

// HashQuery creates a SHA256 hash of the query for cache key
func (v *VectorDBCache) HashQuery(query string) string {
	hash := sha256.Sum256([]byte(query))
	return hex.EncodeToString(hash[:])
}

// float32SliceToVectorString converts []float32 to pgvector string format
func float32SliceToVectorString(embedding []float32) string {
	if len(embedding) == 0 {
		return "[]"
	}
	
	strs := make([]string, len(embedding))
	for i, val := range embedding {
		strs[i] = strconv.FormatFloat(float64(val), 'f', -1, 32)
	}
	return "[" + strings.Join(strs, ",") + "]"
}

// AddEmbedding stores a query embedding in the vector DB for similarity search
// queryText is optional and can be empty string
func (v *VectorDBCache) AddEmbedding(queryHash string, embedding []float32) error {
	return v.AddEmbeddingWithText(queryHash, "", embedding)
}

// GetQueryEmbedding retrieves a query embedding from the database cache
func (v *VectorDBCache) GetQueryEmbedding(ctx context.Context, queryHash string) ([]float32, error) {
	if !v.dbEnabled {
		return nil, nil
	}
	
	db := GetDB()
	if db == nil {
		return nil, nil
	}
	
	var vectorStr string
	err := db.QueryRowContext(
		ctx,
		`SELECT embedding::text FROM query_embeddings WHERE query_hash = $1`,
		queryHash,
	).Scan(&vectorStr)
	
	if err == sql.ErrNoRows {
		return nil, nil // Not found, but not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve query embedding: %w", err)
	}
	
	embedding, err := parseVectorString(vectorStr, v.dimension)
	if err != nil {
		return nil, err
	}
	
	return embedding, nil
}

// AddEmbeddingWithText stores a query embedding with optional query text
func (v *VectorDBCache) AddEmbeddingWithText(queryHash string, queryText string, embedding []float32) error {
	if len(embedding) != v.dimension {
		return fmt.Errorf("embedding dimension mismatch: expected %d, got %d", v.dimension, len(embedding))
	}
	
	// Try database first if enabled
	if v.dbEnabled {
		db := GetDB()
		if db != nil {
			vectorStr := float32SliceToVectorString(embedding)
			
			_, err := db.ExecContext(
				context.Background(),
				`INSERT INTO query_embeddings (query_hash, query_text, embedding) 
				 VALUES ($1, $2, $3::vector) 
				 ON CONFLICT (query_hash) DO NOTHING`,
				queryHash, queryText, vectorStr,
			)
			
			if err == nil {
				// Successfully stored in DB, also store in-memory for fast access if available
				if v.vectorDB != nil {
					v.mu.Lock()
					_ = v.vectorDB.Add(queryHash, embedding)
					v.mu.Unlock()
				}
				return nil
			}
			
			// DB failed, fall through to in-memory
		}
	}
	
	// Fallback to in-memory
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if err := v.vectorDB.Add(queryHash, embedding); err != nil {
		return fmt.Errorf("failed to add embedding to vector DB: %w", err)
	}
	
	return nil
}

// GetResultEmbedding retrieves a result embedding from the database
func (v *VectorDBCache) GetResultEmbedding(ctx context.Context, paperID string) ([]float32, error) {
	if !v.dbEnabled {
		return nil, fmt.Errorf("database not enabled")
	}
	
	db := GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}
	
	var vectorStr string
	err := db.QueryRowContext(
		ctx,
		`SELECT embedding::text FROM result_embeddings WHERE paper_id = $1`,
		paperID,
	).Scan(&vectorStr)
	
	if err == sql.ErrNoRows {
		return nil, nil // Not found, but not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve embedding: %w", err)
	}
	
	// Parse vector string format: [0.1,0.2,0.3,...]
	vectorStr = strings.Trim(vectorStr, "[]")
	if vectorStr == "" {
		return nil, fmt.Errorf("empty vector string")
	}
	
	parts := strings.Split(vectorStr, ",")
	embedding := make([]float32, len(parts))
	for i, part := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(part), 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse vector value at index %d: %w", i, err)
		}
		embedding[i] = float32(val)
	}
	
	if len(embedding) != v.dimension {
		return nil, fmt.Errorf("embedding dimension mismatch: expected %d, got %d", v.dimension, len(embedding))
	}
	
	return embedding, nil
}

// GetResultEmbeddingsBatch retrieves multiple result embeddings from the database
func (v *VectorDBCache) GetResultEmbeddingsBatch(ctx context.Context, paperIDs []string) (map[string][]float32, error) {
	if !v.dbEnabled {
		return nil, fmt.Errorf("database not enabled")
	}
	
	db := GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}
	
	if len(paperIDs) == 0 {
		return make(map[string][]float32), nil
	}
	
	// Use ANY(array) for better performance - single query, no chunking needed
	// PostgreSQL handles arrays efficiently up to very large sizes
	query := `SELECT paper_id, embedding::text FROM result_embeddings WHERE paper_id = ANY($1::text[])`
	
	rows, err := db.QueryContext(ctx, query, paperIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query embeddings: %w", err)
	}
	defer rows.Close()
	
	result := make(map[string][]float32, len(paperIDs))
	for rows.Next() {
		var paperID, vectorStr string
		if err := rows.Scan(&paperID, &vectorStr); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		embedding, err := parseVectorString(vectorStr, v.dimension)
		if err == nil {
			result[paperID] = embedding
		}
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	
	return result, nil
}

// parseVectorString parses pgvector string format to []float32
func parseVectorString(vectorStr string, dimension int) ([]float32, error) {
	vectorStr = strings.Trim(vectorStr, "[]")
	if vectorStr == "" {
		return nil, fmt.Errorf("empty vector string")
	}
	
	parts := strings.Split(vectorStr, ",")
	if len(parts) != dimension {
		return nil, fmt.Errorf("dimension mismatch: expected %d, got %d", dimension, len(parts))
	}
	
	embedding := make([]float32, dimension)
	for i, part := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(part), 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse vector value at index %d: %w", i, err)
		}
		embedding[i] = float32(val)
	}
	
	return embedding, nil
}

// AddResultEmbedding stores a result embedding in the database
func (v *VectorDBCache) AddResultEmbedding(paperID string, embedding []float32) error {
	return v.AddResultEmbeddingsBatch(map[string][]float32{paperID: embedding})
}

// AddResultEmbeddingsBatch stores multiple result embeddings in a single batch insert
// Uses UNNEST for efficient bulk inserts (faster than multiple VALUES)
func (v *VectorDBCache) AddResultEmbeddingsBatch(embeddings map[string][]float32) error {
	if len(embeddings) == 0 {
		return nil
	}
	
	// Validate all embeddings have correct dimension
	for paperID, embedding := range embeddings {
		if len(embedding) != v.dimension {
			return fmt.Errorf("embedding dimension mismatch for paper %s: expected %d, got %d", paperID, v.dimension, len(embedding))
		}
	}
	
	// Try database first if enabled
	if v.dbEnabled {
		db := GetDB()
		if db != nil {
			// Use UNNEST with arrays for more efficient bulk insert
			// PostgreSQL handles this better than many VALUES clauses
			paperIDs := make([]string, 0, len(embeddings))
			vectorStrs := make([]string, 0, len(embeddings))
			
			for paperID, embedding := range embeddings {
				paperIDs = append(paperIDs, paperID)
				vectorStrs = append(vectorStrs, float32SliceToVectorString(embedding))
			}
			
			// Use UNNEST for efficient bulk insert
			query := `
				INSERT INTO result_embeddings (paper_id, embedding)
				SELECT unnest($1::text[]), unnest($2::vector[])
				ON CONFLICT (paper_id) DO NOTHING`
			
			_, err := db.ExecContext(context.Background(), query, paperIDs, vectorStrs)
			if err == nil {
				return nil
			}
			
			// Fallback: Build batch INSERT with multiple VALUES (chunked if needed)
			const maxBatchSize = 100 // PostgreSQL can handle more, but chunk for safety
			allPaperIDs := make([]string, 0, len(embeddings))
			allVectorStrs := make([]string, 0, len(embeddings))
			for paperID, embedding := range embeddings {
				allPaperIDs = append(allPaperIDs, paperID)
				allVectorStrs = append(allVectorStrs, float32SliceToVectorString(embedding))
			}
			
			totalStored := int64(0)
			for i := 0; i < len(allPaperIDs); i += maxBatchSize {
				end := i + maxBatchSize
				if end > len(allPaperIDs) {
					end = len(allPaperIDs)
				}
				
				chunkPaperIDs := allPaperIDs[i:end]
				chunkVectorStrs := allVectorStrs[i:end]
				
				args := make([]interface{}, 0, len(chunkPaperIDs)*2)
				placeholders := make([]string, 0, len(chunkPaperIDs))
				
				argIdx := 1
				for j := range chunkPaperIDs {
					args = append(args, chunkPaperIDs[j], chunkVectorStrs[j])
					placeholders = append(placeholders, fmt.Sprintf("($%d, $%d::vector)", argIdx, argIdx+1))
					argIdx += 2
				}
				
				chunkQuery := fmt.Sprintf(
					`INSERT INTO result_embeddings (paper_id, embedding) 
					 VALUES %s 
					 ON CONFLICT (paper_id) DO NOTHING`,
					strings.Join(placeholders, ", "),
				)
				
				chunkResult, err := db.ExecContext(context.Background(), chunkQuery, args...)
				if err != nil {
					logger.Error("Failed to store result embeddings chunk", err, map[string]interface{}{
						"chunk_start": i,
						"chunk_end":   end,
						"chunk_size":  len(chunkPaperIDs),
					})
					continue
				}
				
				rowsAffected, _ := chunkResult.RowsAffected()
				totalStored += rowsAffected
			}
			
			return nil
		}
	}
	
	return nil
}

// RerankResults reranks search results using vector similarity
// Uses database if available, falls back to in-memory serverlessVector
func (v *VectorDBCache) RerankResults(
	ctx context.Context,
	queryEmbedding []float32,
	results []SearchResult,
	resultEmbeddings [][]float32,
) ([]SearchResult, error) {
	if len(results) == 0 {
		return results, nil
	}
	
	if len(queryEmbedding) != v.dimension {
		return results, fmt.Errorf("query embedding dimension mismatch: expected %d, got %d", v.dimension, len(queryEmbedding))
	}
	
	if len(results) != len(resultEmbeddings) {
		return results, fmt.Errorf("results and embeddings length mismatch: %d vs %d", len(results), len(resultEmbeddings))
	}
	
	// Try database first if enabled
	if v.dbEnabled {
		db := GetDB()
		if db != nil {
			reranked, err := v.rerankWithDB(ctx, db, queryEmbedding, results, resultEmbeddings)
			if err == nil {
				return reranked, nil
			}
			
			// DB failed, fall through to in-memory
		}
	}
	
	// Fallback to in-memory serverlessVector
	return v.rerankInMemory(queryEmbedding, results, resultEmbeddings)
}

// SearchSimilarInDB performs similarity search directly in the database
// Returns top K most similar papers, regardless of whether they're in the input results
func (v *VectorDBCache) SearchSimilarInDB(
	ctx context.Context,
	queryEmbedding []float32,
	limit int,
) ([]string, error) {
	if !v.dbEnabled {
		return nil, fmt.Errorf("database not enabled")
	}
	
	db := GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}
	
	if limit <= 0 {
		limit = 100 // Default limit
	}
	
	queryVectorStr := float32SliceToVectorString(queryEmbedding)
	
	// Use HNSW index for fast similarity search - get top K by dot product
	query := `
		SELECT paper_id 
		FROM result_embeddings 
		ORDER BY embedding <#> $1::vector 
		LIMIT $2`
	
	rows, err := db.QueryContext(ctx, query, queryVectorStr, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query database for similarity search: %w", err)
	}
	defer rows.Close()
	
	paperIDs := make([]string, 0, limit)
	for rows.Next() {
		var paperID string
		if err := rows.Scan(&paperID); err != nil {
			continue
		}
		paperIDs = append(paperIDs, paperID)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating similarity search results: %w", err)
	}
	
	return paperIDs, nil
}

// rerankWithDB performs reranking using the database
// Queries DB for embeddings by paper IDs, calculates dot product in SQL, returns sorted results
func (v *VectorDBCache) rerankWithDB(
	ctx context.Context,
	db *sql.DB,
	queryEmbedding []float32,
	results []SearchResult,
	resultEmbeddings [][]float32,
) ([]SearchResult, error) {
	// Create map of paper IDs to results for lookup
	resultMap := make(map[string]SearchResult, len(results))
	paperIDs := make([]string, 0, len(results))
	for _, result := range results {
		resultMap[result.ID] = result
		paperIDs = append(paperIDs, result.ID)
	}
	
	if len(paperIDs) == 0 {
		return results, nil
	}
	
	// Use ANY(array) for better performance
	queryVectorStr := float32SliceToVectorString(queryEmbedding)
	
	// Single SQL query: filter by paper IDs, calculate dot product, order by similarity
	query := `
		SELECT paper_id, (embedding <#> $1::vector) * -1.0 as similarity 
		FROM result_embeddings 
		WHERE paper_id = ANY($2::text[])
		ORDER BY similarity DESC`
	
	rows, err := db.QueryContext(ctx, query, queryVectorStr, paperIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query database for reranking: %w", err)
	}
	defer rows.Close()
	
	// Build reranked results from DB query (already sorted by similarity)
	reranked := make([]SearchResult, 0, len(results))
	seen := make(map[string]bool, len(results))
	
	for rows.Next() {
		var paperID string
		var similarity float64
		if err := rows.Scan(&paperID, &similarity); err != nil {
			continue // Skip invalid rows
		}
		
		if result, exists := resultMap[paperID]; exists && !seen[paperID] {
			reranked = append(reranked, result)
			seen[paperID] = true
		}
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reranking results: %w", err)
	}
	
	// Add any results that weren't in the database (missing embeddings)
	for _, result := range results {
		if !seen[result.ID] {
			reranked = append(reranked, result)
		}
	}
	
	return reranked, nil
}

// rerankInMemory performs reranking using in-memory serverlessVector
func (v *VectorDBCache) rerankInMemory(
	queryEmbedding []float32,
	results []SearchResult,
	resultEmbeddings [][]float32,
) ([]SearchResult, error) {
	if v.vectorDB == nil {
		return results, nil
	}
	
	v.mu.Lock()
	defer v.mu.Unlock()
	
	// Create temporary IDs for result embeddings
	tempIDs := make([]string, len(results))
	resultMap := make(map[string]SearchResult, len(results))
	
	// Add all result embeddings to vector DB with temporary IDs
	for i, embedding := range resultEmbeddings {
		if len(embedding) != v.dimension {
			tempIDs[i] = fmt.Sprintf("rerank_invalid_%d", i)
			continue
		}
		
		tempID := fmt.Sprintf("rerank_%d", i)
		tempIDs[i] = tempID
		resultMap[tempID] = results[i]
		
		if err := v.vectorDB.Add(tempID, embedding); err != nil {
			// Clean up any embeddings we've added so far
			for j := 0; j < i; j++ {
				if tempIDs[j] != "" {
					v.vectorDB.Delete(tempIDs[j])
				}
			}
			return results, fmt.Errorf("failed to add result embedding %d: %w", i, err)
		}
	}
	
	// Use serverlessVector's optimized Search to get similarity scores
	searchResults, err := v.vectorDB.Search(queryEmbedding, len(results))
	if err != nil {
		// Clean up temporary embeddings
		for _, tempID := range tempIDs {
			if tempID != "" {
				v.vectorDB.Delete(tempID)
			}
		}
		return results, fmt.Errorf("vector search failed: %w", err)
	}
	
	// Clean up temporary embeddings
	for _, tempID := range tempIDs {
		if tempID != "" {
			v.vectorDB.Delete(tempID)
		}
	}
	
	// Map search results back to original SearchResult objects
	reranked := make([]SearchResult, 0, len(results))
	seen := make(map[string]bool, len(results))
	
	// Add results in order of similarity (highest first)
	for _, searchResult := range searchResults.Results {
		if result, exists := resultMap[searchResult.ID]; exists && !seen[searchResult.ID] {
			reranked = append(reranked, result)
			seen[searchResult.ID] = true
		}
	}
	
	// Add any results that weren't found in search (invalid embeddings)
	for i, tempID := range tempIDs {
		if tempID != "" && !seen[tempID] {
			reranked = append(reranked, results[i])
		}
	}
	
	// Ensure we return all results (in case some weren't in search results)
	if len(reranked) < len(results) {
		for _, result := range results {
			found := false
			for _, r := range reranked {
				if r.ID == result.ID {
					found = true
					break
				}
			}
			if !found {
				reranked = append(reranked, result)
			}
		}
	}
	
	return reranked, nil
}


// Size returns the number of cached embeddings (in-memory only)
func (v *VectorDBCache) Size() int {
	if v.vectorDB == nil {
		return 0
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.vectorDB.Size()
}

// Global cache instance (thread-safe)
var globalCache *VectorDBCache
var cacheOnce sync.Once

// GetVectorDBCache returns the global vector DB cache instance
func GetVectorDBCache() *VectorDBCache {
	cacheOnce.Do(func() {
		globalCache = NewVectorDBCache(defaultDimension)
	})
	return globalCache
}
