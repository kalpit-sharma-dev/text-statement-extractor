package analytics

import (
	"sort"
	"statement_analysis_engine_rules/models"
)

// CalculateTopExpenses calculates top expenses
func CalculateTopExpenses(transactions []models.ClassifiedTransaction, limit int) []models.TopExpense {
	expenses := make([]models.TopExpense, 0)

	for _, txn := range transactions {
		if txn.IsIncome || txn.WithdrawalAmt == 0 {
			continue
		}

		merchant := txn.Merchant
		if merchant == "" {
			merchant = txn.Beneficiary
		}
		if merchant == "" {
			merchant = "Unknown"
		}

		expenses = append(expenses, models.TopExpense{
			Merchant: merchant,
			Date:     txn.Date,
			Amount:   txn.WithdrawalAmt,
			Category: txn.Category,
		})
	}

	// Sort by amount descending
	sort.Slice(expenses, func(i, j int) bool {
		return expenses[i].Amount > expenses[j].Amount
	})

	// Take top N
	if limit > len(expenses) {
		limit = len(expenses)
	}

	return expenses[:limit]
}
