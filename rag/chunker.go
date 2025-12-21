package rag

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Chunker handles semantic chunking of documents
type Chunker struct {
	config *Config
}

// NewChunker creates a new chunker with the given configuration
func NewChunker(config *Config) *Chunker {
	return &Chunker{
		config: config,
	}
}

// ChunkStatementData chunks bank statement data into semantic units
// This is critical: we chunk by transaction/entity, not by raw characters
func (c *Chunker) ChunkStatementData(statementData interface{}, sourceID string) ([]*Chunk, error) {
	// Convert statement data to JSON for processing
	jsonData, err := json.Marshal(statementData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal statement data: %w", err)
	}
	
	// Parse into a map to extract structured components
	var dataMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &dataMap); err != nil {
		return nil, fmt.Errorf("failed to parse statement data: %w", err)
	}
	
	var chunks []*Chunk
	chunkIndex := 0
	
	// Chunk 1: Account Summary (single chunk)
	if accountSummary, ok := dataMap["accountSummary"].(map[string]interface{}); ok {
		summaryText := c.formatAccountSummary(accountSummary)
		chunks = append(chunks, &Chunk{
			ID:       fmt.Sprintf("%s_summary_%d", sourceID, chunkIndex),
			SourceID: sourceID,
			Content:  summaryText,
			Metadata: map[string]interface{}{
				"type": "account_summary",
			},
		})
		chunkIndex++
	}
	
	// Chunk 2: Transaction Breakdown (single chunk)
	if txBreakdown, ok := dataMap["transactionBreakdown"].(map[string]interface{}); ok {
		breakdownText := c.formatTransactionBreakdown(txBreakdown)
		chunks = append(chunks, &Chunk{
			ID:       fmt.Sprintf("%s_breakdown_%d", sourceID, chunkIndex),
			SourceID: sourceID,
			Content:  breakdownText,
			Metadata: map[string]interface{}{
				"type": "transaction_breakdown",
			},
		})
		chunkIndex++
	}
	
	// Chunk 3-N: Individual Transactions (chunked in groups to stay within token limits)
	if transactions, ok := dataMap["transactions"].([]interface{}); ok {
		txChunks := c.chunkTransactions(transactions, sourceID, chunkIndex)
		chunks = append(chunks, txChunks...)
		chunkIndex += len(txChunks)
	}
	
	// Chunk: Top Expenses
	if topExpenses, ok := dataMap["topExpenses"].([]interface{}); ok {
		expensesText := c.formatTopExpenses(topExpenses)
		chunks = append(chunks, &Chunk{
			ID:       fmt.Sprintf("%s_top_expenses_%d", sourceID, chunkIndex),
			SourceID: sourceID,
			Content:  expensesText,
			Metadata: map[string]interface{}{
				"type": "top_expenses",
			},
		})
		chunkIndex++
	}
	
	// Chunk: Monthly Summary
	if monthlySummary, ok := dataMap["monthlySummary"].([]interface{}); ok {
		monthlyText := c.formatMonthlySummary(monthlySummary)
		chunks = append(chunks, &Chunk{
			ID:       fmt.Sprintf("%s_monthly_%d", sourceID, chunkIndex),
			SourceID: sourceID,
			Content:  monthlyText,
			Metadata: map[string]interface{}{
				"type": "monthly_summary",
			},
		})
		chunkIndex++
	}
	
	// Chunk: Category Summary
	if categorySummary, ok := dataMap["categorySummary"].(map[string]interface{}); ok {
		categoryText := c.formatCategorySummary(categorySummary)
		chunks = append(chunks, &Chunk{
			ID:       fmt.Sprintf("%s_categories_%d", sourceID, chunkIndex),
			SourceID: sourceID,
			Content:  categoryText,
			Metadata: map[string]interface{}{
				"type": "category_summary",
			},
		})
		chunkIndex++
	}
	
	// Chunk: Top Beneficiaries
	if topBeneficiaries, ok := dataMap["topBeneficiaries"].([]interface{}); ok {
		beneficiariesText := c.formatTopBeneficiaries(topBeneficiaries)
		chunks = append(chunks, &Chunk{
			ID:       fmt.Sprintf("%s_beneficiaries_%d", sourceID, chunkIndex),
			SourceID: sourceID,
			Content:  beneficiariesText,
			Metadata: map[string]interface{}{
				"type": "top_beneficiaries",
			},
		})
		chunkIndex++
	}
	
	return chunks, nil
}

