package analytics

import "classify/statement_analysis_engine_rules/models"

// CalculateCategorySummary calculates category-wise summary
func CalculateCategorySummary(transactions []models.ClassifiedTransaction) models.CategorySummary {
	summary := models.CategorySummary{}

	for _, txn := range transactions {
		// Only count expenses (withdrawals), skip deposits (income)
		if txn.DepositAmt > 0 || txn.WithdrawalAmt == 0 {
			continue
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
