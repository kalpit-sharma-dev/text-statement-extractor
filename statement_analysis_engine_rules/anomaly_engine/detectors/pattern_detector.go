package detectors

import (
	"fmt"
	"math"
	"strings"
	"time"

	"classify/statement_analysis_engine_rules/anomaly_engine/profiles"
	"classify/statement_analysis_engine_rules/anomaly_engine/types"
	"classify/statement_analysis_engine_rules/models"
)

// PatternDetector detects multi-transaction patterns (composite anomalies)
// Detects patterns like: multiple large transfers to same account, high-value recurring payments
type PatternDetector struct {
	BaseDetector
	config *PatternConfig
	history []models.ClassifiedTransaction
}

// PatternConfig holds pattern detection configuration
type PatternConfig struct {
	// Multiple transfers to same account
	SameAccountThreshold    float64 // Minimum amount per transfer
	SameAccountTimeWindow   float64 // Days to look back
	SameAccountMinCount     int     // Minimum number of transfers to flag
	SameAccountTotalAmount  float64 // Total amount threshold
	
	// High-value recurring payments
	HighValueRecurringThreshold float64 // Amount threshold for high-value recurring
	
	// Large bill payments
	LargeBillThreshold float64 // Threshold for unusually large bills
}

// DefaultPatternConfig returns default pattern detection config
func DefaultPatternConfig() *PatternConfig {
	return &PatternConfig{
		SameAccountThreshold:    30000,  // ₹30K per transfer
		SameAccountTimeWindow:   7.0,    // 7 days
		SameAccountMinCount:     2,      // At least 2 transfers
		SameAccountTotalAmount:  100000, // Total ₹1L+
		HighValueRecurringThreshold: 10000, // ₹10K+ recurring
		LargeBillThreshold:      50000,  // ₹50K+ bills
	}
}

// NewPatternDetector creates a new pattern detector
func NewPatternDetector(config *PatternConfig, history []models.ClassifiedTransaction) *PatternDetector {
	if config == nil {
		config = DefaultPatternConfig()
	}
	return &PatternDetector{
		BaseDetector: BaseDetector{name: "PatternDetector"},
		config:       config,
		history:      history,
	}
}

// Detect implements Detector interface
func (p *PatternDetector) Detect(ctx types.TransactionContext, profile *profiles.UserProfile) []types.AnomalySignal {
	signals := make([]types.AnomalySignal, 0)
	
	txn := ctx.Txn
	amount := txn.WithdrawalAmt
	
	if amount == 0 || len(p.history) == 0 {
		return signals
	}
	
	// Pattern 1: Multiple large transfers to same account
	signals = append(signals, p.detectMultipleTransfersToSameAccount(txn, ctx)...)
	
	// Pattern 2: High-value recurring payment
	signals = append(signals, p.detectHighValueRecurring(txn, profile)...)
	
	// Pattern 3: Unusually large bill payment
	signals = append(signals, p.detectLargeBillPayment(txn)...)
	
	return signals
}

// detectMultipleTransfersToSameAccount detects multiple large transfers to same beneficiary
func (p *PatternDetector) detectMultipleTransfersToSameAccount(txn models.ClassifiedTransaction, ctx types.TransactionContext) []types.AnomalySignal {
	signals := make([]types.AnomalySignal, 0)
	
	amount := txn.WithdrawalAmt
	beneficiary := strings.ToUpper(strings.TrimSpace(txn.Beneficiary))
	merchant := strings.ToUpper(strings.TrimSpace(txn.Merchant))
	
	// Use beneficiary if available, otherwise merchant
	targetAccount := beneficiary
	if targetAccount == "" {
		targetAccount = merchant
	}
	
	if targetAccount == "" || amount < p.config.SameAccountThreshold {
		return signals
	}
	
	// Look through recent history
	txnDate, err := parseTransactionDateForPattern(txn.Date)
	if err != nil {
		return signals
	}
	
	var matchingTransfers []models.ClassifiedTransaction
	var totalAmount float64
	
	for _, other := range p.history {
		if other.WithdrawalAmt == 0 {
			continue
		}
		
		otherBeneficiary := strings.ToUpper(strings.TrimSpace(other.Beneficiary))
		otherMerchant := strings.ToUpper(strings.TrimSpace(other.Merchant))
		otherTarget := otherBeneficiary
		if otherTarget == "" {
			otherTarget = otherMerchant
		}
		
		// Check if same account
		if otherTarget == "" || otherTarget != targetAccount {
			continue
		}
		
		// Check amount threshold
		if other.WithdrawalAmt < p.config.SameAccountThreshold {
			continue
		}
		
		// Check time window
		otherDate, err := parseTransactionDateForPattern(other.Date)
		if err != nil {
			continue
		}
		
		daysDiff := math.Abs(txnDate.Sub(otherDate).Hours() / 24)
		if daysDiff > p.config.SameAccountTimeWindow {
			continue
		}
		
		matchingTransfers = append(matchingTransfers, other)
		totalAmount += other.WithdrawalAmt
	}
	
	// Include current transaction
	totalAmount += amount
	count := len(matchingTransfers) + 1
	
	// Check if pattern detected
	if count >= p.config.SameAccountMinCount && totalAmount >= p.config.SameAccountTotalAmount {
		score := 75.0
		if count >= 3 {
			score = 90.0 // Higher score for 3+ transfers (critical pattern)
		} else if totalAmount >= 200000 {
			score = 85.0 // Higher score for ₹2L+ total
		}
		
		explanation := fmt.Sprintf("Multiple large transfers (₹%.0f total, %d transfers) to %s within %.0f days. This pattern may indicate rapid fund movement. Please verify if these transfers are authorized.", 
			totalAmount, count, maskAccount(targetAccount), p.config.SameAccountTimeWindow)
		
		signals = append(signals, types.NewSignal(
			types.SignalMultipleLargeTransfers,
			types.CategoryPattern,
			score,
			explanation,
		))
	}
	
	return signals
}

