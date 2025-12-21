package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"classify/rag"
)

// ChatRequest represents the incoming chat request
type ChatRequest struct {
	Message             string                `json:"message"`
	StatementData       interface{}           `json:"statementData"`
	ConversationHistory []ConversationMessage `json:"conversationHistory,omitempty"`
	APIKey              string                `json:"apiKey,omitempty"`
}

// ConversationMessage represents a message in the conversation history
type ConversationMessage struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

// ChatResponse represents the response from the chat API
type ChatResponse struct {
	Success   bool   `json:"success"`
	Response  string `json:"response,omitempty"`
	Error     string `json:"error,omitempty"`
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp"`
}

// GeminiRequest represents the request to Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents content in Gemini API request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

// GeminiPart represents a part of content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	Error      *GeminiError      `json:"error,omitempty"`
}

// GeminiCandidate represents a candidate response
type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

// GeminiError represents an error from Gemini API
type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// OllamaRequest represents the request to Ollama API
type OllamaRequest struct {
	Model    string                 `json:"model"`
	Messages []OllamaMessage        `json:"messages"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// OllamaMessage represents a message in Ollama API
type OllamaMessage struct {
	Role    string `json:"role"` // "system", "user", or "assistant"
	Content string `json:"content"`
}

// OllamaResponse represents the response from Ollama API
type OllamaResponse struct {
	Message OllamaMessage `json:"message"`
	Error   string        `json:"error,omitempty"`
}

// Global RAG manager (initialized once)
var (
	ragManager *rag.Manager
	ragOnce    sync.Once
	ragMutex   sync.RWMutex
)

// getRAGManager returns the global RAG manager, initializing it if needed
func getRAGManager() (*rag.Manager, error) {
	var initErr error
	ragOnce.Do(func() {
		config := rag.DefaultConfig()
		manager, err := rag.NewManager(config)
		if err != nil {
			initErr = err
			return
		}
		ragManager = manager
	})

	if initErr != nil {
		return nil, initErr
	}

	ragMutex.RLock()
	defer ragMutex.RUnlock()
	return ragManager, nil
}

// chatHandler handles POST requests to /api/chat
func chatHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	enableCORS(w, r)

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var chatReq ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
		sendErrorResponse(w, "Invalid request body", "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Validation
	if chatReq.Message == "" {
		sendErrorResponse(w, "Missing required fields", "The 'message' field is required", http.StatusBadRequest)
		return
	}

	if chatReq.StatementData == nil {
		sendErrorResponse(w, "Missing required fields", "The 'statementData' field is required", http.StatusBadRequest)
		return
	}

	// Get API key (use provided key or fallback to environment variable)
	apiKey := chatReq.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}

	// Determine which AI service to use (Gemini if API key exists, otherwise Ollama)
	useOllama := apiKey == ""

	// Generate a unique source ID for this statement data
	// In production, you might want to hash the statement data to reuse indexes
	sourceID := generateSourceID(chatReq.StatementData)

	// Initialize RAG manager
	ragMgr, err := getRAGManager()
	if err != nil {
		log.Printf("Warning: Failed to initialize RAG manager: %v. Falling back to direct prompt.", err)
		// Fallback to old method if RAG fails
		responseText, err := handleChatWithoutRAG(apiKey, useOllama, chatReq)
		if err != nil {
			sendErrorResponse(w, "Internal server error", fmt.Sprintf("Failed to process chat: %v", err), http.StatusInternalServerError)
			return
		}
		sendSuccessResponse(w, responseText)
		return
	}

	// Check if chunks already exist for this source ID
	hasChunks, err := ragMgr.HasChunks(sourceID)
	if err != nil {
		log.Printf("Warning: Failed to check for existing chunks: %v", err)
	}

	// Index statement data if not already indexed
	if !hasChunks {
		log.Printf("Indexing statement data for source: %s", sourceID)
		if err := ragMgr.IndexStatementData(chatReq.StatementData, sourceID); err != nil {
			log.Printf("Warning: Failed to index statement data: %v. Falling back to direct prompt.", err)
			// Fallback to old method if indexing fails
			responseText, err := handleChatWithoutRAG(apiKey, useOllama, chatReq)
			if err != nil {
				sendErrorResponse(w, "Internal server error", fmt.Sprintf("Failed to process chat: %v", err), http.StatusInternalServerError)
				return
			}
			sendSuccessResponse(w, responseText)
			return
		}
	} else {
		log.Printf("Using existing chunks for source: %s", sourceID)
	}

	// Classify query for optimized retrieval
	classifier := rag.NewQueryClassifier()
	queryType := classifier.ClassifyQuery(chatReq.Message)
	log.Printf("Query classified as: %s", queryType)
	
	// Adjust TopK based on query type (temporarily override config)
	originalTopK := ragMgr.GetConfig().TopK
	optimalTopK := classifier.GetOptimalTopK(queryType, originalTopK)
	if optimalTopK != originalTopK {
		log.Printf("Adjusted TopK from %d to %d for query type %s", originalTopK, optimalTopK, queryType)
		// Note: We'll retrieve with original TopK but can filter later
	}
	
	// Retrieve relevant chunks with scores using RAG
	log.Printf("Retrieving relevant chunks for query: %s", chatReq.Message)
	chunksWithScores, err := ragMgr.RetrieveRelevantChunksWithScores(chatReq.Message, sourceID)
	if err != nil {
		log.Printf("Warning: Failed to retrieve chunks: %v. Falling back to direct prompt.", err)
		// Fallback to old method if retrieval fails
		responseText, err := handleChatWithoutRAG(apiKey, useOllama, chatReq)
		if err != nil {
			sendErrorResponse(w, "Internal server error", fmt.Sprintf("Failed to process chat: %v", err), http.StatusInternalServerError)
			return
		}
		sendSuccessResponse(w, responseText)
		return
	}

	// Check if we have any chunks - if not, fallback to direct prompt
	if len(chunksWithScores) == 0 {
		log.Printf("Warning: No relevant chunks retrieved. Falling back to direct prompt.")
		responseText, err := handleChatWithoutRAG(apiKey, useOllama, chatReq)
		if err != nil {
			sendErrorResponse(w, "Internal server error", fmt.Sprintf("Failed to process chat: %v", err), http.StatusInternalServerError)
			return
		}
		sendSuccessResponse(w, responseText)
		return
	}

	// Log scores for debugging
	scores := make([]float32, len(chunksWithScores))
	for i, cws := range chunksWithScores {
		scores[i] = cws.Score
	}
	log.Printf("Retrieved %d relevant chunks with scores: %v", len(chunksWithScores), scores)

	// Build optimized context from retrieved chunks
	contextStr := buildOptimizedContext(chunksWithScores, chatReq.Message, queryType)
	
	// Check if context is empty
	if len(contextStr) < 50 { // Very short context likely means no chunks
		log.Printf("Warning: Context is empty or too short. Falling back to direct prompt.")
		responseText, err := handleChatWithoutRAG(apiKey, useOllama, chatReq)
		if err != nil {
			sendErrorResponse(w, "Internal server error", fmt.Sprintf("Failed to process chat: %v", err), http.StatusInternalServerError)
			return
		}
		sendSuccessResponse(w, responseText)
		return
	}
	
	chunkCount := len(chunksWithScores) // Count chunks
	
	// Create enhanced system prompt with few-shot examples
	systemPrompt := buildEnhancedSystemPrompt(queryType)
	
	// Build user message with context
	userMessage := fmt.Sprintf("%s\n\n=== USER QUESTION ===\n%s", contextStr, chatReq.Message)

	log.Printf("Context built: %d chunks, total length: %d characters", chunkCount, len(contextStr))

	// Log first 200 chars of context for debugging
	if len(contextStr) > 200 {
		log.Printf("Context preview: %s...", contextStr[:200])
	} else {
		log.Printf("Context preview: %s", contextStr)
	}

	var responseText string
	var apiErr error

	if useOllama {
		// Use Ollama API with RAG context
		log.Printf("Calling Ollama API with RAG context...")
		responseText, apiErr = callOllamaAPIRAG(systemPrompt, userMessage, chatReq.ConversationHistory)
		if apiErr != nil {
			log.Printf("Error calling Ollama API: %v", apiErr)
			// Fallback to direct prompt without RAG
			log.Printf("Falling back to direct prompt (without RAG)...")
			responseText, apiErr = handleChatWithoutRAG(apiKey, useOllama, chatReq)
			if apiErr != nil {
				sendErrorResponse(w, "Internal server error", fmt.Sprintf("Failed to connect to Ollama service: %v", apiErr), http.StatusInternalServerError)
				return
			}
		}
	} else {
		// Use Gemini API with RAG context
		log.Printf("Calling Gemini API with RAG context...")
		responseText, apiErr = callGeminiAPIRAG(apiKey, systemPrompt, userMessage, chatReq.ConversationHistory)
		if apiErr != nil {
			log.Printf("Error calling Gemini API: %v", apiErr)
			// If Gemini fails, try Ollama as fallback
			log.Printf("Falling back to Ollama API...")
			responseText, apiErr = callOllamaAPIRAG(systemPrompt, userMessage, chatReq.ConversationHistory)
			if apiErr != nil {
				// Final fallback to direct prompt
				log.Printf("Falling back to direct prompt (without RAG)...")
				responseText, apiErr = handleChatWithoutRAG(apiKey, true, chatReq)
				if apiErr != nil {
					sendErrorResponse(w, "Internal server error", fmt.Sprintf("Failed to connect to AI service: %v", apiErr), http.StatusInternalServerError)
					return
				}
			}
		}
	}

	// Ensure we have a response
	if responseText == "" {
		log.Printf("Warning: Empty response from AI service. Sending fallback message.")
		responseText = "I apologize, but I couldn't generate a response. Please try again."
	}

	// Validate answer against context
	validator := rag.NewAnswerValidator()
	isValid, reason := validator.ValidateAnswer(responseText, contextStr, chatReq.Message)
	if !isValid {
		log.Printf("Warning: Answer validation failed: %s. Response may be inaccurate.", reason)
		// Continue anyway, but log the warning
	} else {
		log.Printf("Answer validated successfully: %s", reason)
	}

	log.Printf("Successfully generated response (length: %d)", len(responseText))
	sendSuccessResponse(w, responseText)
}

// buildOptimizedContext builds context with better organization
func buildOptimizedContext(chunksWithScores []rag.RetrievedChunk, query string, queryType rag.QueryType) string {
	contextBuilder := strings.Builder{}
	contextBuilder.WriteString("=== BANK STATEMENT CONTEXT ===\n\n")
	contextBuilder.WriteString("The following information is extracted from the bank statement:\n\n")
	
	// Group chunks by type for better organization
	chunksByType := make(map[string][]rag.RetrievedChunk)
	for _, cws := range chunksWithScores {
		chunkType := "Information"
		if cws.Chunk.Metadata != nil {
			if t, ok := cws.Chunk.Metadata["type"].(string); ok {
				chunkType = t
			}
		}
		chunksByType[chunkType] = append(chunksByType[chunkType], cws)
	}
	
	// Order types by relevance to query
	typeOrder := getRelevantTypeOrder(query, chunksByType, queryType)
	
	chunkCount := 0
	for _, chunkType := range typeOrder {
		if chunks, ok := chunksByType[chunkType]; ok {
			contextBuilder.WriteString(fmt.Sprintf("\n--- %s ---\n", formatChunkType(chunkType)))
			
			for _, cws := range chunks {
				chunkCount++
				contextBuilder.WriteString(fmt.Sprintf("\n[%d] Relevance: %.1f%%\n",
					chunkCount, cws.Score*100))
				
				// Truncate very long chunks (keep first 600 chars)
				content := cws.Chunk.Content
				if len(content) > 600 {
					content = content[:600] + "\n[... content truncated for brevity ...]"
				}
				
				contextBuilder.WriteString(content)
				contextBuilder.WriteString("\n")
			}
		}
	}
	
	contextBuilder.WriteString("\n=== END OF CONTEXT ===\n")
	return contextBuilder.String()
}

// getRelevantTypeOrder returns chunk types ordered by relevance to query
func getRelevantTypeOrder(query string, chunksByType map[string][]rag.RetrievedChunk, queryType rag.QueryType) []string {
	queryLower := strings.ToLower(query)
	
	// Priority order based on query type
	switch queryType {
	case rag.QueryTypeCalculation:
		if strings.Contains(queryLower, "total") || strings.Contains(queryLower, "expense") {
			return []string{"account_summary", "transaction_breakdown", "monthly_summary", "category_summary", "transactions"}
		}
		return []string{"account_summary", "transaction_breakdown", "monthly_summary"}
	case rag.QueryTypeComparison:
		return []string{"monthly_summary", "category_summary", "transactions", "account_summary"}
	case rag.QueryTypeListing:
		return []string{"top_expenses", "top_beneficiaries", "category_summary", "transactions"}
	case rag.QueryTypeTrend:
		return []string{"monthly_summary", "transactions", "category_summary", "account_summary"}
	default:
		// Order by average relevance score
		typeScores := make(map[string]float32)
		for t, chunks := range chunksByType {
			sum := float32(0)
			for _, cws := range chunks {
				sum += cws.Score
			}
			if len(chunks) > 0 {
				typeScores[t] = sum / float32(len(chunks))
			}
		}
		
		// Sort types by score (simple implementation)
		typeList := []string{}
		for t := range chunksByType {
			typeList = append(typeList, t)
		}
		
		// Simple sort by score (bubble sort for small lists)
		for i := 0; i < len(typeList)-1; i++ {
			for j := i + 1; j < len(typeList); j++ {
				if typeScores[typeList[i]] < typeScores[typeList[j]] {
					typeList[i], typeList[j] = typeList[j], typeList[i]
				}
			}
		}
		
		return typeList
	}
}

// buildEnhancedSystemPrompt builds system prompt with few-shot examples
func buildEnhancedSystemPrompt(queryType rag.QueryType) string {
	basePrompt := `You are a helpful financial assistant analyzing bank statement data.
Answer the user's question based ONLY on the provided context above. 

IMPORTANT RULES:
1. Use ONLY the information provided in the context - do not make up or assume any data
2. If the answer is not in the context, say "I don't have that information in the statement data"
3. Be precise with numbers - use the exact amounts from the context
4. Use currency symbols (₹) where appropriate
5. Format numbers in Indian numbering system (e.g., ₹1,00,000 instead of ₹100000)
6. Reference specific sections from the context when answering (use chunk numbers like [1], [2])
7. Be conversational but accurate
8. When calculating, show your reasoning step by step
9. Always cite the source chunk number when providing specific numbers`

	// Add few-shot examples based on query type
	examples := getFewShotExamples(queryType)
	
	return basePrompt + "\n\n" + examples
}

// getFewShotExamples returns few-shot examples for the query type
func getFewShotExamples(queryType rag.QueryType) string {
	examples := map[rag.QueryType]string{
		rag.QueryTypeCalculation: `
EXAMPLE 1 - Calculation Question:
Context: [1] Account Summary: Total Expense: ₹10,76,240.21, Total Income: ₹12,00,000.00
Question: What is my net savings?
Answer: Based on the Account Summary [1], your net savings is calculated as:
Total Income: ₹12,00,000.00
Total Expense: ₹10,76,240.21
Net Savings = ₹12,00,000.00 - ₹10,76,240.21 = ₹1,23,759.79

EXAMPLE 2 - Specific Value Question:
Context: [1] Top Expenses: CRED CLUB: ₹2,17,646.22
Question: What was my highest expense?
Answer: According to the Top Expenses section [1], your highest expense was ₹2,17,646.22 to CRED CLUB.`,

		rag.QueryTypeComparison: `
EXAMPLE - Comparison Question:
Context: [1] Monthly Summary: April: ₹50,000, May: ₹75,000
Question: Which month had higher expenses?
Answer: Based on the Monthly Summary [1], May had higher expenses (₹75,000) compared to April (₹50,000).`,

		rag.QueryTypeListing: `
EXAMPLE - Listing Question:
Context: [1] Top Expenses: Item 1: ₹10,000, Item 2: ₹8,000
Question: What are my top expenses?
Answer: According to the Top Expenses section [1], your top expenses are:
1. Item 1: ₹10,000
2. Item 2: ₹8,000`,

		rag.QueryTypeTrend: `
EXAMPLE - Trend Question:
Context: [1] Monthly Summary: April: ₹50,000, May: ₹60,000, June: ₹70,000
Question: How did my expenses change?
Answer: Based on the Monthly Summary [1], your expenses show an increasing trend:
- April: ₹50,000
- May: ₹60,000 (20% increase)
- June: ₹70,000 (16.7% increase from May)`,
	}
	
	if example, ok := examples[queryType]; ok {
		return "EXAMPLES:\n" + example
	}
	return "EXAMPLES:\n" + examples[rag.QueryTypeCalculation] // Default
}

// formatChunkType formats chunk type for better readability
func formatChunkType(chunkType string) string {
	typeMap := map[string]string{
		"account_summary":       "Account Summary",
		"transaction_breakdown": "Transaction Breakdown by Payment Method",
		"transactions":          "Transaction Details",
		"top_expenses":          "Top Expenses",
		"monthly_summary":       "Monthly Summary",
		"category_summary":      "Category-wise Expenses",
		"top_beneficiaries":     "Top Beneficiaries",
	}

	if formatted, ok := typeMap[chunkType]; ok {
		return formatted
	}
	return strings.Title(strings.ReplaceAll(chunkType, "_", " "))
}

// Helper functions

// generateSourceID generates a unique source ID for statement data using SHA256
func generateSourceID(statementData interface{}) string {
	jsonData, _ := json.Marshal(statementData)
	hash := sha256.Sum256(jsonData)
	// Use first 16 bytes of hash for shorter ID
	return fmt.Sprintf("stmt_%x", hash[:16])
}

// sendSuccessResponse sends a successful chat response
func sendSuccessResponse(w http.ResponseWriter, responseText string) {
	chatResp := ChatResponse{
		Success:   true,
		Response:  responseText,
		Timestamp: getCurrentTimestamp(),
	}

	// Set status code before encoding
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(chatResp); err != nil {
		log.Printf("Error encoding response: %v", err)
		// Try to send a simple error response
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success":false,"error":"encoding_error","message":"Failed to encode response"}`)
		return
	}

	log.Printf("Response sent successfully to client (response length: %d)", len(responseText))
}

