package vectorstore

// Chunk represents a chunk for vector storage (to avoid import cycles)
type Chunk struct {
	ID        string
	SourceID  string
	Content   string
	Embedding []float32
	Metadata  map[string]interface{}
}

// VectorStore defines the interface for vector storage
type VectorStore interface {
	// Store stores a single chunk with its embedding
	Store(chunk *Chunk) error
	
	// StoreBatch stores multiple chunks
	StoreBatch(chunks []*Chunk) error
	
	// Search performs similarity search and returns top K chunks with scores
	Search(queryEmbedding []float32, topK int, sourceID string) ([]*Chunk, []float32, error)
	
	// DeleteBySourceID deletes all chunks for a given source ID
	DeleteBySourceID(sourceID string) error
	
	// Count returns the number of stored chunks
	Count() int
}

