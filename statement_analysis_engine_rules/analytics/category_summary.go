package analytics

import "classify/statement_analysis_engine_rules/models"

// CalculateCategorySummary calculates category-wise summary
// NOTE: This is for OPERATIONAL EXPENSES only - investments and income are tracked separately
func CalculateCategorySummary(transactions []models.ClassifiedTransaction) models.CategorySummary {
	summary := models.CategorySummary{}

	// Categories to EXCLUDE from expense summary
	// These are NOT operational expenses - they're tracked separately
	excludedCategories := map[string]bool{
		"Investment":    true,  // Tracked in totalInvestments
		"Investments":   true,  // Tracked in totalInvestments
		"Self_Transfer": true,  // Internal transfers, not expenses
		"Income":        true,  // Income, not expense
		"Refund":        true,  // Refunds reduce spend, not add to it
		"Reimbursement": true,  // Reimbursements are credits, not expenses
	}
	
	// Methods that indicate investments/savings (not expenses)
	investmentMethods := map[string]bool{
		"RD":            true,  // Recurring Deposit
		"FD":            true,  // Fixed Deposit
		"SIP":           true,  // Systematic Investment Plan
		"Investment":    true,  // Generic investment
		"Self_Transfer": true,  // Self-transfer
		"Salary":        true,  // Income, not expense
		"Interest":      true,  // Income, not expense
		"Dividend":      true,  // Income, not expense
	}

	for _, txn := range transactions {
		// Only count operational expenses (withdrawals), skip deposits (income)
		if txn.DepositAmt > 0 || txn.WithdrawalAmt == 0 {
			continue
		}

		// Skip excluded categories - they're tracked separately
		if excludedCategories[txn.Category] || investmentMethods[txn.Method] {
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
		case "Healthcare":
			summary.Healthcare += amount
		case "Education":
			summary.Education += amount
		case "Entertainment":
			summary.Entertainment += amount
		}
		// Note: "Investment", "Income", "Refund" cases excluded - they're NOT expenses
		// Investments tracked in accountSummary.totalInvestments
		// Income tracked in accountSummary.totalIncome
		// Refunds net off against spending
	}

	return summary
}
