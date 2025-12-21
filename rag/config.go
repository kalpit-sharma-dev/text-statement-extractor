package rag

import (
	"os"
)

// Config holds RAG system configuration
type Config struct {
	// Ollama configuration
	OllamaURL      string
	EmbeddingModel string
	ChatModel      string

	// Chunking configuration
	ChunkSize    int     // Target tokens per chunk (300-500)
	ChunkOverlap float64 // Overlap percentage (0.10-0.15)

	// Retrieval configuration
	TopK                int     // Number of chunks to retrieve (default: 5)
	SimilarityThreshold float32 // Minimum similarity score (default: 0.3)

	// Vector store configuration
	UsePostgreSQL bool
	PostgreSQLDSN string
	TableName     string

	// Embedding dimensions (llama3 typically 4096)
	EmbeddingDims int
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	// Get model from environment or default to llama3
	embeddingModel := os.Getenv("OLLAMA_EMBEDDING_MODEL")
	if embeddingModel == "" {
		embeddingModel = "llama3"
	}

	chatModel := os.Getenv("OLLAMA_CHAT_MODEL")
	if chatModel == "" {
		chatModel = "llama3"
	}

	// Determine embedding dimensions based on model
	embeddingDims := 4096 // Default for llama3
	if embeddingModel == "llama4" || chatModel == "llama4" {
		// llama4 may have different dimensions - adjust if needed
		// Check actual dimensions when llama4 is available
		embeddingDims = 4096 // Update this when llama4 dimensions are known
	}

	return &Config{
		OllamaURL:           ollamaURL,
		EmbeddingModel:      embeddingModel,
		ChatModel:           chatModel,
		ChunkSize:           400,  // tokens
		ChunkOverlap:        0.12, // 12% overlap
		TopK:                5,
		SimilarityThreshold: 0.3,   // Minimum similarity score
		UsePostgreSQL:       false, // Will be auto-detected
		PostgreSQLDSN:       os.Getenv("POSTGRES_DSN"),
		TableName:           "statement_chunks",
		EmbeddingDims:       embeddingDims,
	}
}

// GetEmbeddingURL returns the full URL for embeddings API
func (c *Config) GetEmbeddingURL() string {
	return c.OllamaURL + "/api/embeddings"
}

// GetChatURL returns the full URL for chat API
func (c *Config) GetChatURL() string {
	return c.OllamaURL + "/api/chat"
}
