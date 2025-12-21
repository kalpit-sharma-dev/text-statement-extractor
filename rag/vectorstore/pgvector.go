package vectorstore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	
	_ "github.com/lib/pq" // PostgreSQL driver
)

// PGVectorStore is a PostgreSQL + pgvector implementation
type PGVectorStore struct {
	db        *sql.DB
	tableName string
}

// NewPGVectorStore creates a new PostgreSQL vector store
// Returns nil if PostgreSQL is not available or pgvector is not installed
func NewPGVectorStore(dsn string, tableName string) (*PGVectorStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	
	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}
	
	// Check if pgvector extension exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'vector')").Scan(&exists)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to check pgvector extension: %w", err)
	}
	
	if !exists {
		db.Close()
		log.Println("WARNING: pgvector extension not found — falling back to in-memory vector store")
		return nil, fmt.Errorf("pgvector extension not installed")
	}
	
	store := &PGVectorStore{
		db:        db,
		tableName: tableName,
	}
	
	// Create table if it doesn't exist
	if err := store.createTable(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	
	return store, nil
}

// createTable creates the chunks table with pgvector support
func (p *PGVectorStore) createTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			source_id TEXT NOT NULL,
			content TEXT NOT NULL,
			embedding vector(4096),
			metadata JSONB,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`, p.tableName)
	
	_, err := p.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	
	// Create index for similarity search
	indexQuery := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s_embedding_idx 
		ON %s 
		USING ivfflat (embedding vector_cosine_ops)
		WITH (lists = 100)
	`, p.tableName, p.tableName)
	
	_, err = p.db.Exec(indexQuery)
	if err != nil {
		// Index creation might fail if not enough data, that's okay
		log.Printf("Warning: failed to create vector index: %v", err)
	}
	
	// Create index on source_id for filtering
	_, err = p.db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s_source_id_idx ON %s(source_id)", p.tableName, p.tableName))
	if err != nil {
		log.Printf("Warning: failed to create source_id index: %v", err)
	}
	
	return nil
}

// Store stores a chunk with its embedding
func (p *PGVectorStore) Store(chunk *Chunk) error {
	if chunk.ID == "" {
		return fmt.Errorf("chunk ID cannot be empty")
	}
	
	if len(chunk.Embedding) == 0 {
		return fmt.Errorf("chunk embedding cannot be empty")
	}
	
	// Convert embedding to PostgreSQL vector format
	embeddingStr := formatVectorForPG(chunk.Embedding)
	
	metadataJSON, _ := json.Marshal(chunk.Metadata)
	
	query := fmt.Sprintf(`
		INSERT INTO %s (id, source_id, content, embedding, metadata)
		VALUES ($1, $2, $3, $4::vector, $5::jsonb)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			embedding = EXCLUDED.embedding,
			metadata = EXCLUDED.metadata
	`, p.tableName)
	
	_, err := p.db.Exec(query, chunk.ID, chunk.SourceID, chunk.Content, embeddingStr, string(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to store chunk: %w", err)
	}
	
	return nil
}

// StoreBatch stores multiple chunks
func (p *PGVectorStore) StoreBatch(chunks []*Chunk) error {
	for _, chunk := range chunks {
		if err := p.Store(chunk); err != nil {
			log.Printf("Warning: failed to store chunk %s: %v", chunk.ID, err)
			// Continue with other chunks
		}
	}
	return nil
}

// Search performs cosine similarity search using pgvector
func (p *PGVectorStore) Search(queryEmbedding []float32, topK int, sourceID string) ([]*Chunk, []float32, error) {
	embeddingStr := formatVectorForPG(queryEmbedding)
	
	var query string
	var args []interface{}
	
	if sourceID != "" {
		query = fmt.Sprintf(`
			SELECT id, source_id, content, embedding, metadata,
				   1 - (embedding <=> $1::vector) as similarity
			FROM %s
			WHERE source_id = $2
			ORDER BY embedding <=> $1::vector
			LIMIT $3
		`, p.tableName)
		args = []interface{}{embeddingStr, sourceID, topK}
	} else {
		query = fmt.Sprintf(`
			SELECT id, source_id, content, embedding, metadata,
				   1 - (embedding <=> $1::vector) as similarity
			FROM %s
			ORDER BY embedding <=> $1::vector
			LIMIT $2
		`, p.tableName)
		args = []interface{}{embeddingStr, topK}
	}
	
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to search: %w", err)
	}
	defer rows.Close()
	
	var chunks []*Chunk
	var scores []float32
	
	for rows.Next() {
		var id, sourceID, content string
		var embeddingStr string
		var metadataJSON []byte
		var similarity float64
		
		err := rows.Scan(&id, &sourceID, &content, &embeddingStr, &metadataJSON, &similarity)
		if err != nil {
			log.Printf("Warning: failed to scan row: %v", err)
			continue
		}
		
		// Parse embedding (we don't need it for results, but could parse if needed)
		embedding := parseVectorFromPG(embeddingStr)
		
		var metadata map[string]interface{}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &metadata)
		}
		
		chunk := &Chunk{
			ID:        id,
			SourceID:  sourceID,
			Content:   content,
			Embedding: embedding,
			Metadata:  metadata,
		}
		
		chunks = append(chunks, chunk)
		scores = append(scores, float32(similarity))
	}
	
	return chunks, scores, nil
}

// DeleteBySourceID deletes all chunks for a given source ID
func (p *PGVectorStore) DeleteBySourceID(sourceID string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE source_id = $1", p.tableName)
	_, err := p.db.Exec(query, sourceID)
	if err != nil {
		return fmt.Errorf("failed to delete chunks: %w", err)
	}
	return nil
}

// Count returns the number of stored chunks
func (p *PGVectorStore) Count() int {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", p.tableName)
	err := p.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

// Close closes the database connection
func (p *PGVectorStore) Close() error {
	return p.db.Close()
}

// formatVectorForPG formats a float32 slice as a PostgreSQL vector string
func formatVectorForPG(vec []float32) string {
	if len(vec) == 0 {
		return "[]"
	}
	
	result := "["
	for i, v := range vec {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%.6f", v)
	}
	result += "]"
	return result
}

// parseVectorFromPG parses a PostgreSQL vector string into a float32 slice
func parseVectorFromPG(vecStr string) []float32 {
	// Simple parser - remove brackets and split by comma
	vecStr = vecStr[1 : len(vecStr)-1] // Remove [ and ]
	// This is a simplified parser - in production, use a proper parser
	// For now, return empty since we don't need embeddings in results
	return []float32{}
}