// detectHighValueRecurring detects high-value recurring payments
func (p *PatternDetector) detectHighValueRecurring(txn models.ClassifiedTransaction, profile *profiles.UserProfile) []types.AnomalySignal {
	signals := make([]types.AnomalySignal, 0)
	
	amount := txn.WithdrawalAmt
	if amount < p.config.HighValueRecurringThreshold {
		return signals
	}
	
	// Suppress for known legitimate recurring payments (EMIs, loans, credit card bills)
	// These are expected to be recurring and shouldn't be flagged as anomalies
	narrationUpper := strings.ToUpper(txn.Narration)
	category := txn.Category
	method := txn.Method
	
	// Check if it's a known recurring payment type
	isLegitimateRecurring := false
	
	// EMI/Loan payments are inherently recurring
	if method == "EMI" || category == "Loan" || 
		strings.Contains(narrationUpper, "EMI") || 
		strings.Contains(narrationUpper, "LOAN EMI") ||
		strings.Contains(narrationUpper, "STAFF LOAN EMI") {
		isLegitimateRecurring = true
	}
	
	// Credit card bill payments (via CRED, ACH D to banks, etc.)
	if category == "Bills_Utilities" && 
		(strings.Contains(narrationUpper, "CRED") || 
		 strings.Contains(narrationUpper, "ACH D") ||
		 strings.Contains(narrationUpper, "CREDIT CARD") ||
		 strings.Contains(narrationUpper, "CARD BILL")) {
		isLegitimateRecurring = true
	}
	
	// Rent payments (typically to same person/entity monthly)
	if category == "Bills_Utilities" && 
		(strings.Contains(narrationUpper, "RENT") || 
		 strings.Contains(narrationUpper, "HOUSE RENT")) {
		isLegitimateRecurring = true
	}
	
	// Insurance premiums
	if category == "Bills_Utilities" && method == "Insurance" {
		isLegitimateRecurring = true
	}
	
	// If it's a legitimate recurring payment, don't flag it
	if isLegitimateRecurring {
		return signals
	}
	
	merchant := strings.ToUpper(strings.TrimSpace(txn.Merchant))
	beneficiary := strings.ToUpper(strings.TrimSpace(txn.Beneficiary))
	target := merchant
	if beneficiary != "" {
		target = beneficiary
	}
	
	// Check if this appears to be recurring (same merchant/beneficiary, similar amount)
	similarCount := 0
	var similarAmounts []float64
	
	for _, other := range p.history {
		if other.WithdrawalAmt == 0 {
			continue
		}
		
		otherMerchant := strings.ToUpper(strings.TrimSpace(other.Merchant))
		otherBeneficiary := strings.ToUpper(strings.TrimSpace(other.Beneficiary))
		otherTarget := otherMerchant
		if otherBeneficiary != "" {
			otherTarget = otherBeneficiary
		}
		
		if otherTarget != target {
			continue
		}
		
		// Check if amount is similar (within 10%)
		amountDiff := math.Abs(amount - other.WithdrawalAmt)
		amountRatio := amountDiff / amount
		if amountRatio <= 0.1 {
			similarCount++
			similarAmounts = append(similarAmounts, other.WithdrawalAmt)
		}
	}
	
	// If 2+ similar transactions, it's likely recurring
	if similarCount >= 2 {
		score := 60.0
		if amount >= 50000 {
			score = 70.0 // Higher score for ₹50K+
		}
		
		annualCost := amount * 12
		explanation := fmt.Sprintf("High-value recurring payment of ₹%.0f detected. This appears to be a regular monthly payment. Annual commitment: ₹%.0f. Please verify this matches your intent.", 
			amount, annualCost)
		
		signals = append(signals, types.NewSignal(
			types.SignalHighValueRecurring,
			types.CategoryPattern,
			score,
			explanation,
		))
	}
	
	return signals
}

// detectLargeBillPayment detects unusually large bill payments
func (p *PatternDetector) detectLargeBillPayment(txn models.ClassifiedTransaction) []types.AnomalySignal {
	signals := make([]types.AnomalySignal, 0)
	
	amount := txn.WithdrawalAmt
	category := txn.Category
	
	// Check if it's a bill/utility payment
	isBillPayment := category == "Bills_Utilities" || 
		strings.Contains(strings.ToUpper(txn.Merchant), "CRED") ||
		strings.Contains(strings.ToUpper(txn.Merchant), "BILL") ||
		strings.Contains(strings.ToUpper(txn.Merchant), "UTILITY")
	
	if !isBillPayment || amount < p.config.LargeBillThreshold {
		return signals
	}
	
	score := 65.0
	if amount >= 100000 {
		score = 75.0 // Higher score for ₹1L+
	}
	
	explanation := fmt.Sprintf("Large bill payment of ₹%.0f detected. This is unusually high for a bill payment. CRED is typically used for credit card bills. Please verify what this payment covers and if the amount is expected.", amount)
	
	signals = append(signals, types.NewSignal(
		types.SignalLargeBillPayment,
		types.CategoryPattern,
		score,
		explanation,
	))
	
	return signals
}

// Helper functions

func parseTransactionDateForPattern(dateStr string) (time.Time, error) {
	layouts := []string{
		"02/01/2006",
		"2006-01-02",
		"02-01-2006",
		"01/02/2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Now(), nil
}

func maskAccount(account string) string {
	if len(account) <= 4 {
		return account
	}
	// Show last 4 characters, mask the rest
	return "XXXXXX" + account[len(account)-4:]
}

