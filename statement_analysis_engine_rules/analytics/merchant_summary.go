package analytics

import "classify/statement_analysis_engine_rules/models"

// CalculateMerchantSummary calculates merchant-wise summary
func CalculateMerchantSummary(transactions []models.ClassifiedTransaction) models.MerchantSummary {
	summary := models.MerchantSummary{}

	for _, txn := range transactions {
		// Only count expenses (withdrawals), skip deposits (income)
		if txn.DepositAmt > 0 || txn.WithdrawalAmt == 0 {
			continue
		}

		amount := txn.WithdrawalAmt
		merchant := txn.Merchant

		switch merchant {
		case "Amazon":
			summary.Amazon += amount
		case "Flipkart":
			summary.Flipkart += amount
		case "Swiggy":
			summary.Swiggy += amount
		case "Zomato":
			summary.Zomato += amount
		case "Uber":
			summary.Uber += amount
		}
	}

	return summary
}
