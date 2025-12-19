package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"strings"
)

// PrepareTransactionsForResponse converts ClassifiedTransaction to TransactionDetail for API response
// This includes all transactions for heatmap, spending patterns, and detailed analysis
func PrepareTransactionsForResponse(transactions []models.ClassifiedTransaction) []models.TransactionDetail {
	result := make([]models.TransactionDetail, 0, len(transactions))

	for _, txn := range transactions {
		// Determine transaction type and amount
		var txnType string
		var amount float64

		if txn.DepositAmt > 0 && txn.WithdrawalAmt == 0 {
			txnType = "Credit"
			amount = txn.DepositAmt
		} else if txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
			txnType = "Debit"
			amount = txn.WithdrawalAmt
		} else if txn.DepositAmt > 0 && txn.WithdrawalAmt > 0 {
			// Both present - use the larger one
			if txn.DepositAmt > txn.WithdrawalAmt {
				txnType = "Credit"
				amount = txn.DepositAmt
			} else {
				txnType = "Debit"
				amount = txn.WithdrawalAmt
			}
		} else {
			// No amount, skip
			continue
		}

		// Extract time from narration if available (banks usually don't provide separate time)
		time := extractTimeFromNarration(txn.Narration)

		// Extract reference number from narration
		refNumber := extractReferenceNumber(txn.Narration, txn.ChequeRefNo)

		// Use merchant name, or beneficiary if merchant is empty
		merchant := txn.Merchant
		if merchant == "" || merchant == "Unknown" {
			merchant = txn.Beneficiary
		}
		if merchant == "" {
			// Try to extract a meaningful name from narration
			merchant = extractMerchantFromNarration(txn.Narration)
		}

		detail := models.TransactionDetail{
			// Required fields
			Date:          txn.Date,
			Amount:        amount,
			Type:          txnType,
			Category:      txn.Category,
			Merchant:      merchant,
			PaymentMethod: txn.Method,

			// Optional fields
			Time:            time,
			Description:     txn.Narration,
			Balance:         txn.ClosingBalance,
			ReferenceNumber: refNumber,
			Beneficiary:     txn.Beneficiary,
			IsRecurring:     txn.IsRecurring,
		}

		result = append(result, detail)
	}

	return result
}

// extractTimeFromNarration tries to extract time from narration
// Most banks don't provide time, but some include it in narration
func extractTimeFromNarration(narration string) string {
	// Look for patterns like "14:30:00" or "14:30" in narration
	// This is a best-effort extraction
	narration = strings.ToUpper(narration)
	
	// Common patterns:
	// - "TIME: 14:30:00"
	// - "AT 14:30"
	// - "14:30:00 HRS"
	
	// For now, return empty as most bank statements don't include time
	// Can be enhanced if specific time patterns are found in narrations
	return ""
}

// extractReferenceNumber extracts reference number from narration or uses ChequeRefNo
func extractReferenceNumber(narration string, chequeRefNo string) string {
	// If ChequeRefNo is provided, use it
	if chequeRefNo != "" && chequeRefNo != "0" {
		return chequeRefNo
	}

	// Try to extract reference number from narration
	// Common patterns:
	// - "REF:123456789"
	// - "IMPS-123456789-"
	// - "UPI/123456789/"
	
	narration = strings.ToUpper(narration)
	
	// Look for UPI reference
	if strings.Contains(narration, "UPI/") || strings.Contains(narration, "UPI-") {
		parts := strings.Split(narration, "/")
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	
	// Look for IMPS/NEFT/RTGS reference
	if strings.Contains(narration, "IMPS-") || strings.Contains(narration, "NEFT-") || strings.Contains(narration, "RTGS-") {
		parts := strings.Split(narration, "-")
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	
	// Look for REF: or REF NO:
	if strings.Contains(narration, "REF") {
		parts := strings.Fields(narration)
		for i, part := range parts {
			if strings.Contains(part, "REF") && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}
	
	// Return empty if no pattern matches
	return ""
}

// extractMerchantFromNarration extracts a meaningful merchant name from narration as fallback
func extractMerchantFromNarration(narration string) string {
	// If narration is short, use it as is
	if len(narration) <= 30 {
		return narration
	}
	
	// Take first 30 characters and add ellipsis
	return narration[:30] + "..."
}

