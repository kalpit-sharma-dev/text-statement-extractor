# RAG Implementation Complete Guide
## Chat Integration API - Retrieval-Augmented Generation System

**Version:** 1.0  
**Last Updated:** December 2025  
**Purpose:** Complete documentation of the RAG implementation for financial statement chat assistant

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Problem Statement](#problem-statement)
3. [Architecture Overview](#architecture-overview)
4. [System Components](#system-components)
5. [Data Flow](#data-flow)
6. [Implementation Details](#implementation-details)
7. [Configuration](#configuration)
8. [Setup & Deployment](#setup--deployment)
9. [API Integration](#api-integration)
10. [Performance & Optimization](#performance--optimization)
11. [Error Handling & Fallbacks](#error-handling--fallbacks)
12. [Testing & Validation](#testing--validation)
13. [Troubleshooting](#troubleshooting)
14. [Future Enhancements](#future-enhancements)

---

## Executive Summary

### What is RAG?

**Retrieval-Augmented Generation (RAG)** is a technique that enhances LLM responses by:
1. **Chunking** large documents into smaller, manageable pieces
2. **Embedding** each chunk into a vector space
3. **Storing** embeddings in a vector database
4. **Retrieving** only relevant chunks based on user queries
5. **Sending** only retrieved chunks to the LLM (never the full document)

### Why RAG for This System?

**Problem:** Bank statements can contain thousands of transactions. Sending the entire statement JSON to an LLM causes:
- **Context truncation** (exceeds token limits)
- **Irrelevant information** overwhelming the model
- **Slower responses** (processing unnecessary data)
- **Higher costs** (more tokens = more cost)

**Solution:** RAG ensures:
- ✅ **No truncation** - Only relevant chunks sent (always within limits)
- ✅ **Relevant context** - Only information related to the query
- ✅ **Faster responses** - Less data to process
- ✅ **Better accuracy** - Focused context leads to better answers
- ✅ **Scalable** - Works with statements of any size

### Key Metrics

- **Chunk Size:** 300-500 tokens per chunk
- **Retrieval:** Top 5 chunks (configurable)
- **Similarity Threshold:** 0.3 (30% minimum relevance)
- **Embedding Dimensions:** 4096 (llama3)
- **Indexing Time:** ~5-10 seconds for first request
- **Query Time:** ~100-500ms (after indexing)

---

## Problem Statement

### Original Challenge

When users ask questions about their bank statements:
- **Before RAG:** Entire statement JSON sent to LLM
  - Risk of truncation if statement is large
  - Irrelevant transactions included
  - Slower processing
  - Higher token costs

### Example Problem

**User Query:** "What was my total expense?"

**Without RAG:**
```
Send to LLM:
- Account Summary
- All 500 transactions
- Transaction Breakdown
- Monthly Summary
- Category Summary
- Top Expenses
- Top Beneficiaries
- ... (entire statement, 50KB+)
```

**With RAG:**
```
Send to LLM:
- Account Summary chunk (95% relevance)
- Transaction Breakdown chunk (87% relevance)
- Monthly Summary chunk (75% relevance)
(Only 3-5 relevant chunks, ~2KB)
```

### Solution Benefits

1. **No Truncation:** Context always fits within model limits
2. **Relevant Information:** Only chunks related to the query
3. **Faster:** Less data = faster processing
4. **Accurate:** Focused context = better answers
5. **Cost-Effective:** Fewer tokens = lower costs

---

## Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Chat API Request                          │
│              POST /api/chat                                 │
│  { message, statementData, conversationHistory, apiKey }    │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                  Chat Handler                                │
│  - Validates request                                         │
│  - Generates source ID (SHA256 hash)                        │
│  - Initializes RAG Manager                                   │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              RAG Pipeline (First Request)                    │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐ │
│  │   Chunker    │───▶│  Embeddings  │───▶│Vector Store  │ │
│  │              │    │              │    │              │ │
│  │ - Semantic   │    │ - Ollama API │    │ - PostgreSQL │ │
│  │   chunking   │    │ - Concurrent │   │   + pgvector │ │
│  │ - By entity  │    │ - Batch      │    │ - In-memory  │ │
│  └──────────────┘    └──────────────┘    └──────────────┘ │
│                                                              │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              RAG Pipeline (Query Request)                    │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐ │
│  │  Retriever   │───▶│Vector Store  │───▶│Context Build │ │
│  │              │    │              │    │              │ │
│  │ - Embed query│    │ - Cosine     │    │ - Format     │ │
│  │ - Search     │    │   similarity │    │ - Add scores │ │
│  │ - Filter    │    │ - Top-K      │    │ - Add types  │ │
│  └──────────────┘    └──────────────┘    └──────────────┘ │
│                                                              │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                  LLM API Call                               │
│                                                              │
│  System Prompt: "Use ONLY provided context..."              │
│  User Message: Context + Question                           │
│                                                              │
│  ┌──────────────┐         ┌──────────────┐                 │
│  │   Gemini     │   OR    │    Ollama    │                 │
│  │   API        │         │    llama3/4  │                 │
│  └──────────────┘         └──────────────┘                 │
│                                                              │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                  Response to Client                          │
│  { success: true, response: "...", timestamp: "..." }        │
└─────────────────────────────────────────────────────────────┘
```

### Component Interaction

```
Chat Handler
    │
    ├──► RAG Manager (Singleton)
    │       │
    │       ├──► Chunker
    │       │       └──► Semantic chunking by entity
    │       │
    │       ├──► Embedding Client
    │       │       └──► Ollama /api/embeddings
    │       │
    │       ├──► Vector Store
    │       │       ├──► PostgreSQL + pgvector (Primary)
    │       │       └──► In-Memory (Fallback)
    │       │
    │       └──► Retriever
    │               └──► Cosine similarity search
    │
    └──► LLM API (Gemini or Ollama)
            └──► Generate response from context
```

---

## System Components

### 1. Configuration (`rag/config.go`)

**Purpose:** Centralized configuration management

**Key Fields:**
```go
type Config struct {
    // Ollama configuration
    OllamaURL      string  // Default: http://localhost:11434
    EmbeddingModel string  // Default: llama3 (from OLLAMA_EMBEDDING_MODEL)
    ChatModel      string  // Default: llama3 (from OLLAMA_CHAT_MODEL)
    
    // Chunking configuration
    ChunkSize    int     // Target: 400 tokens per chunk
    ChunkOverlap float64 // 12% overlap between chunks
    
    // Retrieval configuration
    TopK                int     // Number of chunks to retrieve (default: 5)
    SimilarityThreshold float32 // Minimum similarity (default: 0.3)
    
    // Vector store configuration
    UsePostgreSQL bool   // Auto-detected
    PostgreSQLDSN string // From POSTGRES_DSN env var
    TableName     string // Default: statement_chunks
    
    // Embedding dimensions
    EmbeddingDims int    // 4096 for llama3
}
```

**Environment Variables:**
- `OLLAMA_URL` - Ollama server URL (default: http://localhost:11434)
- `OLLAMA_EMBEDDING_MODEL` - Model for embeddings (default: llama3)
- `OLLAMA_CHAT_MODEL` - Model for chat (default: llama3)
- `POSTGRES_DSN` - PostgreSQL connection string (optional)

**Default Values:**
- Chunk size: 400 tokens
- Chunk overlap: 12%
- Top K: 5 chunks
- Similarity threshold: 0.3 (30%)
- Embedding dimensions: 4096

---

### 2. Chunker (`rag/chunker.go`)

**Purpose:** Semantic chunking of bank statement data

#### Chunking Strategy

**Critical Principle:** Chunk by **semantic unit** (transaction/entity), NOT by raw characters.

#### Chunk Types Created

1. **Account Summary** (1 chunk)
   - Account number, customer name
   - Opening/closing balance
   - Total income, expense, investments
   - Net savings, savings rate

2. **Transaction Breakdown** (1 chunk)
   - Breakdown by payment method (UPI, IMPS, NEFT, etc.)
   - Amount and count for each method

3. **Individual Transactions** (Multiple chunks)
   - Grouped in batches of ~6 transactions per chunk
   - Each transaction: date, amount, type, category, merchant, method

4. **Top Expenses** (1 chunk)
   - Top 10 expenses with merchant, amount, category, date

5. **Monthly Summary** (1 chunk)
   - Monthly income, expense, balance
   - Top category per month
   - Expense spike percentage

6. **Category Summary** (1 chunk)
   - Category-wise expense totals
   - Shopping, Bills_Utilities, Travel, Dining, etc.

7. **Top Beneficiaries** (1 chunk)
   - Top 10 beneficiaries with name, amount, type

#### Chunk Structure

```go
type Chunk struct {
    ID        string                 // Unique identifier: "stmt_xxx_summary_0"
    SourceID  string                 // Statement identifier: "stmt_xxx"
    Content   string                 // Formatted text content
    Embedding []float32              // Vector embedding (4096 dimensions)
    Metadata  map[string]interface{} // { "type": "account_summary" }
}
```

#### Chunking Logic

**Transaction Chunking:**
- Target: ~400 tokens per chunk
- Group size: ~6 transactions per chunk
- Format: "Date: DD/MM/YYYY | Amount: ₹X.XX | Type: Debit | Category: X | Merchant: Y | Method: Z"

**Example Chunk:**
```
Account Summary:
Account: XXXX1725
Customer: John Doe
Period: 01/04/2025 - 17/12/2025
Opening Balance: ₹50,000.00
Closing Balance: ₹1,50,000.00
Total Income: ₹12,00,000.00
Total Expense: ₹10,76,240.21
Total Investments: ₹1,00,000.00
Net Savings: ₹1,23,759.79
Savings Rate: 10.31%
```

#### Why Semantic Chunking?

- **Preserves Context:** Related data stays together
- **Better Retrieval:** Semantic units are more meaningful
- **Accurate Answers:** Complete information per chunk
- **Efficient:** No partial transactions split across chunks

---

### 3. Embedding Client (`rag/embeddings.go`)

**Purpose:** Generate vector embeddings for chunks and queries

#### Embedding Process

1. **Single Embedding Generation:**
   ```go
   GenerateEmbedding(text string) ([]float32, error)
   ```
   - Calls Ollama `/api/embeddings` endpoint
   - Model: `llama3` (configurable)
   - Returns: 4096-dimensional vector

2. **Batch Embedding Generation:**
   ```go
   GenerateEmbeddingsBatch(texts []string) ([][]float32, error)
   ```
   - **Concurrent processing** with goroutines
   - **Semaphore pattern** (max 5 concurrent requests)
   - **Error handling:** Continues even if some fail
   - **Logging:** Progress tracking for each embedding

3. **Chunk Embedding:**
   ```go
   EmbedChunks(chunks []*Chunk) error
   ```
   - Generates embeddings for all chunks
   - Assigns embeddings to chunks
   - Filters out chunks without embeddings

#### HTTP Client Configuration

```go
client: &http.Client{
    Timeout: 60 * time.Second, // Prevents hanging
}
```

#### Ollama API Request Format

```json
{
  "model": "llama3",
  "prompt": "Account Summary:\nTotal Expense: ₹10,76,240.21..."
}
```

#### Ollama API Response Format

```json
{
  "embedding": [0.123, -0.456, 0.789, ...] // 4096 float64 values
}
```

#### Performance Optimizations

- **Concurrent Processing:** 5 parallel requests
- **Timeout Protection:** 60-second timeout
- **Error Resilience:** Continues on individual failures
- **Progress Logging:** Tracks embedding generation

---

### 4. Vector Store (`rag/vectorstore/`)

**Purpose:** Store and search vector embeddings

#### Architecture

Two implementations with automatic fallback:

1. **PostgreSQL + pgvector** (Primary)
   - Production-ready
   - Persistent storage
   - Efficient similarity search
   - Scalable

2. **In-Memory Store** (Fallback)
   - No dependencies
   - Fast for small datasets
   - Lost on restart
   - Development/testing

#### Vector Store Interface

```go
type VectorStore interface {
    Store(chunk *Chunk) error
    StoreBatch(chunks []*Chunk) error
    Search(queryEmbedding []float32, topK int, sourceID string) ([]*Chunk, []float32, error)
    DeleteBySourceID(sourceID string) error
    Count() int
}
```

#### PostgreSQL Implementation (`pgvector.go`)

**Initialization:**
1. Connect to PostgreSQL
2. Check if `pgvector` extension exists
3. Create table if not exists
4. Create indexes (IVFFlat for similarity, source_id for filtering)

**Table Schema:**
```sql
CREATE TABLE statement_chunks (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(4096),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Indexes:**
- `ivfflat` index on embedding (cosine similarity)
- B-tree index on `source_id` (filtering)
- B-tree index on `created_at` (cleanup)

**Search Query:**
```sql
SELECT id, source_id, content, metadata,
       1 - (embedding <=> $1::vector) as similarity
FROM statement_chunks
WHERE source_id = $2
ORDER BY embedding <=> $1::vector
LIMIT 5;
```

**Auto-Detection:**
- Checks PostgreSQL connection
- Verifies pgvector extension
- Falls back to in-memory if unavailable
- Logs warning on fallback

#### In-Memory Implementation (`memory.go`)

**Storage:**
- Map-based storage: `map[string]*Chunk`
- Thread-safe with `sync.RWMutex`
- Fast for small datasets (< 10K chunks)

**Search Algorithm:**
1. Iterate all chunks
2. Filter by `sourceID` if provided
3. Calculate cosine similarity for each
4. Sort by score (descending)
5. Return top K

**Cosine Similarity:**
```go
cosineSimilarity(a, b []float32) float32 {
    dotProduct = Σ(a[i] * b[i])
    normA = √Σ(a[i]²)
    normB = √Σ(b[i]²)
    return dotProduct / (normA * normB)
}
```

**Performance:**
- O(n) search complexity
- Suitable for < 10K chunks
- No persistence (lost on restart)

---

### 5. Retriever (`rag/retriever.go`)

**Purpose:** Retrieve relevant chunks based on user queries

#### Retrieval Process

1. **Query Embedding:**
   - Generate embedding for user query
   - Same model and dimensions as chunk embeddings

2. **Similarity Search:**
   - Search vector store with query embedding
   - Filter by `sourceID` (statement-specific)
   - Get top K chunks with scores

3. **Sorting:**
   - Sort chunks by similarity score (descending)
   - Most relevant chunks first

4. **Filtering:**
   - Apply similarity threshold (default: 0.3)
   - Filter out low-relevance chunks
   - Always return at least top chunk (if available)

#### RetrievedChunk Structure

```go
type RetrievedChunk struct {
    Chunk *Chunk   // The chunk data
    Score float32  // Similarity score (0.0 to 1.0)
}
```

#### Methods

1. **Retrieve(query, sourceID):**
   - Returns chunks only (no scores)
   - Used for simple retrieval

2. **RetrieveWithScores(query, sourceID):**
   - Returns chunks with similarity scores
   - Used for context building with relevance info

#### Similarity Threshold

- **Default:** 0.3 (30% similarity)
- **Purpose:** Filter out irrelevant chunks
- **Behavior:** If all chunks filtered, return top chunk anyway
- **Configurable:** Via `SimilarityThreshold` in config

---

### 6. Manager (`rag/manager.go`)

**Purpose:** Orchestrates the entire RAG pipeline

#### Responsibilities

1. **Initialization:**
   - Creates chunker, embedding client, vector store, retriever
   - Auto-detects PostgreSQL availability
   - Falls back gracefully

2. **Indexing:**
   - Chunks statement data
   - Generates embeddings
   - Stores in vector store
   - Handles existing chunks (replacement)

3. **Retrieval:**
   - Retrieves relevant chunks for queries
   - Returns chunks with scores

4. **Lifecycle Management:**
   - Thread-safe operations
   - Chunk existence checking
   - Chunk deletion

#### Key Methods

1. **IndexStatementData(statementData, sourceID):**
   - Main indexing entry point
   - Chunk → Embed → Store pipeline
   - Replaces existing chunks if present

2. **RetrieveRelevantChunks(query, sourceID):**
   - Retrieves chunks for query
   - Returns chunks only

3. **RetrieveRelevantChunksWithScores(query, sourceID):**
   - Retrieves chunks with similarity scores
   - Used for context building

4. **HasChunks(sourceID):**
   - Checks if chunks exist for source
   - Used to avoid re-indexing

5. **DeleteStatementData(sourceID):**
   - Deletes all chunks for a source
   - Used for cleanup/replacement

#### Thread Safety

- **Read-Write Mutex:** `sync.RWMutex`
- **Read operations:** Use `RLock()` (allows concurrent reads)
- **Write operations:** Use `Lock()` (exclusive)
- **Deadlock Prevention:** `HasChunks()` called before acquiring write lock

---

### 7. Chat Handler (`chat_handler.go`)

**Purpose:** HTTP handler for chat API endpoint

#### Request Flow

1. **Request Validation:**
   - Validate HTTP method (POST only)
   - Parse JSON request body
   - Validate required fields (message, statementData)

2. **Source ID Generation:**
   - Generate unique source ID using SHA256 hash
   - Format: `stmt_{first_16_bytes_of_hash}`
   - Enables chunk reuse for same statements

3. **RAG Manager Initialization:**
   - Get or create singleton RAG manager
   - Initialize on first request

4. **Chunk Existence Check:**
   - Check if chunks already exist for source ID
   - Skip indexing if chunks exist

5. **Indexing (if needed):**
   - Chunk statement data
   - Generate embeddings
   - Store in vector store
   - Log progress at each step

6. **Retrieval:**
   - Retrieve relevant chunks with scores
   - Handle empty results (fallback)

7. **Context Building:**
   - Build structured context from chunks
   - Include chunk types and relevance scores
   - Format for LLM consumption

8. **LLM API Call:**
   - Build system prompt
   - Build user message with context
   - Call Gemini or Ollama API
   - Handle errors with fallbacks

9. **Response:**
   - Format JSON response
   - Return to client

#### Context Building Logic

**Context Format:**
```
=== BANK STATEMENT CONTEXT ===

The following information is extracted from the bank statement:

[1] Account Summary (Relevance: 95.2%):
Account Summary:
Total Expense: ₹10,76,240.21
...

[2] Transaction Breakdown by Payment Method (Relevance: 87.5%):
Transaction Breakdown by Payment Method:
UPI: ₹5,00,000 (100 transactions)
...

=== END OF CONTEXT ===

=== USER QUESTION ===
What was my total expense?
```

**Key Features:**
- Clear section headers
- Numbered chunks
- Chunk type labels
- Relevance scores (as percentages)
- Sorted by relevance (most relevant first)

#### System Prompt

```
You are a helpful financial assistant analyzing bank statement data.
Answer the user's question based ONLY on the provided context above. 

IMPORTANT RULES:
1. Use ONLY the information provided in the context - do not make up or assume any data
2. If the answer is not in the context, say "I don't have that information in the statement data"
3. Be precise with numbers - use the exact amounts from the context
4. Use currency symbols (₹) where appropriate
5. Format numbers in Indian numbering system (e.g., ₹1,00,000 instead of ₹100000)
6. Reference specific sections from the context when answering
7. Be conversational but accurate
```

#### Fallback Mechanisms

1. **RAG Initialization Fails:**
   - Fallback to direct prompt (old method)

2. **Indexing Fails:**
   - Fallback to direct prompt

3. **Retrieval Fails:**
   - Fallback to direct prompt

4. **Empty Chunks:**
   - Fallback to direct prompt

5. **LLM API Fails:**
   - Gemini fails → Try Ollama
   - Ollama fails → Try direct prompt

**All fallbacks are logged with warnings.**

---

## Data Flow

### First Request (Indexing + Query)

```
1. User sends chat request
   ↓
2. Chat Handler validates request
   ↓
3. Generate source ID (SHA256 hash)
   ↓
4. Initialize RAG Manager
   ↓
5. Check if chunks exist → NO
   ↓
6. INDEXING PHASE:
   ├─► Chunker: Split statement into semantic chunks
   │   └─► Creates 10-20 chunks (account summary, transactions, etc.)
   │
   ├─► Embedding Client: Generate embeddings
   │   ├─► For each chunk: Call Ollama /api/embeddings
   │   ├─► Concurrent processing (5 at a time)
   │   └─► Store embeddings in chunks
   │
   └─► Vector Store: Store chunks
       ├─► Try PostgreSQL + pgvector
       └─► Fallback to in-memory if unavailable
   ↓
7. RETRIEVAL PHASE:
   ├─► Embedding Client: Generate query embedding
   │   └─► Call Ollama /api/embeddings with user query
   │
   ├─► Vector Store: Search similar chunks
   │   ├─► Calculate cosine similarity
   │   ├─► Sort by score
   │   └─► Return top 5 chunks
   │
   └─► Retriever: Filter by threshold
       └─► Return chunks with scores >= 0.3
   ↓
8. CONTEXT BUILDING:
   ├─► Format chunks with types and scores
   ├─► Build structured context
   └─► Create user message
   ↓
9. LLM API CALL:
   ├─► Build system prompt
   ├─► Send context + question to LLM
   └─► Get response
   ↓
10. Return response to user
```

### Subsequent Requests (Query Only)

```
1. User sends chat request
   ↓
2. Generate source ID (same hash for same statement)
   ↓
3. Check if chunks exist → YES
   ↓
4. SKIP INDEXING (chunks already exist)
   ↓
5. RETRIEVAL PHASE:
   ├─► Generate query embedding
   ├─► Search vector store
   └─► Get relevant chunks
   ↓
6. CONTEXT BUILDING + LLM CALL
   ↓
7. Return response
```

**Performance:** Subsequent requests are 5-10x faster (no indexing needed)

---

## Implementation Details

### Source ID Generation

**Purpose:** Uniquely identify each statement for chunk reuse

**Algorithm:**
```go
func generateSourceID(statementData interface{}) string {
    jsonData, _ := json.Marshal(statementData)
    hash := sha256.Sum256(jsonData)
    return fmt.Sprintf("stmt_%x", hash[:16]) // First 16 bytes
}
```

**Benefits:**
- Same statement → Same source ID
- Chunks reused across multiple queries
- No re-indexing needed
- Collision-resistant (SHA256)

**Example:**
- Statement hash: `45b64c7eea7597dfea3c927bde4b7672`
- Source ID: `stmt_45b64c7eea7597df`

### Chunk ID Format

**Pattern:** `{sourceID}_{type}_{index}`

**Examples:**
- `stmt_xxx_summary_0` - Account summary
- `stmt_xxx_breakdown_1` - Transaction breakdown
- `stmt_xxx_transactions_2` - First transaction batch
- `stmt_xxx_transactions_3` - Second transaction batch
- `stmt_xxx_top_expenses_8` - Top expenses

### Embedding Generation

**Process:**
1. Extract chunk content (text)
2. Call Ollama `/api/embeddings`
3. Convert `[]float64` to `[]float32`
4. Store in chunk

**Concurrency:**
- Max 5 concurrent requests
- Semaphore pattern for rate limiting
- Continues on individual failures

**Error Handling:**
- Logs failed embeddings
- Skips chunks without embeddings
- Returns error if all fail

### Vector Storage

#### PostgreSQL Storage

**Connection:**
- DSN format: `postgres://user:password@host/dbname?sslmode=disable`
- Auto-detected on startup
- Falls back if unavailable

**Table Creation:**
- Auto-creates on first use
- Includes pgvector extension check
- Creates indexes automatically

**Vector Format:**
- PostgreSQL vector type: `vector(4096)`
- Stored as: `[0.123, -0.456, 0.789, ...]`
- Cosine similarity operator: `<=>`

#### In-Memory Storage

**Structure:**
```go
type MemoryVectorStore struct {
    chunks map[string]*Chunk  // ID -> Chunk mapping
    mu     sync.RWMutex       // Thread-safe access
}
```

**Search:**
- Linear search through all chunks
- O(n) complexity
- Fast for < 10K chunks

### Similarity Search

**Algorithm:** Cosine Similarity

**Formula:**
```
similarity = (A · B) / (||A|| × ||B||)

Where:
- A · B = dot product
- ||A|| = magnitude of A
- ||B|| = magnitude of B
```

**Range:** 0.0 to 1.0
- 1.0 = Identical
- 0.0 = Completely different
- > 0.7 = Very similar
- < 0.3 = Low similarity (filtered)

### Context Building

**Structure:**
```
=== BANK STATEMENT CONTEXT ===

The following information is extracted from the bank statement:

[1] Account Summary (Relevance: 95.2%):
{chunk content}

[2] Transaction Breakdown (Relevance: 87.5%):
{chunk content}

=== END OF CONTEXT ===

=== USER QUESTION ===
{user question}
```

**Features:**
- Clear section markers
- Numbered chunks
- Chunk type labels
- Relevance scores
- Sorted by relevance

**Why This Format?**
- LLM can easily identify sections
- Relevance scores help prioritize
- Clear boundaries prevent confusion
- Structured format improves parsing

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_URL` | `http://localhost:11434` | Ollama server URL |
| `OLLAMA_EMBEDDING_MODEL` | `llama3` | Model for embeddings |
| `OLLAMA_CHAT_MODEL` | `llama3` | Model for chat/completion |
| `POSTGRES_DSN` | (none) | PostgreSQL connection string |

### Configuration Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `ChunkSize` | 400 | Target tokens per chunk |
| `ChunkOverlap` | 0.12 | 12% overlap between chunks |
| `TopK` | 5 | Number of chunks to retrieve |
| `SimilarityThreshold` | 0.3 | Minimum similarity score (30%) |
| `EmbeddingDims` | 4096 | Embedding vector dimensions |
| `TableName` | `statement_chunks` | PostgreSQL table name |

### Model Configuration

**Default Models:**
- Embedding: `llama3`
- Chat: `llama3`

**Switching to Llama4:**
```bash
export OLLAMA_CHAT_MODEL="llama4"
export OLLAMA_EMBEDDING_MODEL="llama4"  # Optional
```

**Hybrid Approach (Recommended):**
```bash
export OLLAMA_EMBEDDING_MODEL="llama3"  # Stable embeddings
export OLLAMA_CHAT_MODEL="llama4"       # Better responses
```

---

## Setup & Deployment

### Prerequisites

1. **Go 1.21+**
2. **Ollama** installed and running
3. **PostgreSQL** (optional, for production)
4. **pgvector extension** (optional)

### Step 1: Install Dependencies

```bash
go get github.com/lib/pq
```

### Step 2: Install Ollama

```bash
# Download from https://ollama.ai
# Or use package manager
```

### Step 3: Pull Models

```bash
ollama pull llama3
# Optional: ollama pull llama4
```

### Step 4: Start Ollama

```bash
ollama serve
# Usually runs automatically
```

### Step 5: Verify Ollama

```bash
curl http://localhost:11434/api/tags
# Should return list of available models
```

### Step 6: PostgreSQL Setup (Optional)

```sql
-- Install pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Table will be auto-created by application
```

### Step 7: Set Environment Variables

```bash
# Optional: Custom Ollama URL
export OLLAMA_URL="http://localhost:11434"

# Optional: Custom models
export OLLAMA_CHAT_MODEL="llama4"
export OLLAMA_EMBEDDING_MODEL="llama3"

# Optional: PostgreSQL
export POSTGRES_DSN="postgres://user:password@localhost/dbname?sslmode=disable"
```

### Step 8: Build and Run

```bash
go build -o server.exe example_base64_usage.go chat_handler.go extract_statement.go
./server.exe
```

### Step 9: Verify Endpoints

```bash
# Health check
curl http://localhost:8080/api/health

# Chat endpoint
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"test","statementData":{...}}'
```

---

## API Integration

### Endpoint: `POST /api/chat`

**Request:**
```json
{
  "message": "What was my total expense?",
  "statementData": {
    "accountSummary": { ... },
    "transactions": [ ... ],
    "transactionBreakdown": { ... },
    ...
  },
  "conversationHistory": [
    {
      "role": "user",
      "content": "What was my highest expense?"
    },
    {
      "role": "assistant",
      "content": "Your highest expense was ₹2,17,646.22 to CRED CLUB."
    }
  ],
  "apiKey": "optional-gemini-api-key"
}
```

**Response:**
```json
{
  "success": true,
  "response": "Based on your statement, your total expense was ₹10,76,240.21...",
  "timestamp": "2025-12-21T23:36:00Z"
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "error_type",
  "message": "Error description",
  "timestamp": "2025-12-21T23:36:00Z"
}
```

### Request Flow

1. **Parse Request:**
   - Validate JSON
   - Check required fields
   - Extract message and statement data

2. **Generate Source ID:**
   - SHA256 hash of statement data
   - Format: `stmt_{hash}`

3. **Initialize RAG:**
   - Get or create RAG manager
   - Initialize components

4. **Check Existing Chunks:**
   - Query vector store
   - Skip indexing if chunks exist

5. **Index (if needed):**
   - Chunk → Embed → Store

6. **Retrieve:**
   - Embed query
   - Search vector store
   - Get top K chunks

7. **Build Context:**
   - Format chunks
   - Add metadata
   - Create structured context

8. **Call LLM:**
   - Build prompt
   - Call Gemini or Ollama
   - Get response

9. **Return Response:**
   - Format JSON
   - Send to client

---

## Performance & Optimization

### Indexing Performance

**First Request:**
- Chunking: ~100-200ms
- Embedding generation: ~2-5 seconds (10-20 chunks, concurrent)
- Storage: ~50-100ms
- **Total: ~5-10 seconds**

**Optimizations:**
- Concurrent embedding generation (5 parallel)
- Batch storage operations
- Skip if chunks exist

### Query Performance

**After Indexing:**
- Query embedding: ~200-500ms
- Vector search: ~10-50ms (PostgreSQL) or ~100-200ms (in-memory)
- Context building: ~10ms
- LLM call: ~1-3 seconds
- **Total: ~1.5-4 seconds**

**Optimizations:**
- Chunk reuse (no re-indexing)
- Efficient vector search (indexed)
- Similarity threshold filtering

### Scalability

**In-Memory Store:**
- Suitable for: < 10K chunks
- Performance: O(n) search
- Limitation: Memory usage

**PostgreSQL + pgvector:**
- Suitable for: Millions of chunks
- Performance: O(log n) with index
- Scalable: Yes

### Memory Usage

**Per Chunk:**
- Content: ~1-5 KB
- Embedding: 4096 × 4 bytes = 16 KB
- Metadata: ~100 bytes
- **Total: ~17-21 KB per chunk**

**Example:**
- 20 chunks = ~340-420 KB
- 1000 chunks = ~17-21 MB
- 10,000 chunks = ~170-210 MB

---

## Error Handling & Fallbacks

### Error Categories

1. **Initialization Errors:**
   - RAG manager creation fails
   - **Fallback:** Direct prompt (old method)

2. **Indexing Errors:**
   - Chunking fails
   - Embedding generation fails
   - Storage fails
   - **Fallback:** Direct prompt

3. **Retrieval Errors:**
   - Query embedding fails
   - Vector search fails
   - **Fallback:** Direct prompt

4. **LLM API Errors:**
   - Gemini API fails
   - Ollama API fails
   - **Fallback:** Try alternative API, then direct prompt

5. **Empty Results:**
   - No chunks retrieved
   - All chunks filtered
   - **Fallback:** Direct prompt

### Fallback Chain

```
RAG Attempt
    ↓ (fails)
Direct Prompt with Full Statement
    ↓ (fails)
Error Response to Client
```

### Error Logging

All errors are logged with:
- Error type
- Error message
- Stack trace (if available)
- Context information

**Example Logs:**
```
WARNING: Failed to index statement data: embedding generation failed
Falling back to direct prompt (without RAG)...
```

---

## Testing & Validation

### Unit Testing

**Test Chunking:**
```go
chunker := NewChunker(config)
chunks, err := chunker.ChunkStatementData(statementData, "test_id")
// Verify chunks created, types correct, content formatted
```

**Test Embeddings:**
```go
client := NewEmbeddingClient(config)
embedding, err := client.GenerateEmbedding("test text")
// Verify embedding dimensions, non-zero values
```

**Test Retrieval:**
```go
retriever := NewRetriever(client, store, config)
chunks, err := retriever.Retrieve("test query", "source_id")
// Verify relevant chunks returned, scores valid
```

### Integration Testing

**Test Full Pipeline:**
1. Index statement data
2. Query with various questions
3. Verify responses are accurate
4. Check context contains relevant chunks

**Test Fallbacks:**
1. Simulate Ollama failure
2. Verify fallback to direct prompt
3. Check error responses

### Manual Testing

**Test with curl:**
```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What was my total expense?",
    "statementData": { ... }
  }'
```

**Check Logs:**
- Indexing progress
- Retrieval results
- Context preview
- Response generation

---

## Troubleshooting

### Issue: Stuck at Indexing

**Symptoms:**
- Log shows "Indexing statement data..."
- No further logs
- Request times out

**Causes:**
1. Ollama not running
2. Model not available
3. Network timeout
4. Deadlock (fixed in latest version)

**Solutions:**
1. Check Ollama: `curl http://localhost:11434/api/tags`
2. Verify model: `ollama list`
3. Check logs for specific errors
4. Test embeddings directly: `curl -X POST http://localhost:11434/api/embeddings -d '{"model":"llama3","prompt":"test"}'`

### Issue: Empty Chunks Retrieved

**Symptoms:**
- Retrieval succeeds but no chunks
- Falls back to direct prompt

**Causes:**
1. Similarity threshold too high
2. Query embedding doesn't match chunk embeddings
3. No chunks indexed

**Solutions:**
1. Lower similarity threshold (config)
2. Check if chunks were indexed
3. Verify source ID matches
4. Check embedding generation logs

### Issue: Poor Answer Quality

**Symptoms:**
- Answers are incorrect
- Answers reference wrong data
- Answers say "I don't have that information"

**Causes:**
1. Wrong chunks retrieved
2. Context not clear
3. LLM not following instructions

**Solutions:**
1. Check retrieved chunks (logs)
2. Verify chunk content
3. Check similarity scores
4. Review system prompt
5. Lower similarity threshold

### Issue: Slow Performance

**Symptoms:**
- Requests take > 10 seconds
- Indexing takes too long

**Causes:**
1. Too many chunks
2. Ollama slow
3. No concurrency

**Solutions:**
1. Check chunk count (should be 10-30)
2. Optimize Ollama (more RAM/GPU)
3. Verify concurrent embedding generation
4. Use PostgreSQL for faster search

### Issue: PostgreSQL Not Working

**Symptoms:**
- Warning: "pgvector not available"
- Falls back to in-memory

**Solutions:**
1. Install pgvector: `CREATE EXTENSION vector;`
2. Check PostgreSQL connection
3. Verify DSN format
4. Check table creation logs

---

## Future Enhancements

### Planned Improvements

1. **Async Indexing:**
   - Index in background
   - Return immediately to user
   - Better UX

2. **Embedding Caching:**
   - Cache embeddings for identical chunks
   - Reduce API calls
   - Faster indexing

3. **HNSW Index:**
   - Replace IVFFlat with HNSW
   - Better accuracy
   - Faster search

4. **Hybrid Search:**
   - Combine semantic + keyword search
   - Better retrieval
   - More accurate

5. **Re-ranking:**
   - Re-rank retrieved chunks
   - Better relevance
   - Improved answers

6. **Streaming Responses:**
   - Server-Sent Events (SSE)
   - Real-time streaming
   - Better UX

7. **Session Management:**
   - Store conversation sessions
   - Better context
   - Multi-turn conversations

8. **Analytics:**
   - Track popular queries
   - Monitor performance
   - Optimize retrieval

---

## Code Structure

```
rag/
├── config.go              # Configuration management
├── types.go               # Shared types (Chunk)
├── chunker.go             # Semantic chunking
├── embeddings.go          # Embedding generation
├── retriever.go           # Similarity search
├── manager.go             # RAG pipeline orchestration
├── schema.sql             # PostgreSQL schema
└── vectorstore/
    ├── interface.go       # VectorStore interface
    ├── memory.go          # In-memory implementation
    └── pgvector.go        # PostgreSQL implementation

chat_handler.go            # HTTP handler for /api/chat
```

---

## Key Design Decisions

### 1. Why Semantic Chunking?

**Decision:** Chunk by transaction/entity, not by characters

**Reasoning:**
- Preserves context
- Better retrieval accuracy
- Complete information per chunk
- More meaningful for LLM

### 2. Why Two Vector Stores?

**Decision:** PostgreSQL (primary) + In-memory (fallback)

**Reasoning:**
- Production needs persistence
- Development needs simplicity
- Automatic fallback ensures reliability
- No single point of failure

### 3. Why Similarity Threshold?

**Decision:** Filter chunks below 0.3 similarity

**Reasoning:**
- Prevents irrelevant context
- Improves answer quality
- Reduces token usage
- Configurable for tuning

### 4. Why Concurrent Embeddings?

**Decision:** 5 parallel embedding requests

**Reasoning:**
- 3-5x faster than sequential
- Ollama can handle concurrent requests
- Semaphore prevents overwhelming
- Better resource utilization

### 5. Why Source ID Hashing?

**Decision:** SHA256 hash of statement data

**Reasoning:**
- Enables chunk reuse
- Avoids re-indexing
- Collision-resistant
- Deterministic

---

## Conclusion

This RAG implementation provides a robust, scalable solution for handling large bank statements in a chat interface. It ensures:

- ✅ **No truncation** - Context always fits
- ✅ **Relevant answers** - Only relevant chunks sent
- ✅ **Fast responses** - Optimized pipeline
- ✅ **Reliable** - Multiple fallbacks
- ✅ **Scalable** - Works with any statement size
- ✅ **Production-ready** - Error handling, logging, monitoring

The system is designed to be maintainable, testable, and extensible for future enhancements.

---

## Appendix

### A. Chunk Types Reference

| Type | Description | Example ID |
|------|-------------|------------|
| `account_summary` | Account overview | `stmt_xxx_summary_0` |
| `transaction_breakdown` | Payment method breakdown | `stmt_xxx_breakdown_1` |
| `transactions` | Transaction batches | `stmt_xxx_transactions_2` |
| `top_expenses` | Top expenses list | `stmt_xxx_top_expenses_8` |
| `monthly_summary` | Monthly summaries | `stmt_xxx_monthly_9` |
| `category_summary` | Category-wise totals | `stmt_xxx_categories_10` |
| `top_beneficiaries` | Top beneficiaries | `stmt_xxx_beneficiaries_11` |

### B. API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/chat` | POST | Chat with statement data |
| `/api/health` | GET | Health check |
| `/classify` | POST | Statement classification (existing) |

### C. Environment Variables Summary

```bash
# Ollama Configuration
OLLAMA_URL=http://localhost:11434
OLLAMA_CHAT_MODEL=llama3
OLLAMA_EMBEDDING_MODEL=llama3

# PostgreSQL (Optional)
POSTGRES_DSN=postgres://user:password@localhost/dbname?sslmode=disable

# Gemini (Optional)
GEMINI_API_KEY=your-key
```

### D. Log Messages Reference

**Indexing:**
- `"Chunking statement data for source: ..."`
- `"Created X chunks"`
- `"Generating embeddings for X chunks"`
- `"Generating embedding X/Y..."`
- `"Successfully generated X/Y embeddings"`
- `"Storing X valid chunks..."`
- `"Successfully indexed statement data: X chunks stored"`

**Retrieval:**
- `"Retrieving relevant chunks for query: ..."`
- `"Retrieved X relevant chunks with scores: [...]"`
- `"Filtered out chunk ... with low similarity: X.XXX"`

**Context:**
- `"Context built: X chunks, total length: X characters"`
- `"Context preview: ..."`

**LLM:**
- `"Calling Ollama API with RAG context..."`
- `"Calling Gemini API with RAG context..."`
- `"Response sent successfully to client"`

---

**End of Document**

