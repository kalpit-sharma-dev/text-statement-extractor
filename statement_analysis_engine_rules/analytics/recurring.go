package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"classify/statement_analysis_engine_rules/rules"
	"classify/statement_analysis_engine_rules/utils"
)

// CalculateRecurringPayments identifies recurring payments
func CalculateRecurringPayments(transactions []models.ClassifiedTransaction) []models.RecurringPayment {
	recurringMap := make(map[string]*recurringData)

	for _, txn := range transactions {
		// Only process withdrawals (expenses) for recurring payments
		// Recurring deposits (like salary) are income, not expenses
		if txn.WithdrawalAmt == 0 || txn.DepositAmt > 0 {
			continue
		}
		
		// Check if transaction is recurring
		// Include if IsRecurring flag is set OR if rules detect it as recurring
		// Use withdrawal amount only since we're checking for recurring expenses
		isRecurring := txn.IsRecurring || rules.IsRecurringPayment(txn.Narration, txn.WithdrawalAmt, txn.Date, []string{})
		if !isRecurring {
			continue
		}

		// Extract recurring details
		name, pattern := rules.ExtractRecurringDetails(txn.Narration)
		if name == "" {
			name = txn.Merchant
			if name == "" {
				name = txn.Beneficiary
			}
		}

		key := name + "|" + pattern
		if recurringMap[key] == nil {
			recurringMap[key] = &recurringData{
				name:       name,
				pattern:    pattern,
				amount:     txn.WithdrawalAmt,
				dayOfMonth: extractDayOfMonth(txn.Date),
			}
		} else {
			// Accumulate total amount and count for proper averaging
			// For now, use the latest amount (could be enhanced to track count and average)
			// This is a simplification - ideally we'd track count and calculate true average
			recurringMap[key].amount = txn.WithdrawalAmt // Use latest amount
		}
	}

	result := make([]models.RecurringPayment, 0, len(recurringMap))
	for _, data := range recurringMap {
		result = append(result, models.RecurringPayment{
			Name:       data.name,
			Amount:     data.amount,
			DayOfMonth: data.dayOfMonth,
			Pattern:    data.pattern,
		})
	}

	return result
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
