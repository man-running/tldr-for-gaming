-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create query_embeddings table
CREATE TABLE IF NOT EXISTS query_embeddings (
    id SERIAL PRIMARY KEY,
    query_hash TEXT UNIQUE NOT NULL,
    query_text TEXT,
    embedding VECTOR(512) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create result_embeddings table
CREATE TABLE IF NOT EXISTS result_embeddings (
    id SERIAL PRIMARY KEY,
    paper_id TEXT NOT NULL,
    embedding VECTOR(512) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create HNSW indexes for efficient similarity search (dot product)
CREATE INDEX IF NOT EXISTS query_embeddings_embedding_idx ON query_embeddings 
    USING hnsw (embedding vector_ip_ops);

CREATE INDEX IF NOT EXISTS result_embeddings_embedding_idx ON result_embeddings 
    USING hnsw (embedding vector_ip_ops);

-- Create index on paper_id for faster lookups
CREATE INDEX IF NOT EXISTS result_embeddings_paper_id_idx ON result_embeddings (paper_id);

