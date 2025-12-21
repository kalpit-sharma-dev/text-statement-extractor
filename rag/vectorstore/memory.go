package vectorstore

import (
	"fmt"
	"math"
	"sync"
)

// MemoryVectorStore is an in-memory vector store implementation
// Used as fallback when PostgreSQL + pgvector is not available
type MemoryVectorStore struct {
	chunks map[string]*Chunk
	mu     sync.RWMutex
}

// NewMemoryVectorStore creates a new in-memory vector store
func NewMemoryVectorStore() *MemoryVectorStore {
	return &MemoryVectorStore{
		chunks: make(map[string]*Chunk),
	}
}

// Store stores a chunk with its embedding
func (m *MemoryVectorStore) Store(chunk *Chunk) error {
	if chunk.ID == "" {
		return fmt.Errorf("chunk ID cannot be empty")
	}
	
	if len(chunk.Embedding) == 0 {
		return fmt.Errorf("chunk embedding cannot be empty")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.chunks[chunk.ID] = chunk
	return nil
}

// StoreBatch stores multiple chunks
func (m *MemoryVectorStore) StoreBatch(chunks []*Chunk) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, chunk := range chunks {
		if chunk.ID == "" {
			continue
		}
		if len(chunk.Embedding) == 0 {
			continue
		}
		m.chunks[chunk.ID] = chunk
	}
	
	return nil
}

// Search performs cosine similarity search
func (m *MemoryVectorStore) Search(queryEmbedding []float32, topK int, sourceID string) ([]*Chunk, []float32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	type scoredChunk struct {
		chunk *Chunk
		score float32
	}
	
	var scoredChunks []scoredChunk
	
	// Calculate cosine similarity for all chunks
	for _, chunk := range m.chunks {
		// Filter by sourceID if provided
		if sourceID != "" && chunk.SourceID != sourceID {
			continue
		}
		
		if len(chunk.Embedding) == 0 {
			continue
		}
		
		similarity := cosineSimilarity(queryEmbedding, chunk.Embedding)
		scoredChunks = append(scoredChunks, scoredChunk{
			chunk: chunk,
			score: similarity,
		})
	}
	
	// Sort by score (descending)
	for i := 0; i < len(scoredChunks)-1; i++ {
		for j := i + 1; j < len(scoredChunks); j++ {
			if scoredChunks[i].score < scoredChunks[j].score {
				scoredChunks[i], scoredChunks[j] = scoredChunks[j], scoredChunks[i]
			}
		}
	}
	
	// Return top K
	if topK > len(scoredChunks) {
		topK = len(scoredChunks)
	}
	
	chunks := make([]*Chunk, topK)
	scores := make([]float32, topK)
	
	for i := 0; i < topK; i++ {
		chunks[i] = scoredChunks[i].chunk
		scores[i] = scoredChunks[i].score
	}
	
	return chunks, scores, nil
}

// DeleteBySourceID deletes all chunks for a given source ID
func (m *MemoryVectorStore) DeleteBySourceID(sourceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for id, chunk := range m.chunks {
		if chunk.SourceID == sourceID {
			delete(m.chunks, id)
		}
	}
	
	return nil
}

// Count returns the number of stored chunks
func (m *MemoryVectorStore) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.chunks)
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}
	
	var dotProduct float32
	var normA, normB float32
	
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	
	if normA == 0 || normB == 0 {
		return 0
	}
	
	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

