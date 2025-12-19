package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"classify/statement_analysis_engine_rules/rules"
	"classify/statement_analysis_engine_rules/utils"
	"math"
	"strings"
)

// CalculateRecurringPayments identifies recurring payments using pattern detection
func CalculateRecurringPayments(transactions []models.ClassifiedTransaction) []models.RecurringPayment {
	// Step 1: Group transactions by beneficiary/merchant
	beneficiaryTransactions := make(map[string][]transactionData)

	for _, txn := range transactions {
		// Only process withdrawals (expenses) for recurring payments
		if txn.WithdrawalAmt == 0 || txn.DepositAmt > 0 {
			continue
		}

		// Get beneficiary/merchant name
		name := txn.Beneficiary
		if name == "" {
			name = txn.Merchant
		}
		if name == "" || name == "Unknown" {
			continue
		}

		// Normalize name for grouping
		nameKey := strings.ToUpper(strings.TrimSpace(name))

		beneficiaryTransactions[nameKey] = append(beneficiaryTransactions[nameKey], transactionData{
			date:   txn.Date,
			amount: txn.WithdrawalAmt,
			narration: txn.Narration,
		})
	}

	// Step 2: Analyze each beneficiary for recurring patterns
	result := make([]models.RecurringPayment, 0)

	for name, txns := range beneficiaryTransactions {
		// Need at least 2 transactions to detect a pattern
		if len(txns) < 2 {
			continue
		}

		// Check if it's a recurring pattern
		isRecurring, avgAmount, dayOfMonth := analyzeRecurringPattern(txns)
		
		// Also check if narration explicitly indicates recurring payment
		hasRecurringKeyword := false
		for _, txn := range txns {
			if rules.IsRecurringPayment(txn.narration, txn.amount, txn.date, []string{}) {
				hasRecurringKeyword = true
				break
			}
		}

		if isRecurring || hasRecurringKeyword {
			result = append(result, models.RecurringPayment{
				Name:       name,
				Amount:     avgAmount,
				DayOfMonth: dayOfMonth,
				Pattern:    "Monthly",
			})
		}
	}

	return result
}

type transactionData struct {
	date      string
	amount    float64
	narration string
}

// analyzeRecurringPattern checks if transactions show a recurring pattern
func analyzeRecurringPattern(txns []transactionData) (bool, float64, int) {
	if len(txns) < 2 {
		return false, 0, 0
	}

	// Parse dates and extract days of month
	daysOfMonth := make([]int, 0)
	amounts := make([]float64, 0)

	for _, txn := range txns {
		t, err := utils.ParseDate(txn.date)
		if err != nil {
			continue
		}
		daysOfMonth = append(daysOfMonth, t.Day())
		amounts = append(amounts, txn.amount)
	}

	if len(daysOfMonth) < 2 {
		return false, 0, 0
	}

	// Calculate average amount
	avgAmount := 0.0
	for _, amt := range amounts {
		avgAmount += amt
	}
	avgAmount /= float64(len(amounts))

	// Check if amounts are similar (within 20% variance for flexibility)
	// Some recurring payments may vary slightly (e.g., utility bills)
	amountVariance := 0.0
	for _, amt := range amounts {
		diff := math.Abs(amt - avgAmount)
		variance := diff / avgAmount
		if variance > amountVariance {
			amountVariance = variance
		}
	}

	// Calculate average day of month
	avgDay := 0
	for _, day := range daysOfMonth {
		avgDay += day
	}
	avgDay /= len(daysOfMonth)

	// Check if days are consistent (within ±5 days for flexibility)
	// Some recurring payments may shift by a few days
	dayVariance := 0
	for _, day := range daysOfMonth {
		diff := int(math.Abs(float64(day - avgDay)))
		if diff > dayVariance {
			dayVariance = diff
		}
	}

	// Criteria for recurring payment:
	// 1. Amounts are similar (variance < 30% for flexibility with varying bills)
	// 2. Days of month are consistent (within ±7 days)
	// 3. OR: Amounts are very similar (variance < 5%) regardless of day consistency
	//    (for cases like EMIs that might be debited on different days but same amount)
	
	isAmountConsistent := amountVariance < 0.30 // 30% variance allowed
	isAmountVerySimilar := amountVariance < 0.05 // 5% variance (very similar)
	isDayConsistent := dayVariance <= 7 // Within 7 days

	isRecurring := (isAmountConsistent && isDayConsistent) || isAmountVerySimilar

	return isRecurring, avgAmount, avgDay
}

type recurringData struct {
	name       string
	pattern    string
	amount     float64
	dayOfMonth int
}

func extractDayOfMonth(dateStr string) int {
	t, err := utils.ParseDate(dateStr)
	if err != nil {
		return 0
	}
	return t.Day()
}
