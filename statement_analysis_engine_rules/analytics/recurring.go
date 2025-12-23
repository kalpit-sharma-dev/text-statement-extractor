package analytics

import (
	"classify/statement_analysis_engine_rules/models"
)

// CalculateRecurringPayments identifies recurring payments using comprehensive detection
// This now uses the new comprehensive recurring payment detection system
func CalculateRecurringPayments(transactions []models.ClassifiedTransaction) []models.RecurringPayment {
	detector := NewRecurringPaymentDetector(transactions)
	return detector.DetectRecurringPayments()
}
