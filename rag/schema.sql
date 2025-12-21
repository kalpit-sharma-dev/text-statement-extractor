-- PostgreSQL schema for RAG vector store with pgvector
-- This schema is automatically created by the application, but provided here for reference

-- Enable pgvector extension (must be done by database admin)
-- CREATE EXTENSION IF NOT EXISTS vector;

-- Table for storing statement chunks with embeddings
CREATE TABLE IF NOT EXISTS statement_chunks (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(4096),  -- Adjust dimension based on your embedding model
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Index for similarity search using IVFFlat (Inverted File Index)
-- This index speeds up cosine similarity searches
CREATE INDEX IF NOT EXISTS statement_chunks_embedding_idx 
ON statement_chunks 
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- Index on source_id for filtering by statement
CREATE INDEX IF NOT EXISTS statement_chunks_source_id_idx 
ON statement_chunks(source_id);

-- Index on created_at for cleanup operations
CREATE INDEX IF NOT EXISTS statement_chunks_created_at_idx 
ON statement_chunks(created_at);

-- Example query for similarity search:
-- SELECT id, source_id, content, metadata,
--        1 - (embedding <=> $1::vector) as similarity
-- FROM statement_chunks
-- WHERE source_id = $2
-- ORDER BY embedding <=> $1::vector
-- LIMIT 5;

-- Notes:
-- 1. The 'lists' parameter in IVFFlat should be set to rows/1000 for optimal performance
-- 2. For production, consider using HNSW index instead of IVFFlat for better accuracy
-- 3. Adjust vector dimension (4096) based on your embedding model
-- 4. Consider adding TTL/cleanup logic for old chunks

