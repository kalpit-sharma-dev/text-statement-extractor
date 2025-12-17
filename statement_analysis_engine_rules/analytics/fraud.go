package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"math"
)

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

		amount := txn.WithdrawalAmt

		// Flag large transactions
		if amount > 50000 {
			alerts = append(alerts, models.FraudAlert{
				Amount:   amount,
				Merchant: txn.Merchant,
			})
			if amount > 100000 {
				riskLevel = "Medium"
			}
		}

		// Flag unknown merchants with large amounts
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