// chunkTransactions chunks transactions in groups to respect token limits
func (c *Chunker) chunkTransactions(transactions []interface{}, sourceID string, startIndex int) []*Chunk {
	var chunks []*Chunk
	var currentChunk strings.Builder
	currentCount := 0
	chunkIndex := startIndex
	
	// Estimate: each transaction ~50-100 tokens
	// Target: ~400 tokens per chunk = ~4-8 transactions
	transactionsPerChunk := 6
	
	for i, tx := range transactions {
		txText := c.formatTransaction(tx)
		
		// If adding this transaction would exceed limit, start new chunk
		if currentCount >= transactionsPerChunk && currentChunk.Len() > 0 {
			chunks = append(chunks, &Chunk{
				ID:       fmt.Sprintf("%s_transactions_%d", sourceID, chunkIndex),
				SourceID: sourceID,
				Content:  currentChunk.String(),
				Metadata: map[string]interface{}{
					"type": "transactions",
					"count": currentCount,
				},
			})
			chunkIndex++
			currentChunk.Reset()
			currentCount = 0
		}
		
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(txText)
		currentCount++
		
		// Handle last chunk
		if i == len(transactions)-1 && currentChunk.Len() > 0 {
			chunks = append(chunks, &Chunk{
				ID:       fmt.Sprintf("%s_transactions_%d", sourceID, chunkIndex),
				SourceID: sourceID,
				Content:  currentChunk.String(),
				Metadata: map[string]interface{}{
					"type": "transactions",
					"count": currentCount,
				},
			})
		}
	}
	
	return chunks
}

// Formatting helpers
func (c *Chunker) formatAccountSummary(summary map[string]interface{}) string {
	var parts []string
	parts = append(parts, "Account Summary:")
	
	if val, ok := summary["accountNumberMasked"]; ok {
		parts = append(parts, fmt.Sprintf("Account: %v", val))
	}
	if val, ok := summary["customerName"]; ok {
		parts = append(parts, fmt.Sprintf("Customer: %v", val))
	}
	if val, ok := summary["statementPeriod"]; ok {
		parts = append(parts, fmt.Sprintf("Period: %v", val))
	}
	if val, ok := summary["openingBalance"]; ok {
		parts = append(parts, fmt.Sprintf("Opening Balance: ₹%.2f", toFloat64(val)))
	}
	if val, ok := summary["closingBalance"]; ok {
		parts = append(parts, fmt.Sprintf("Closing Balance: ₹%.2f", toFloat64(val)))
	}
	if val, ok := summary["totalIncome"]; ok {
		parts = append(parts, fmt.Sprintf("Total Income: ₹%.2f", toFloat64(val)))
	}
	if val, ok := summary["totalExpense"]; ok {
		parts = append(parts, fmt.Sprintf("Total Expense: ₹%.2f", toFloat64(val)))
	}
	if val, ok := summary["totalInvestments"]; ok {
		parts = append(parts, fmt.Sprintf("Total Investments: ₹%.2f", toFloat64(val)))
	}
	if val, ok := summary["netSavings"]; ok {
		parts = append(parts, fmt.Sprintf("Net Savings: ₹%.2f", toFloat64(val)))
	}
	if val, ok := summary["savingsRatePercent"]; ok {
		parts = append(parts, fmt.Sprintf("Savings Rate: %.2f%%", toFloat64(val)))
	}
	
	return strings.Join(parts, "\n")
}

