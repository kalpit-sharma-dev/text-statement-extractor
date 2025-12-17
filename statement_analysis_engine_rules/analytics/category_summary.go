package analytics

import "statement_analysis_engine_rules/models"

// CalculateCategorySummary calculates category-wise summary
func CalculateCategorySummary(transactions []models.ClassifiedTransaction) models.CategorySummary {
	summary := models.CategorySummary{}

	for _, txn := range transactions {
		if txn.IsIncome {
			continue // Only count expenses
		}

		amount := txn.WithdrawalAmt

		switch txn.Category {
		case "Food_Delivery":
			summary.FoodDelivery += amount
		case "Dining":
			summary.Dining += amount
		case "Travel":
			summary.Travel += amount
		case "Shopping":
			summary.Shopping += amount
		case "Groceries":
			summary.Groceries += amount
		case "Bills_Utilities":
			summary.BillsUtilities += amount
		}
	}

	return summary
}
