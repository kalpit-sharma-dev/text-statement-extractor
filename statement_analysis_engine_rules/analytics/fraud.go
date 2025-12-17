package analytics

import (
	"math"
	"statement_analysis_engine_rules/models"
)

// CalculateFraudRisk calculates fraud risk indicators
func CalculateFraudRisk(transactions []models.ClassifiedTransaction) models.FraudRisk {
	riskLevel := "Low"
	alerts := make([]models.FraudAlert, 0)

	// Check for unusual transactions
	for _, txn := range transactions {
		if txn.IsIncome {
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
		amount := txn.WithdrawalAmt
		if txn.IsIncome {
			amount = txn.DepositAmt
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

			txnType := "Debit"
			if txn.IsIncome {
				txnType = "Credit"
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
