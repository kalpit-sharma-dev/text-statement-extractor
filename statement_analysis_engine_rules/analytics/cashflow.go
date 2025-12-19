package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"classify/statement_analysis_engine_rules/utils"
	"strings"
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
// Automatically detects salary transactions and calculates spending patterns
func CalculateSalaryUtilization(transactions []models.ClassifiedTransaction, salaryAmount float64, salaryDate string) models.SalaryUtilization {
	// Auto-detect salary transactions if parameters not provided
	salaryTransactions := findSalaryTransactions(transactions)
	
	if len(salaryTransactions) == 0 {
		return models.SalaryUtilization{}
	}

	// Use the most recent salary transaction
	latestSalary := salaryTransactions[len(salaryTransactions)-1]
	avgSalaryAmount := calculateAverageSalary(salaryTransactions)

	// Calculate spending in first 3, 7, 15 days after latest salary
	spent3Days := 0.0
	spent7Days := 0.0
	spent15Days := 0.0

	salaryDate = latestSalary.Date
	salaryTime, err := utils.ParseDate(salaryDate)
	if err != nil {
		return models.SalaryUtilization{}
	}

	// Iterate through all transactions after salary date
	for _, txn := range transactions {
		// Only count actual expenses (exclude investments and self-transfers)
		if txn.DepositAmt > 0 || txn.WithdrawalAmt == 0 {
			continue
		}

		// Skip investments and self-transfers
		if txn.Category == "Investment" || txn.Category == "Investments" || 
		   txn.Category == "Self_Transfer" {
			continue
		}
		if txn.Method == "RD" || txn.Method == "FD" || txn.Method == "SIP" || 
		   txn.Method == "Investment" {
			continue
		}

		txnTime, err := utils.ParseDate(txn.Date)
		if err != nil {
			continue
		}

		// Only count transactions after salary date
		if txnTime.Before(salaryTime) {
			continue
		}

		// Calculate days after salary
		daysAfter := int(txnTime.Sub(salaryTime).Hours() / 24)
		if daysAfter < 0 {
			continue
		}

		amount := txn.WithdrawalAmt

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

	// Calculate percentages (avoid division by zero)
	spent3DaysPercent := 0.0
	spent7DaysPercent := 0.0
	spent15DaysPercent := 0.0
	if avgSalaryAmount > 0 {
		spent3DaysPercent = (spent3Days / avgSalaryAmount) * 100
		spent7DaysPercent = (spent7Days / avgSalaryAmount) * 100
		spent15DaysPercent = (spent15Days / avgSalaryAmount) * 100
	}

	// Calculate average daily operational expense (exclude investments)
	operationalExpense := calculateOperationalExpense(transactions)
	totalDays := calculateStatementDays(transactions)
	avgDailyExpense := 0.0
	if totalDays > 0 {
		avgDailyExpense = operationalExpense / float64(totalDays)
	}

	// Estimate days salary lasts
	daysSalaryLasts := 0
	if avgDailyExpense > 0 {
		daysSalaryLasts = int(avgSalaryAmount / avgDailyExpense)
	}

	// Calculate fixed vs variable expenses based on recurring patterns
	fixedExpenses, variableExpenses := calculateFixedVsVariable(transactions)

	return models.SalaryUtilization{
		SpentFirst3Days:  spent3DaysPercent,
		SpentFirst7Days:  spent7DaysPercent,
		SpentFirst15Days: spent15DaysPercent,
		DaysSalaryLasts:  daysSalaryLasts,
		FixedExpenses:    fixedExpenses,
		VariableExpenses: variableExpenses,
	}
}

// findSalaryTransactions finds all salary transactions
func findSalaryTransactions(transactions []models.ClassifiedTransaction) []models.ClassifiedTransaction {
	salaries := make([]models.ClassifiedTransaction, 0)
	
	for _, txn := range transactions {
		// Salary is always a deposit (income), never a withdrawal
		if txn.DepositAmt == 0 || txn.WithdrawalAmt > 0 {
			continue
		}

		// Check if it's marked as salary method or has large regular deposit
		if txn.Method == "Salary" || txn.Category == "Salary" {
			salaries = append(salaries, txn)
			continue
		}

		// Look for salary patterns in narration
		// Large regular deposits (> ₹50,000) might be salary
		if txn.DepositAmt >= 50000 {
			// Check for salary keywords in narration
			narrationUpper := strings.ToUpper(txn.Narration)
			salaryKeywords := []string{"SALARY", "SAL", "PAYROLL", "WAGES", "SALARY CREDIT"}
			for _, keyword := range salaryKeywords {
				if strings.Contains(narrationUpper, keyword) {
					salaries = append(salaries, txn)
					break
				}
			}
		}
	}

	return salaries
}

// calculateAverageSalary calculates average salary from salary transactions
func calculateAverageSalary(salaries []models.ClassifiedTransaction) float64 {
	if len(salaries) == 0 {
		return 0
	}

	total := 0.0
	for _, sal := range salaries {
		total += sal.DepositAmt
	}

	return total / float64(len(salaries))
}

// calculateOperationalExpense calculates total operational expenses (excluding investments)
func calculateOperationalExpense(transactions []models.ClassifiedTransaction) float64 {
	total := 0.0
	
	for _, txn := range transactions {
		// Only count withdrawals (expenses), not deposits
		if txn.WithdrawalAmt == 0 || txn.DepositAmt > 0 {
			continue
		}

		// Skip investments and self-transfers
		if txn.Category == "Investment" || txn.Category == "Investments" || 
		   txn.Category == "Self_Transfer" {
			continue
		}
		if txn.Method == "RD" || txn.Method == "FD" || txn.Method == "SIP" || 
		   txn.Method == "Investment" {
			continue
		}

		total += txn.WithdrawalAmt
	}

	return total
}

// calculateStatementDays calculates number of days in statement period
func calculateStatementDays(transactions []models.ClassifiedTransaction) int {
	if len(transactions) == 0 {
		return 30 // Default
	}

	firstDate, err1 := utils.ParseDate(transactions[0].Date)
	lastDate, err2 := utils.ParseDate(transactions[len(transactions)-1].Date)

	if err1 != nil || err2 != nil {
		return 30 // Default
	}

	days := int(lastDate.Sub(firstDate).Hours() / 24)
	if days <= 0 {
		return 30 // Default
	}

	return days
}

// calculateFixedVsVariable calculates fixed vs variable expense percentages
func calculateFixedVsVariable(transactions []models.ClassifiedTransaction) (float64, float64) {
	totalExpense := 0.0
	recurringExpense := 0.0

	for _, txn := range transactions {
		// Only count operational expenses
		if txn.WithdrawalAmt == 0 || txn.DepositAmt > 0 {
			continue
		}

		// Skip investments
		if txn.Category == "Investment" || txn.Category == "Investments" || 
		   txn.Category == "Self_Transfer" {
			continue
		}
		if txn.Method == "RD" || txn.Method == "FD" || txn.Method == "SIP" || 
		   txn.Method == "Investment" {
			continue
		}

		amount := txn.WithdrawalAmt
		totalExpense += amount

		// Fixed expenses are recurring payments (rent, EMI, subscriptions, utilities)
		if txn.IsRecurring {
			recurringExpense += amount
		} else if txn.Method == "EMI" || txn.Category == "Loan" {
			recurringExpense += amount
		} else if strings.Contains(strings.ToUpper(txn.Narration), "RENT") || 
		          strings.Contains(strings.ToUpper(txn.Narration), "SUBSCRIPTION") {
			recurringExpense += amount
		}
	}

	fixedPercent := 0.0
	variablePercent := 0.0

	if totalExpense > 0 {
		fixedPercent = (recurringExpense / totalExpense) * 100
		variablePercent = ((totalExpense - recurringExpense) / totalExpense) * 100
	}

	return fixedPercent, variablePercent
}

