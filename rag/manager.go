package rag

import (
	"fmt"
	"log"
	"sync"

	"classify/rag/vectorstore"
)

// Manager manages the RAG pipeline
type Manager struct {
	config          *Config
	chunker         *Chunker
	embeddingClient *EmbeddingClient
	vectorStore     vectorstore.VectorStore
	retriever       *Retriever
	initialized     bool
	mu              sync.RWMutex
}

// NewManager creates and initializes a new RAG manager
func NewManager(config *Config) (*Manager, error) {
	manager := &Manager{
		config:  config,
		chunker: NewChunker(config),
	}

	// Initialize embedding client
	manager.embeddingClient = NewEmbeddingClient(config)

	// Initialize vector store (try PostgreSQL first, fallback to memory)
	var vs vectorstore.VectorStore
	if config.PostgreSQLDSN != "" {
		pgStore, err := vectorstore.NewPGVectorStore(config.PostgreSQLDSN, config.TableName)
		if err != nil {
			log.Printf("WARNING: pgvector not available — falling back to in-memory vector store: %v", err)
			vs = vectorstore.NewMemoryVectorStore()
		} else {
			log.Println("Using PostgreSQL + pgvector for vector storage")
			vs = pgStore
		}
	} else {
		log.Println("No PostgreSQL DSN provided — using in-memory vector store")
		vs = vectorstore.NewMemoryVectorStore()
	}

	manager.vectorStore = vs

	// Initialize retriever
	manager.retriever = NewRetriever(manager.embeddingClient, manager.vectorStore, config)

	manager.initialized = true

	return manager, nil
}

// HasChunks checks if chunks already exist for a source ID
func (m *Manager) HasChunks(sourceID string) (bool, error) {
	if !m.initialized {
		return false, fmt.Errorf("RAG manager not initialized")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Try to retrieve chunks - if we get any, they exist
	chunks, _, err := m.vectorStore.Search(make([]float32, m.config.EmbeddingDims), 1, sourceID)
	if err != nil {
		return false, err
	}

	// Check if any chunks have the correct source ID
	for _, chunk := range chunks {
		if chunk.SourceID == sourceID {
			return true, nil
		}
	}

	return false, nil
}

// IndexStatementData indexes statement data using the RAG pipeline
// This is the main entry point for indexing: Chunk → Embed → Store
// If chunks already exist for this sourceID, they will be replaced
func (m *Manager) IndexStatementData(statementData interface{}, sourceID string) error {
	if !m.initialized {
		return fmt.Errorf("RAG manager not initialized")
	}

	// Check if chunks already exist BEFORE acquiring lock (to avoid deadlock)
	hasChunks, err := m.HasChunks(sourceID)
	if err == nil && hasChunks {
		log.Printf("Chunks already exist for source %s, replacing...", sourceID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Delete old chunks if they exist (now that we have the lock)
	if hasChunks {
		if err := m.vectorStore.DeleteBySourceID(sourceID); err != nil {
			log.Printf("Warning: failed to delete old chunks: %v", err)
			// Continue anyway - new chunks will have different IDs
		}
	}

	// Step 1: Chunk the statement data
	log.Printf("Chunking statement data for source: %s", sourceID)
	chunks, err := m.chunker.ChunkStatementData(statementData, sourceID)
	if err != nil {
		log.Printf("Error during chunking: %v", err)
		return fmt.Errorf("failed to chunk statement data: %w", err)
	}

	log.Printf("Created %d chunks", len(chunks))

	if len(chunks) == 0 {
		return fmt.Errorf("no chunks created from statement data")
	}

	// Step 2: Generate embeddings
	log.Printf("Generating embeddings for %d chunks", len(chunks))
	if err := m.embeddingClient.EmbedChunks(chunks); err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Step 3: Filter out chunks without embeddings (critical!)
	validChunks := make([]*Chunk, 0, len(chunks))
	for _, chunk := range chunks {
		if len(chunk.Embedding) > 0 {
			validChunks = append(validChunks, chunk)
		} else {
			log.Printf("Warning: chunk %s has no embedding, skipping", chunk.ID)
		}
	}

	if len(validChunks) == 0 {
		return fmt.Errorf("no valid chunks with embeddings to store")
	}

	log.Printf("Storing %d valid chunks (filtered from %d total)", len(validChunks), len(chunks))

	// Step 4: Store in vector store
	// Convert rag.Chunk to vectorstore.Chunk
	vsChunks := make([]*vectorstore.Chunk, len(validChunks))
	for i, chunk := range validChunks {
		vsChunks[i] = &vectorstore.Chunk{
			ID:        chunk.ID,
			SourceID:  chunk.SourceID,
			Content:   chunk.Content,
			Embedding: chunk.Embedding,
			Metadata:  chunk.Metadata,
		}
	}
	if err := m.vectorStore.StoreBatch(vsChunks); err != nil {
		return fmt.Errorf("failed to store chunks: %w", err)
	}

	log.Printf("Successfully indexed statement data: %d chunks stored", len(validChunks))
	return nil
}

// RetrieveRelevantChunks retrieves relevant chunks for a query
func (m *Manager) RetrieveRelevantChunks(query string, sourceID string) ([]*Chunk, error) {
	if !m.initialized {
		return nil, fmt.Errorf("RAG manager not initialized")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.retriever.Retrieve(query, sourceID)
}

// RetrieveRelevantChunksWithScores retrieves relevant chunks with similarity scores
func (m *Manager) RetrieveRelevantChunksWithScores(query string, sourceID string) ([]RetrievedChunk, error) {
	if !m.initialized {
		return nil, fmt.Errorf("RAG manager not initialized")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.retriever.RetrieveWithScores(query, sourceID)
}

// DeleteStatementData deletes all chunks for a given source ID
func (m *Manager) DeleteStatementData(sourceID string) error {
	if !m.initialized {
		return fmt.Errorf("RAG manager not initialized")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return m.vectorStore.DeleteBySourceID(sourceID)
}

// GetVectorStore returns the vector store (for testing/debugging)
func (m *Manager) GetVectorStore() vectorstore.VectorStore {
	return m.vectorStore
}

// GetEmbeddingClient returns the embedding client (for query embedding generation)
func (m *Manager) GetEmbeddingClient() *EmbeddingClient {
	return m.embeddingClient
}

// GetConfig returns the configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}
