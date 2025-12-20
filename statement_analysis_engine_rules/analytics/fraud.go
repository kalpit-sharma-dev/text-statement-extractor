package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"math"
	"strings"
)

// WhitelistedMerchants are known legitimate merchants that should NOT trigger fraud alerts
// These are trusted investment platforms, banks, and self-transfers
var WhitelistedMerchants = []string{
	// Investment platforms
	"ZERODHA", "ZERODHA BROKING", "GROWW", "UPSTOX", "COIN", "KITE",
	"ANGEL BROKING", "ICICI SECURITIES", "HDFC SECURITIES", "KOTAK SECURITIES",
	"SHAREKHAN", "MOTILAL OSWAL", "IIFL", "5PAISA",
	// Clearing corporations
	"INDIAN CLEARING CORPORATION", "NSDL", "CDSL",
	// Banks (for self-transfers)
	"HDFC BANK", "ICICI BANK", "SBI", "AXIS BANK", "KOTAK", "IDFC",
	// Insurance
	"LIC", "HDFC LIFE", "ICICI PRUDENTIAL", "SBI LIFE", "MAXLIFE", "BAJAJ ALLIANZ",
	// Crypto exchanges
	"WAZIRX", "COINDCX", "COINSWITCH", "ZEBPAY",
	// Mutual funds
	"SIP", "MUTUAL FUND", "RD", "FD",
	// Government
	"INCOMETAX", "GST", "PAYGOV",
}

// WhitelistedCategories are categories that should NOT trigger fraud alerts
var WhitelistedCategories = []string{
	"Investment", "Self_Transfer", "Income", "Bills_Utilities", "Loan",
}

// isWhitelisted checks if a transaction is from a known legitimate source
func isWhitelisted(txn models.ClassifiedTransaction) bool {
	// Check if category is whitelisted
	for _, cat := range WhitelistedCategories {
		if txn.Category == cat {
			return true
		}
	}
	
	// Check if method indicates self-transfer or investment
	if txn.Method == "Self_Transfer" || txn.Method == "Investment" || 
	   txn.Method == "RD" || txn.Method == "FD" || txn.Method == "SIP" {
		return true
	}
	
	// Check if merchant is whitelisted
	merchantUpper := strings.ToUpper(txn.Merchant)
	narrationUpper := strings.ToUpper(txn.Narration)
	
	for _, whitelisted := range WhitelistedMerchants {
		if strings.Contains(merchantUpper, whitelisted) || strings.Contains(narrationUpper, whitelisted) {
			return true
		}
	}
	
	return false
}

// CalculateFraudRisk calculates fraud risk indicators
func CalculateFraudRisk(transactions []models.ClassifiedTransaction) models.FraudRisk {
	riskLevel := "Low"
	alerts := make([]models.FraudAlert, 0)

	// Check for unusual transactions
	for _, txn := range transactions {
		// Only check withdrawals (expenses) for fraud risk
		// Skip deposits (income) and transactions with no withdrawal amount
		if txn.DepositAmt > 0 || txn.WithdrawalAmt == 0 {
			continue
		}
		
		// Skip whitelisted transactions (known investments, self-transfers, etc.)
		if isWhitelisted(txn) {
			continue
		}

		amount := txn.WithdrawalAmt

		// Flag large transactions (only if NOT whitelisted)
		if amount > 50000 {
			alerts = append(alerts, models.FraudAlert{
				Amount:   amount,
				Merchant: txn.Merchant,
			})
			if amount > 100000 {
				riskLevel = "Medium"
			}
		}

		// Flag unknown merchants with large amounts (only if NOT whitelisted)
		if (txn.Merchant == "" || txn.Merchant == "Unknown") && amount > 10000 {
			alerts = append(alerts, models.FraudAlert{
				Amount:   amount,
				Merchant: "Unknown Vendor",
			})
		}
	}

	// Limit alerts to recent ones
	if len(alerts) > 5 {
		alerts = alerts[:5]
	}

	if len(alerts) > 3 {
		riskLevel = "High"
	}

	return models.FraudRisk{
		RiskLevel:    riskLevel,
		RecentAlerts: alerts,
	}
}

// CalculateBigTicketMovements identifies big ticket movements
func CalculateBigTicketMovements(transactions []models.ClassifiedTransaction, threshold float64) []models.BigTicketMovement {
	movements := make([]models.BigTicketMovement, 0)

	for _, txn := range transactions {
		// Skip regular income transactions - they are expected income, not "big ticket movements"
		// Salary, Interest, and Dividend are regular income streams that don't need to be flagged
		if txn.Method == "Salary" || txn.Method == "Interest" || txn.Method == "Dividend" {
			continue
		}
		
		// Determine amount and type - can be either deposit or withdrawal
		var amount float64
		var txnType string
		
		if txn.DepositAmt > 0 && txn.WithdrawalAmt == 0 {
			// Deposit (credit)
			amount = txn.DepositAmt
			txnType = "Credit"
		} else if txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
			// Withdrawal (debit)
			amount = txn.WithdrawalAmt
			txnType = "Debit"
		} else if txn.DepositAmt > 0 && txn.WithdrawalAmt > 0 {
			// Both present - use the larger one
			if txn.DepositAmt > txn.WithdrawalAmt {
				amount = txn.DepositAmt
				txnType = "Credit"
			} else {
				amount = txn.WithdrawalAmt
				txnType = "Debit"
			}
		} else {
			// No amount, skip
			continue
		}

		if math.Abs(amount) >= threshold {
			impact := "Low Impact"
			if amount >= threshold*2 {
				impact = "High Impact"
			} else if amount >= threshold*1.5 {
				impact = "Medium Impact"
			}

			description := txn.Merchant
			if description == "" {
				description = txn.Beneficiary
			}
			if description == "" {
				description = txn.Narration
				if len(description) > 50 {
					description = description[:50] + "..."
				}
			}

			movements = append(movements, models.BigTicketMovement{
				Description: description,
				Amount:      amount,
				Date:        txn.Date,
				Type:        txnType,
				Category:    txn.Category,
				Impact:      impact,
			})
		}
	}

	return movements
}
