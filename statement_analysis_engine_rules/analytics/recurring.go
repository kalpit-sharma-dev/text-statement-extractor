package analytics

import (
	"statement_analysis_engine_rules/models"
	"statement_analysis_engine_rules/rules"
	"statement_analysis_engine_rules/utils"
)

// CalculateRecurringPayments identifies recurring payments
func CalculateRecurringPayments(transactions []models.ClassifiedTransaction) []models.RecurringPayment {
	recurringMap := make(map[string]*recurringData)

	for _, txn := range transactions {
		if !txn.IsRecurring && !rules.IsRecurringPayment(txn.Narration, txn.WithdrawalAmt, txn.Date, []string{}) {
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
			// Average amount if multiple occurrences
			recurringMap[key].amount = (recurringMap[key].amount + txn.WithdrawalAmt) / 2
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
