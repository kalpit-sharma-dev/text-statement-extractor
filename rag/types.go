package rag

// Chunk represents a single chunk of text with metadata
type Chunk struct {
	ID        string                 `json:"id"`        // Unique chunk identifier
	SourceID  string                 `json:"source_id"` // Identifier for the source document
	Content   string                 `json:"content"`   // The actual text content
	Embedding []float32              `json:"embedding,omitempty"` // Vector embedding (set after embedding)
	Metadata  map[string]interface{} `json:"metadata,omitempty"` // Additional metadata
}

