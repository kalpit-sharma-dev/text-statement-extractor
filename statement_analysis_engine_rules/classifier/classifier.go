package classifier

import (
	"statement_analysis_engine_rules/models"
	"statement_analysis_engine_rules/rules"
	"statement_analysis_engine_rules/utils"
)

// ClassifyTransaction classifies a single transaction
func ClassifyTransaction(txn models.ClassifiedTransaction) models.ClassifiedTransaction {
	// Normalize narration
	normalizedNarration := utils.NormalizeNarration(txn.Narration)

	// Classify method
	txn.Method = rules.ClassifyMethod(normalizedNarration)

	// Extract merchant
	txn.Merchant = rules.ExtractMerchantName(normalizedNarration)
	if txn.Merchant == "Unknown" {
		txn.Merchant = ""
	}

	// Classify category
	txn.Category = rules.ClassifyCategory(normalizedNarration, txn.Merchant)

	// Extract beneficiary
	txn.Beneficiary = rules.ExtractBeneficiary(normalizedNarration, txn.Method)

	// Determine if income or expense
	txn.IsIncome = txn.DepositAmt > 0

	// Check if bill payment
	if rules.IsBillPayment(normalizedNarration) {
		if txn.Category == "Other" {
			txn.Category = "Bills_Utilities"
		}
	}

	return txn
}

// ClassifyTransactions classifies a list of transactions
func ClassifyTransactions(transactions []models.ClassifiedTransaction) []models.ClassifiedTransaction {
	classified := make([]models.ClassifiedTransaction, len(transactions))
	previousNarrations := make([]string, 0)

	for i, txn := range transactions {
		// Classify transaction
		classified[i] = ClassifyTransaction(txn)

		// Check for recurring payments
		previousNarrations = append(previousNarrations, txn.Narration)
		if len(previousNarrations) > 10 {
			previousNarrations = previousNarrations[1:] // Keep last 10
		}

		classified[i].IsRecurring = rules.IsRecurringPayment(
			txn.Narration,
			txn.WithdrawalAmt+txn.DepositAmt,
			txn.Date,
			previousNarrations,
		)
	}

	return classified
}

// ConvertFromTxtTransaction converts from extracted statement transaction to classified transaction
func ConvertFromTxtTransaction(date, narration, chequeRefNo, valueDate string, withdrawalAmt, depositAmt, closingBalance float64) models.ClassifiedTransaction {
	return models.ClassifiedTransaction{
		Date:           date,
		Narration:      narration,
		ChequeRefNo:    chequeRefNo,
		ValueDate:      valueDate,
		WithdrawalAmt:  withdrawalAmt,
		DepositAmt:     depositAmt,
		ClosingBalance: closingBalance,
		Method:         "",
		Category:       "",
		Merchant:       "",
		Beneficiary:    "",
		IsIncome:       depositAmt > 0,
		IsRecurring:    false,
	}
}
