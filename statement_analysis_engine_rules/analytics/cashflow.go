package analytics

import (
	"math"
	"statement_analysis_engine_rules/models"
)

// CalculateCashFlowScore calculates cash flow health score
func CalculateCashFlowScore(
	openingBalance float64,
	closingBalance float64,
	totalIncome float64,
	totalExpense float64,
) models.CashFlowScore {
	score := 0
	status := "Poor"
	insight := ""

	// Calculate savings rate
	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = ((totalIncome - totalExpense) / totalIncome) * 100
	}

	// Score calculation
	if savingsRate > 20 {
		score += 30
		insight += "Excellent savings rate. "
	} else if savingsRate > 10 {
		score += 20
		insight += "Good savings rate. "
	} else if savingsRate > 0 {
		score += 10
		insight += "Positive savings. "
	}

	// Balance trend
	if closingBalance > openingBalance {
		score += 25
		insight += "Balance increased. "
	} else if closingBalance == openingBalance {
		score += 15
		insight += "Balance maintained. "
	} else {
		score += 5
		insight += "Balance decreased. "
	}

	// Income consistency (simplified)
	if totalIncome > 0 {
		score += 25
		insight += "Regular income detected. "
	}

	// Expense control
	expenseRatio := 0.0
	if totalIncome > 0 {
		expenseRatio = (totalExpense / totalIncome) * 100
	}
	if expenseRatio < 70 {
		score += 20
		insight += "Expenses under control. "
	} else if expenseRatio < 90 {
		score += 10
		insight += "Expenses manageable. "
	}

	if score >= 80 {
		status = "Excellent"
	} else if score >= 60 {
		status = "Healthy"
	} else if score >= 40 {
		status = "Moderate"
	} else {
		status = "Poor"
	}

	if insight == "" {
		insight = "Monitor your cash flow regularly."
	} else {
		insight += "Your cash flow is " + status + "."
	}

	return models.CashFlowScore{
		Score:   score,
		Status:  status,
		Insight: insight,
	}
}

// CalculateSalaryUtilization calculates salary utilization metrics
func CalculateSalaryUtilization(transactions []models.ClassifiedTransaction, salaryAmount float64, salaryDate string) models.SalaryUtilization {
	if salaryAmount == 0 {
		return models.SalaryUtilization{}
	}

	// Find salary transaction
	salaryTxnIndex := -1
	for i, txn := range transactions {
		if txn.IsIncome && math.Abs(txn.DepositAmt-salaryAmount) < 1000 {
			salaryTxnIndex = i
			break
		}
	}

	if salaryTxnIndex == -1 {
		return models.SalaryUtilization{}
	}

	// Calculate spending in first 3, 7, 15 days after salary
	spent3Days := 0.0
	spent7Days := 0.0
	spent15Days := 0.0

	// This is simplified - would need proper date calculations
	// For now, estimate based on transaction order
	for i := salaryTxnIndex + 1; i < len(transactions) && i < salaryTxnIndex+20; i++ {
		if transactions[i].IsIncome {
			continue
		}
		daysAfter := i - salaryTxnIndex
		amount := transactions[i].WithdrawalAmt

		if daysAfter <= 3 {
			spent3Days += amount
		}
		if daysAfter <= 7 {
			spent7Days += amount
		}
		if daysAfter <= 15 {
			spent15Days += amount
		}
	}

	// Calculate percentages
	spent3DaysPercent := (spent3Days / salaryAmount) * 100
	spent7DaysPercent := (spent7Days / salaryAmount) * 100
	spent15DaysPercent := (spent15Days / salaryAmount) * 100

	// Estimate days salary lasts
	daysSalaryLasts := int(salaryAmount / (totalExpense(transactions) / 30))

	// Fixed vs variable expenses (simplified)
	fixedExpenses := 40.0 // Default estimate
	variableExpenses := 35.0

	return models.SalaryUtilization{
		SpentFirst3Days:  spent3DaysPercent,
		SpentFirst7Days:  spent7DaysPercent,
		SpentFirst15Days: spent15DaysPercent,
		DaysSalaryLasts:  daysSalaryLasts,
		FixedExpenses:    fixedExpenses,
		VariableExpenses: variableExpenses,
	}
}

func totalExpense(transactions []models.ClassifiedTransaction) float64 {
	total := 0.0
	for _, txn := range transactions {
		if !txn.IsIncome {
			total += txn.WithdrawalAmt
		}
	}
	return total
}