func (c *Chunker) formatTransactionBreakdown(breakdown map[string]interface{}) string {
	var parts []string
	parts = append(parts, "Transaction Breakdown by Payment Method:")
	
	for method, data := range breakdown {
		if methodMap, ok := data.(map[string]interface{}); ok {
			amount := toFloat64(methodMap["amount"])
			count := toInt(methodMap["count"])
			parts = append(parts, fmt.Sprintf("%s: ₹%.2f (%d transactions)", method, amount, count))
		}
	}
	
	return strings.Join(parts, "\n")
}

func (c *Chunker) formatTransaction(tx interface{}) string {
	if txMap, ok := tx.(map[string]interface{}); ok {
		var parts []string
		if date, ok := txMap["date"].(string); ok {
			parts = append(parts, fmt.Sprintf("Date: %s", date))
		}
		if amount, ok := txMap["amount"]; ok {
			parts = append(parts, fmt.Sprintf("Amount: ₹%.2f", toFloat64(amount)))
		}
		if txType, ok := txMap["type"].(string); ok {
			parts = append(parts, fmt.Sprintf("Type: %s", txType))
		}
		if category, ok := txMap["category"].(string); ok {
			parts = append(parts, fmt.Sprintf("Category: %s", category))
		}
		if merchant, ok := txMap["merchant"].(string); ok {
			parts = append(parts, fmt.Sprintf("Merchant: %s", merchant))
		}
		if method, ok := txMap["paymentMethod"].(string); ok {
			parts = append(parts, fmt.Sprintf("Method: %s", method))
		}
		return strings.Join(parts, " | ")
	}
	return fmt.Sprintf("%v", tx)
}

func (c *Chunker) formatTopExpenses(expenses []interface{}) string {
	var parts []string
	parts = append(parts, "Top Expenses:")
	
	for i, exp := range expenses {
		if i >= 10 { // Limit to top 10
			break
		}
		if expMap, ok := exp.(map[string]interface{}); ok {
			merchant := toString(expMap["merchant"])
			amount := toFloat64(expMap["amount"])
			category := toString(expMap["category"])
			date := toString(expMap["date"])
			parts = append(parts, fmt.Sprintf("%d. %s - ₹%.2f (%s) on %s", i+1, merchant, amount, category, date))
		}
	}
	
	return strings.Join(parts, "\n")
}

func (c *Chunker) formatMonthlySummary(monthly []interface{}) string {
	var parts []string
	parts = append(parts, "Monthly Summary:")
	
	for _, month := range monthly {
		if monthMap, ok := month.(map[string]interface{}); ok {
			monthName := toString(monthMap["month"])
			income := toFloat64(monthMap["income"])
			expense := toFloat64(monthMap["expense"])
			closing := toFloat64(monthMap["closingBalance"])
			topCategory := toString(monthMap["topCategory"])
			parts = append(parts, fmt.Sprintf("%s: Income ₹%.2f, Expense ₹%.2f, Balance ₹%.2f, Top Category: %s", 
				monthName, income, expense, closing, topCategory))
		}
	}
	
	return strings.Join(parts, "\n")
}

func (c *Chunker) formatCategorySummary(categories map[string]interface{}) string {
	var parts []string
	parts = append(parts, "Category Summary:")
	
	for category, amount := range categories {
		parts = append(parts, fmt.Sprintf("%s: ₹%.2f", category, toFloat64(amount)))
	}
	
	return strings.Join(parts, "\n")
}

func (c *Chunker) formatTopBeneficiaries(beneficiaries []interface{}) string {
	var parts []string
	parts = append(parts, "Top Beneficiaries:")
	
	for i, ben := range beneficiaries {
		if i >= 10 {
			break
		}
		if benMap, ok := ben.(map[string]interface{}); ok {
			name := toString(benMap["name"])
			amount := toFloat64(benMap["amount"])
			benType := toString(benMap["type"])
			parts = append(parts, fmt.Sprintf("%d. %s - ₹%.2f (%s)", i+1, name, amount, benType))
		}
	}
	
	return strings.Join(parts, "\n")
}

// Helper functions for type conversion
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	default:
		return 0
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