// handleChatWithoutRAG handles chat without RAG (fallback method)
func handleChatWithoutRAG(apiKey string, useOllama bool, chatReq ChatRequest) (string, error) {
	// Convert statement data to JSON string
	statementDataJSON, err := json.MarshalIndent(chatReq.StatementData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to process statement data: %w", err)
	}

	// Create system prompt for AI assistant (old method - sends full document)
	systemPrompt := fmt.Sprintf(`You are a helpful financial assistant analyzing bank statement data. 
Answer the user's question based on the provided statement data. Be concise, friendly, and data-driven.

Statement Data:
%s

Instructions:
- Use currency symbols (₹) where appropriate
- Reference specific amounts, dates, and categories from the data
- Keep responses conversational but informative
- If the question cannot be answered from the data, politely say so
- Format numbers in Indian numbering system (e.g., ₹1,00,000 instead of ₹100000)
- Be accurate and cite specific numbers from the data when possible`, string(statementDataJSON))

	if useOllama {
		return callOllamaAPI(systemPrompt, chatReq.Message, chatReq.ConversationHistory)
	}
	return callGeminiAPI(apiKey, systemPrompt, chatReq.Message, chatReq.ConversationHistory)
}

// callOllamaAPIRAG calls Ollama API with RAG context (no conversation history in context)
func callOllamaAPIRAG(systemPrompt, userMessage string, conversationHistory []ConversationMessage) (string, error) {
	// Get model from RAG config
	ragMgr, err := getRAGManager()
	chatModel := "llama3" // Default
	if err == nil && ragMgr != nil {
		chatModel = ragMgr.GetConfig().ChatModel
	}

	var messages []OllamaMessage

	// Add system message
	messages = append(messages, OllamaMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	// Add conversation history (but NOT in the context - only previous Q&A)
	for _, msg := range conversationHistory {
		role := msg.Role
		if role == "assistant" {
			role = "assistant"
		} else {
			role = "user"
		}
		messages = append(messages, OllamaMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Add current user message with RAG context
	messages = append(messages, OllamaMessage{
		Role:    "user",
		Content: userMessage,
	})

	return callOllamaWithMessages(messages, chatModel)
}

// callGeminiAPIRAG calls Gemini API with RAG context
func callGeminiAPIRAG(apiKey, systemPrompt, userMessage string, conversationHistory []ConversationMessage) (string, error) {
	// Build Gemini API request with conversation history
	var contents []GeminiContent

	// Add conversation history if available
	if len(conversationHistory) > 0 {
		for _, msg := range conversationHistory {
			role := "user"
			if msg.Role == "assistant" {
				role = "model"
			}
			contents = append(contents, GeminiContent{
				Parts: []GeminiPart{{Text: msg.Content}},
				Role:  role,
			})
		}
	}

	// Add system context and current message with RAG context
	fullPrompt := fmt.Sprintf("%s\n\n%s", systemPrompt, userMessage)
	contents = append(contents, GeminiContent{
		Parts: []GeminiPart{{Text: fullPrompt}},
		Role:  "user",
	})

	geminiReq := GeminiRequest{
		Contents: contents,
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return "", fmt.Errorf("failed to prepare request: %w", err)
	}

	// Call Gemini API
	geminiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent?key=%s", apiKey)
	resp, err := http.Post(geminiURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to connect to Gemini API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Gemini response: %w", err)
	}

	// Parse Gemini response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	// Handle Gemini API errors
	if geminiResp.Error != nil {
		errorMsg := geminiResp.Error.Message
		if strings.Contains(errorMsg, "API key not valid") || strings.Contains(errorMsg, "invalid API key") {
			return "", fmt.Errorf("invalid API key: %s", errorMsg)
		}
		if strings.Contains(errorMsg, "429") || strings.Contains(errorMsg, "quota") || strings.Contains(errorMsg, "Quota exceeded") {
			return "", fmt.Errorf("rate limit exceeded: %s", errorMsg)
		}
		return "", fmt.Errorf("Gemini API error: %s", errorMsg)
	}

	// Extract response text
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("AI service returned an empty response")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// callOllamaWithMessages is a helper to call Ollama with a message array
func callOllamaWithMessages(messages []OllamaMessage, model string) (string, error) {
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	ollamaReq := OllamaRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Options: map[string]interface{}{
			"temperature":    0.0,  // Deterministic output
			"top_p":          0.9,  // Nucleus sampling (focus on top tokens)
			"top_k":          40,   // Limit vocabulary
			"repeat_penalty": 1.1,  // Reduce repetition
			"num_predict":    512,  // Max tokens in response
		},
	}

	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return "", fmt.Errorf("failed to prepare Ollama request: %w", err)
	}

	url := ollamaURL + "/api/chat"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 120 * time.Second, // 2 minute timeout for chat requests
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("Ollama API error: %s", ollamaResp.Error)
	}

	if ollamaResp.Message.Content == "" {
		return "", fmt.Errorf("Ollama returned an empty response")
	}

	return ollamaResp.Message.Content, nil
}

// sendErrorResponse sends an error response
func sendErrorResponse(w http.ResponseWriter, errorType, message string, statusCode int) {
	chatResp := ChatResponse{
		Success:   false,
		Error:     errorType,
		Message:   message,
		Timestamp: getCurrentTimestamp(),
	}

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(chatResp); err != nil {
		log.Printf("Error encoding error response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// callGeminiAPI calls the Gemini API and returns the response text
func callGeminiAPI(apiKey, systemPrompt, userMessage string, conversationHistory []ConversationMessage) (string, error) {
	// Build conversation context
	var conversationContext strings.Builder
	if len(conversationHistory) > 0 {
		conversationContext.WriteString("\n\nPrevious conversation:\n")
		for _, msg := range conversationHistory {
			roleLabel := "User"
			if msg.Role == "assistant" {
				roleLabel = "Assistant"
			}
			conversationContext.WriteString(fmt.Sprintf("%s: %s\n", roleLabel, msg.Content))
		}
	}

	// Create prompt for Gemini
	prompt := fmt.Sprintf(`%s
%s

User Question: "%s"`, systemPrompt, conversationContext.String(), userMessage)

	// Build Gemini API request with conversation history
	var contents []GeminiContent

	// Add conversation history if available
	if len(conversationHistory) > 0 {
		for _, msg := range conversationHistory {
			role := "user"
			if msg.Role == "assistant" {
				role = "model"
			}
			contents = append(contents, GeminiContent{
				Parts: []GeminiPart{{Text: msg.Content}},
				Role:  role,
			})
		}
	}

	// Add system context and current message as a single user message
	contents = append(contents, GeminiContent{
		Parts: []GeminiPart{{Text: prompt}},
		Role:  "user",
	})

	geminiReq := GeminiRequest{
		Contents: contents,
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return "", fmt.Errorf("failed to prepare request: %w", err)
	}

	// Call Gemini API
	geminiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent?key=%s", apiKey)
	resp, err := http.Post(geminiURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to connect to Gemini API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Gemini response: %w", err)
	}

	// Parse Gemini response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	// Handle Gemini API errors
	if geminiResp.Error != nil {
		errorMsg := geminiResp.Error.Message
		if strings.Contains(errorMsg, "API key not valid") || strings.Contains(errorMsg, "invalid API key") {
			return "", fmt.Errorf("invalid API key: %s", errorMsg)
		}
		if strings.Contains(errorMsg, "429") || strings.Contains(errorMsg, "quota") || strings.Contains(errorMsg, "Quota exceeded") {
			return "", fmt.Errorf("rate limit exceeded: %s", errorMsg)
		}
		return "", fmt.Errorf("Gemini API error: %s", errorMsg)
	}

	// Extract response text
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("AI service returned an empty response")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// callOllamaAPI calls the Ollama API and returns the response text
func callOllamaAPI(systemPrompt, userMessage string, conversationHistory []ConversationMessage) (string, error) {
	// Get model from RAG config or environment
	chatModel := os.Getenv("OLLAMA_CHAT_MODEL")
	if chatModel == "" {
		// Try to get from RAG config
		ragMgr, err := getRAGManager()
		if err == nil && ragMgr != nil {
			chatModel = ragMgr.GetConfig().ChatModel
		} else {
			chatModel = "llama3" // Default fallback
		}
	}

	// Build messages array for Ollama
	var messages []OllamaMessage

	// Add system message
	messages = append(messages, OllamaMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	// Add conversation history
	for _, msg := range conversationHistory {
		role := msg.Role
		if role == "assistant" {
			role = "assistant"
		} else {
			role = "user"
		}
		messages = append(messages, OllamaMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Add current user message
	messages = append(messages, OllamaMessage{
		Role:    "user",
		Content: userMessage,
	})

	// Build Ollama request with model from config
	ollamaReq := OllamaRequest{
		Model:    chatModel,
		Messages: messages,
		Stream:   false,
		Options: map[string]interface{}{
			"temperature":    0.0,  // Deterministic output
			"top_p":          0.9,  // Nucleus sampling (focus on top tokens)
			"top_k":          40,   // Limit vocabulary
			"repeat_penalty": 1.1,  // Reduce repetition
			"num_predict":    512,  // Max tokens in response
		},
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return "", fmt.Errorf("failed to prepare Ollama request: %w", err)
	}

	// Get Ollama URL from environment or use default
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434/api/chat"
	}

	// Call Ollama API
	resp, err := http.Post(ollamaURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ollama API: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Ollama response: %w", err)
	}

	// Parse Ollama response
	var ollamaResp OllamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	// Handle Ollama API errors
	if ollamaResp.Error != "" {
		return "", fmt.Errorf("Ollama API error: %s", ollamaResp.Error)
	}

	// Extract response text
	if ollamaResp.Message.Content == "" {
		return "", fmt.Errorf("Ollama returned an empty response")
	}

	return ollamaResp.Message.Content, nil
}

// getCurrentTimestamp returns current timestamp in ISO 8601 format
func getCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// healthHandler handles GET requests to /api/health
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	enableCORS(w, r)

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	healthResp := map[string]interface{}{
		"status":    "healthy",
		"service":   "chat-api",
		"timestamp": getCurrentTimestamp(),
	}

	if err := json.NewEncoder(w).Encode(healthResp); err != nil {
		log.Printf("Error encoding health response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
