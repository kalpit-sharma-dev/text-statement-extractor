package analytics

import "classify/statement_analysis_engine_rules/models"

// CalculateCategorySummary calculates category-wise summary
// NOTE: This is for OPERATIONAL EXPENSES only - investments are tracked separately
func CalculateCategorySummary(transactions []models.ClassifiedTransaction) models.CategorySummary {
	summary := models.CategorySummary{}

	// Investment categories/methods to exclude
	investmentCategories := map[string]bool{
		"Investment":    true,
		"Investments":   true,
		"Self_Transfer": true,
	}
	investmentMethods := map[string]bool{
		"RD":         true,
		"FD":         true,
		"SIP":        true,
		"Investment": true,
	}

	for _, txn := range transactions {
		// Only count operational expenses (withdrawals), skip deposits (income)
		if txn.DepositAmt > 0 || txn.WithdrawalAmt == 0 {
			continue
		}

		// Skip investments - they're tracked separately, not as expenses
		isInvestment := investmentCategories[txn.Category] || investmentMethods[txn.Method]
		if isInvestment {
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
		case "Loan", "Loan_EMI", "LOAN_EMI":
			// Loan EMI expenses (keep this as it's an operational expense)
			summary.Loan += amount
		}
		// Note: "Investment" case removed - investments are NOT expenses
		// They're tracked separately in accountSummary.totalInvestments
	}

	return summary
}
