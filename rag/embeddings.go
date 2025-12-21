package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// EmbeddingClient handles embedding generation via Ollama
type EmbeddingClient struct {
	config *Config
	client *http.Client
}

// NewEmbeddingClient creates a new embedding client
func NewEmbeddingClient(config *Config) *EmbeddingClient {
	return &EmbeddingClient{
		config: config,
		client: &http.Client{
			Timeout: 60 * time.Second, // 60 second timeout for embedding requests
		},
	}
}

// EmbeddingRequest represents a request to Ollama embeddings API
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse represents the response from Ollama embeddings API
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// GenerateEmbedding generates an embedding for a single text
func (e *EmbeddingClient) GenerateEmbedding(text string) ([]float32, error) {
	req := EmbeddingRequest{
		Model:  e.config.EmbeddingModel,
		Prompt: text,
	}
	
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}
	
	url := e.config.GetEmbeddingURL()
	resp, err := e.client.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama embeddings API: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama embeddings API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}
	
	// Convert []float64 to []float32
	embedding := make([]float32, len(embeddingResp.Embedding))
	for i, v := range embeddingResp.Embedding {
		embedding[i] = float32(v)
	}
	
	return embedding, nil
}

// GenerateEmbeddingsBatch generates embeddings for multiple texts concurrently
// Uses goroutines to parallelize embedding generation for better performance
func (e *EmbeddingClient) GenerateEmbeddingsBatch(texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	type result struct {
		index     int
		embedding []float32
		err       error
	}
	
	// Channel for results
	results := make(chan result, len(texts))
	
	// Generate embeddings concurrently (limit concurrency to avoid overwhelming Ollama)
	maxConcurrency := 5
	semaphore := make(chan struct{}, maxConcurrency)
	
	for i, text := range texts {
		go func(idx int, txt string) {
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			log.Printf("Generating embedding %d/%d...", idx+1, len(texts))
			embedding, err := e.GenerateEmbedding(txt)
			if err != nil {
				log.Printf("Failed to generate embedding %d: %v", idx, err)
			} else {
				log.Printf("Successfully generated embedding %d/%d", idx+1, len(texts))
			}
			results <- result{
				index:     idx,
				embedding: embedding,
				err:       err,
			}
		}(i, text)
	}
	
	// Collect results
	successCount := 0
	for i := 0; i < len(texts); i++ {
		res := <-results
		if res.err != nil {
			log.Printf("Warning: failed to generate embedding for text %d: %v", res.index, res.err)
			// Leave embeddings[res.index] as nil
		} else {
			embeddings[res.index] = res.embedding
			successCount++
		}
	}
	
	if successCount == 0 {
		return nil, fmt.Errorf("failed to generate any embeddings")
	}
	
	log.Printf("Successfully generated %d/%d embeddings", successCount, len(texts))
	return embeddings, nil
}

// EmbedChunks generates embeddings for all chunks
func (e *EmbeddingClient) EmbedChunks(chunks []*Chunk) error {
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}
	
	embeddings, err := e.GenerateEmbeddingsBatch(texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}
	
	// Assign embeddings to chunks
	for i, chunk := range chunks {
		if i < len(embeddings) && embeddings[i] != nil {
			chunk.Embedding = embeddings[i]
		}
	}
	
	return nil
}

