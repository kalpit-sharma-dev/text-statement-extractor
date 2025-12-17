package analytics

import "classify/statement_analysis_engine_rules/models"

// CalculateTransactionBreakdown calculates breakdown by transaction method
func CalculateTransactionBreakdown(transactions []models.ClassifiedTransaction) models.TransactionBreakdown {
	breakdown := models.TransactionBreakdown{}

	for _, txn := range transactions {
		// Determine amount based on transaction type
		// For deposits (income), use DepositAmt
		// For withdrawals (expenses), use WithdrawalAmt
		var amount float64
		if txn.DepositAmt > 0 && txn.WithdrawalAmt == 0 {
			amount = txn.DepositAmt
		} else if txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
			amount = txn.WithdrawalAmt
		} else if txn.DepositAmt > 0 && txn.WithdrawalAmt > 0 {
			// Both amounts present - use the larger one
			if txn.DepositAmt > txn.WithdrawalAmt {
				amount = txn.DepositAmt
			} else {
				amount = txn.WithdrawalAmt
			}
		} else {
			// No amount, skip
			continue
		}

		switch txn.Method {
		case "UPI":
			breakdown.UPI.Amount += amount
			breakdown.UPI.Count++
		case "IMPS":
			breakdown.IMPS.Amount += amount
			breakdown.IMPS.Count++
		case "NEFT":
			breakdown.NEFT.Amount += amount
			breakdown.NEFT.Count++
		case "EMI":
			breakdown.EMI.Amount += amount
			breakdown.EMI.Count++
		case "RD", "FD", "SIP":
			// Investment transactions - don't count in EMI
			// These are tracked separately or can be added to a new field if needed
			// For now, they won't be counted in transaction breakdown
			// (They're investments, not payment methods)
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
