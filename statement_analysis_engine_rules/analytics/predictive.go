package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"classify/statement_analysis_engine_rules/utils"
	"time"
)

// CalculatePredictiveInsights calculates predictive insights
func CalculatePredictiveInsights(
	transactions []models.ClassifiedTransaction,
	closingBalance float64,
) models.PredictiveInsights {
	// Calculate average daily expense
	avgDailyExpense := calculateAverageDailyExpense(transactions)
	projected30DaySpend := avgDailyExpense * 30

	// Predict low balance date
	predictedLowBalanceDate := predictLowBalanceDate(closingBalance, avgDailyExpense)

	// Calculate upcoming EMI impact
	upcomingEMI := calculateUpcomingEMI(transactions)

	// Generate savings recommendation
	savingsRecommendation := generateSavingsRecommendation(closingBalance, avgDailyExpense)

	return models.PredictiveInsights{
		Projected30DaySpend:     projected30DaySpend,
		PredictedLowBalanceDate: predictedLowBalanceDate,
		UpcomingEMIImpact:       upcomingEMI,
		SavingsRecommendation:   savingsRecommendation,
	}
}

func calculateAverageDailyExpense(transactions []models.ClassifiedTransaction) float64 {
	totalExpense := 0.0
	days := 0

	firstDate := ""
	lastDate := ""

	for _, txn := range transactions {
		// Only count withdrawals (expenses), not deposits
		if txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
			totalExpense += txn.WithdrawalAmt
			if firstDate == "" {
				firstDate = txn.Date
			}
			lastDate = txn.Date
		}
	}

	if firstDate == "" || lastDate == "" {
		return 0
	}

	first, _ := utils.ParseDate(firstDate)
	last, _ := utils.ParseDate(lastDate)

	if first.IsZero() || last.IsZero() {
		days = 30 // Default
	} else {
		days = int(last.Sub(first).Hours() / 24)
		if days == 0 {
			days = 1
		}
	}

	if days > 0 {
		return totalExpense / float64(days)
	}
	return 0
}

func predictLowBalanceDate(currentBalance float64, avgDailyExpense float64) string {
	if avgDailyExpense == 0 {
		return "N/A"
	}

	daysUntilLow := int(currentBalance / avgDailyExpense)
	if daysUntilLow < 0 {
		daysUntilLow = 0
	}

	futureDate := time.Now().AddDate(0, 0, daysUntilLow)
	return utils.FormatDate(futureDate, "DD/MM/YYYY")
}

func calculateUpcomingEMI(transactions []models.ClassifiedTransaction) float64 {
	// Find recurring EMI payments
	emiAmount := 0.0
	for _, txn := range transactions {
		// EMI is always a withdrawal (expense), not a deposit
		if txn.Method == "EMI" && txn.IsRecurring && txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
			emiAmount = txn.WithdrawalAmt
			break
		}
	}
	return emiAmount
}

func generateSavingsRecommendation(balance float64, avgDailyExpense float64) string {
	if balance > 100000 {
		return "Move ₹50k to FD to earn 7% interest"
	} else if balance > 50000 {
		return "Move ₹25k to FD to earn 7% interest"
	} else if balance > 20000 {
		return "Consider starting a recurring deposit"
	}
	return "Build emergency fund of 3-6 months expenses"
}

// CalculateTransactionTrends calculates transaction trends
func CalculateTransactionTrends(monthlySummary []models.MonthlySummary, categorySummary models.CategorySummary) models.TransactionTrends {
	highestSpendMonth := ""
	maxExpense := 0.0

	for _, month := range monthlySummary {
		if month.Expense > maxExpense {
			maxExpense = month.Expense
			highestSpendMonth = month.Month
		}
	}

	largestCategory := "Other"
	maxCategoryAmount := 0.0

	categories := map[string]float64{
		"Food_Delivery":   categorySummary.FoodDelivery,
		"Dining":          categorySummary.Dining,
		"Travel":          categorySummary.Travel,
		"Shopping":        categorySummary.Shopping,
		"Groceries":       categorySummary.Groceries,
		"Bills_Utilities": categorySummary.BillsUtilities,
	}

	for category, amount := range categories {
		if amount > maxCategoryAmount {
			maxCategoryAmount = amount
			largestCategory = category
		}
	}

	return models.TransactionTrends{
		HighestSpendMonth: highestSpendMonth,
		LargestCategory:   largestCategory,
	}
}
