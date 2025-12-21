package rag

import (
	"log"
	"sort"
	"strings"
)

// Reranker re-ranks chunks based on query relevance
type Reranker struct {
	embeddingClient *EmbeddingClient
	config          *Config
}

// NewReranker creates a new reranker
func NewReranker(embeddingClient *EmbeddingClient, config *Config) *Reranker {
	return &Reranker{
		embeddingClient: embeddingClient,
		config:          config,
	}
}

// RerankChunks re-ranks chunks using query-chunk relevance scoring
func (r *Reranker) RerankChunks(query string, chunks []RetrievedChunk) ([]RetrievedChunk, error) {
	if len(chunks) <= 1 {
		return chunks, nil
	}

	// Calculate relevance scores using keyword matching and semantic similarity
	type scoredChunk struct {
		chunk RetrievedChunk
		score float32
	}

	scored := make([]scoredChunk, len(chunks))

	for i, cws := range chunks {
		score := cws.Score // Start with semantic similarity

		// Boost score based on keyword matches
		contentLower := strings.ToLower(cws.Chunk.Content)

		// Extract key terms from query
		queryTerms := extractKeyTerms(query)

		// Count matches
		matches := 0
		for _, term := range queryTerms {
			if strings.Contains(contentLower, term) {
				matches++
			}
		}

		// Boost score: +0.1 per matching term (max +0.5)
		keywordBoost := float32(matches) * 0.1
		if keywordBoost > 0.5 {
			keywordBoost = 0.5
		}

		// Boost score based on chunk type relevance
		typeBoost := r.getTypeRelevanceBoost(query, cws.Chunk)

		// Final score: semantic + keyword + type
		finalScore := score + keywordBoost + typeBoost
		if finalScore > 1.0 {
			finalScore = 1.0
		}

		scored[i] = scoredChunk{
			chunk: cws,
			score: finalScore,
		}
	}

	// Sort by final score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Convert back
	reranked := make([]RetrievedChunk, len(chunks))
	for i, sc := range scored {
		reranked[i] = sc.chunk
		reranked[i].Score = sc.score // Update score
	}

	log.Printf("Reranked %d chunks (score range: %.3f - %.3f)", len(reranked), 
		reranked[len(reranked)-1].Score, reranked[0].Score)

	return reranked, nil
}

// extractKeyTerms extracts important terms from query
func extractKeyTerms(query string) []string {
	// Remove common stop words
	stopWords := map[string]bool{
		"what": true, "was": true, "my": true, "the": true,
		"a": true, "an": true, "is": true, "are": true,
		"how": true, "many": true, "much": true, "which": true,
		"where": true, "when": true, "who": true, "why": true,
		"this": true, "that": true, "these": true, "those": true,
		"and": true, "or": true, "but": true, "with": true,
		"from": true, "to": true, "of": true, "in": true,
		"on": true, "at": true, "by": true, "for": true,
	}

	words := strings.Fields(strings.ToLower(query))
	terms := []string{}

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:")
		if !stopWords[word] && len(word) > 2 {
			terms = append(terms, word)
		}
	}

	return terms
}

// getTypeRelevanceBoost returns boost based on chunk type relevance to query
func (r *Reranker) getTypeRelevanceBoost(query string, chunk *Chunk) float32 {
	queryLower := strings.ToLower(query)
	chunkType := ""

	if chunk.Metadata != nil {
		if t, ok := chunk.Metadata["type"].(string); ok {
			chunkType = t
		}
	}

	// Type-specific boosts
	typeBoosts := map[string]map[string]float32{
		"account_summary": {
			"total": 0.15, "expense": 0.15, "income": 0.15, "balance": 0.15,
			"summary": 0.1, "overview": 0.1,
		},
		"transactions": {
			"transaction": 0.15, "payment": 0.15, "debit": 0.15, "credit": 0.15,
			"transfer": 0.1,
		},
		"top_expenses": {
			"highest": 0.2, "maximum": 0.2, "top": 0.2, "largest": 0.2,
			"biggest": 0.15, "most": 0.15,
		},
		"category_summary": {
			"category": 0.15, "spending": 0.15, "expense": 0.15,
			"categories": 0.1, "group": 0.1,
		},
		"monthly_summary": {
			"month": 0.2, "monthly": 0.2, "period": 0.15,
			"months": 0.15, "over time": 0.1,
		},
		"transaction_breakdown": {
			"method": 0.15, "payment method": 0.15, "upi": 0.1,
			"imps": 0.1, "neft": 0.1,
		},
		"top_beneficiaries": {
			"beneficiary": 0.15, "merchant": 0.15, "vendor": 0.1,
			"payee": 0.1, "recipient": 0.1,
		},
	}

	if boosts, ok := typeBoosts[chunkType]; ok {
		for term, boost := range boosts {
			if strings.Contains(queryLower, term) {
				return boost
			}
		}
	}

	return 0.0
}

