package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"sort"
)

// CalculateTopExpenses calculates top expenses
func CalculateTopExpenses(transactions []models.ClassifiedTransaction, limit int) []models.TopExpense {
	expenses := make([]models.TopExpense, 0)

	for _, txn := range transactions {
		// Only count expenses (withdrawals), skip deposits (income)
		if txn.DepositAmt > 0 || txn.WithdrawalAmt == 0 {
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
	if limit <= 0 {
		return []models.TopExpense{} // Return empty if invalid limit
	}
	if limit > len(expenses) {
		limit = len(expenses)
	}

	return expenses[:limit]
}
