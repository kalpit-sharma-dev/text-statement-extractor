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
		case "Shopping":
			summary.Shopping += amount
		case "Bills_Utilities":
			summary.BillsUtilities += amount
		case "Travel":
			summary.Travel += amount
		case "Dining":
			summary.Dining += amount
		case "Groceries":
			summary.Groceries += amount
		case "Food_Delivery":
			summary.FoodDelivery += amount
		case "Fuel":
			summary.Fuel += amount
		case "Investment":
			// Investment expenses (withdrawals for investments)
			summary.Investments += amount
		}
	}

	return summary
}
