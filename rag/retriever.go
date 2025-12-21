package rag

import (
	"fmt"
	"log"
	
	"classify/rag/vectorstore"
)

// Retriever handles retrieval of relevant chunks based on queries
type Retriever struct {
	embeddingClient *EmbeddingClient
	vectorStore     vectorstore.VectorStore
	config          *Config
}

// NewRetriever creates a new retriever
func NewRetriever(embeddingClient *EmbeddingClient, vectorStore vectorstore.VectorStore, config *Config) *Retriever {
	return &Retriever{
		embeddingClient: embeddingClient,
		vectorStore:     vectorStore,
		config:          config,
	}
}

// RetrievedChunk represents a chunk with its similarity score
type RetrievedChunk struct {
	Chunk  *Chunk
	Score  float32
}

// Retrieve finds the most relevant chunks for a query
func (r *Retriever) Retrieve(query string, sourceID string) ([]*Chunk, error) {
	chunksWithScores, err := r.RetrieveWithScores(query, sourceID)
	if err != nil {
		return nil, err
	}
	
	// Extract just the chunks
	chunks := make([]*Chunk, len(chunksWithScores))
	for i, cws := range chunksWithScores {
		chunks[i] = cws.Chunk
	}
	
	return chunks, nil
}

// RetrieveWithScores finds the most relevant chunks with their similarity scores
func (r *Retriever) RetrieveWithScores(query string, sourceID string) ([]RetrievedChunk, error) {
	// Generate embedding for the query
	queryEmbedding, err := r.embeddingClient.GenerateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}
	
	// Search for similar chunks
	vsChunks, scores, err := r.vectorStore.Search(queryEmbedding, r.config.TopK, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to search vector store: %w", err)
	}
	
	// Convert vectorstore.Chunk to rag.Chunk with scores
	chunksWithScores := make([]RetrievedChunk, len(vsChunks))
	for i, vsChunk := range vsChunks {
		score := float32(0.0)
		if i < len(scores) {
			score = scores[i]
		}
		
		chunksWithScores[i] = RetrievedChunk{
			Chunk: &Chunk{
				ID:        vsChunk.ID,
				SourceID:  vsChunk.SourceID,
				Content:   vsChunk.Content,
				Embedding: vsChunk.Embedding,
				Metadata:  vsChunk.Metadata,
			},
			Score: score,
		}
	}
	
	// Sort by score descending (most relevant first)
	for i := 0; i < len(chunksWithScores)-1; i++ {
		for j := i + 1; j < len(chunksWithScores); j++ {
			if chunksWithScores[i].Score < chunksWithScores[j].Score {
				chunksWithScores[i], chunksWithScores[j] = chunksWithScores[j], chunksWithScores[i]
			}
		}
	}
	
	// Filter out chunks with very low similarity
	similarityThreshold := r.config.SimilarityThreshold
	filtered := make([]RetrievedChunk, 0, len(chunksWithScores))
	
	for _, cws := range chunksWithScores {
		if cws.Score >= similarityThreshold {
			filtered = append(filtered, cws)
		} else {
			log.Printf("Filtered out chunk %s with low similarity: %.3f", cws.Chunk.ID, cws.Score)
		}
	}
	
	if len(filtered) == 0 && len(chunksWithScores) > 0 {
		// If all chunks were filtered but we have some, return at least the top one
		log.Printf("Warning: All chunks filtered by similarity threshold, returning top chunk anyway")
		return chunksWithScores[:1], nil
	}
	
	return filtered, nil
}

