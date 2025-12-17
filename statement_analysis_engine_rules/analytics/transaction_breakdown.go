package analytics

import "statement_analysis_engine_rules/models"

// CalculateTransactionBreakdown calculates breakdown by transaction method
func CalculateTransactionBreakdown(transactions []models.ClassifiedTransaction) models.TransactionBreakdown {
	breakdown := models.TransactionBreakdown{}

	for _, txn := range transactions {
		amount := txn.WithdrawalAmt
		if txn.IsIncome {
			amount = txn.DepositAmt
		}

		switch txn.Method {
		case "UPI":
			breakdown.UPI.Amount += amount
			breakdown.UPI.Count++
		case "IMPS":
			breakdown.IMPS.Amount += amount
			breakdown.IMPS.Count++
		case "EMI":
			breakdown.EMI.Amount += amount
			breakdown.EMI.Count++
		case "ACH", "BillPaid":
			breakdown.BillPaid.Amount += amount
			breakdown.BillPaid.Count++
		case "DebitCard":
			breakdown.DebitCard.Amount += amount
			breakdown.DebitCard.Count++
		case "NetBanking":
			breakdown.NetBanking.Amount += amount
			breakdown.NetBanking.Count++
		}
	}

	return breakdown
}
