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
		case "RTGS":
			breakdown.RTGS.Amount += amount
			breakdown.RTGS.Count++
		case "EMI":
			breakdown.EMI.Amount += amount
			breakdown.EMI.Count++
		case "ACH", "BillPaid":
			breakdown.BillPaid.Amount += amount
			breakdown.BillPaid.Count++
		case "DebitCard":
			breakdown.DebitCard.Amount += amount
			breakdown.DebitCard.Count++
		case "ATMWithdrawal":
			// ATM Withdrawal - count separately (not in Other)
			breakdown.ATMWithdrawal.Amount += amount
			breakdown.ATMWithdrawal.Count++
		case "NetBanking":
			breakdown.NetBanking.Amount += amount
			breakdown.NetBanking.Count++
		case "Salary":
			// Salary - count separately (not in Other)
			breakdown.Salary.Amount += amount
			breakdown.Salary.Count++
		case "RD":
			// Recurring Deposit - count separately AND in Investment total
			breakdown.RD.Amount += amount
			breakdown.RD.Count++
			// Also add to Investment for consolidated view
			breakdown.Investment.Amount += amount
			breakdown.Investment.Count++
		case "FD":
			// Fixed Deposit - count separately AND in Investment total
			breakdown.FD.Amount += amount
			breakdown.FD.Count++
			// Also add to Investment for consolidated view
			breakdown.Investment.Amount += amount
			breakdown.Investment.Count++
		case "SIP":
			// Systematic Investment Plan - count separately AND in Investment total
			breakdown.SIP.Amount += amount
			breakdown.SIP.Count++
			// Also add to Investment for consolidated view
			breakdown.Investment.Amount += amount
			breakdown.Investment.Count++
		case "Interest":
			// Interest - count separately (not in Other)
			breakdown.Interest.Amount += amount
			breakdown.Interest.Count++
		case "Cheque":
			// Cheque - count separately (not in Other)
			breakdown.Cheque.Amount += amount
			breakdown.Cheque.Count++
		case "Dividend":
			// Dividend - count separately (not in Other)
			breakdown.Dividend.Amount += amount
			breakdown.Dividend.Count++
		case "Investment":
			// Investment transactions (like Indian Clearing Corporation)
			// Count separately (not in Other)
			breakdown.Investment.Amount += amount
			breakdown.Investment.Count++
		case "Insurance":
			// Insurance premium - count in BillPaid (it's a bill payment)
			breakdown.BillPaid.Amount += amount
			breakdown.BillPaid.Count++
		case "Self_Transfer":
			// Self-transfer (INF/INFT) - internal fund transfer
			// These are transfers to own investment/savings accounts, so count as Investment
			breakdown.Investment.Amount += amount
			breakdown.Investment.Count++
		case "OnlineShopping":
			// Online shopping (ONL) - count as Other or based on underlying method
			// The actual payment method (UPI, Card, etc.) is already counted elsewhere
			breakdown.Other.Amount += amount
			breakdown.Other.Count++
		case "TaxPayment":
			// Tax payment (DTAX/IDTX) - count as BillPaid
			breakdown.BillPaid.Amount += amount
			breakdown.BillPaid.Count++
		case "Other":
			// Explicitly classified as Other
			breakdown.Other.Amount += amount
			breakdown.Other.Count++
		default:
			// Catch all unmatched methods (empty string, unknown methods, etc.)
			breakdown.Other.Amount += amount
			breakdown.Other.Count++
		}
	}

	return breakdown
}
