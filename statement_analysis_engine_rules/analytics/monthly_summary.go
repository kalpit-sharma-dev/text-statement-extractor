package analytics

import (
	"sort"
	"statement_analysis_engine_rules/models"
	"statement_analysis_engine_rules/utils"
)

// CalculateMonthlySummary calculates monthly summary
func CalculateMonthlySummary(transactions []models.ClassifiedTransaction) []models.MonthlySummary {
	monthlyData := make(map[string]*models.MonthlySummary)

	for _, txn := range transactions {
		month := utils.GetMonthName(txn.Date)
		if month == "" {
			continue
		}

		if monthlyData[month] == nil {
			monthlyData[month] = &models.MonthlySummary{
				Month: month,
			}
		}

		if txn.IsIncome {
			monthlyData[month].Income += txn.DepositAmt
		} else {
			monthlyData[month].Expense += txn.WithdrawalAmt
		}

		// Update closing balance (use last transaction's balance for the month)
		monthlyData[month].ClosingBalance = txn.ClosingBalance
	}

	// Calculate category breakdown per month and expense spike
	monthlyList := make([]models.MonthlySummary, 0, len(monthlyData))
	for month, data := range monthlyData {
		// Calculate top category for the month
		data.TopCategory = getTopCategoryForMonth(transactions, month)

		// Calculate expense spike (simplified - compare with previous month average)
		data.ExpenseSpikePercent = calculateExpenseSpike(monthlyData, month, data.Expense)

		monthlyList = append(monthlyList, *data)
	}

	// Sort by month
	sort.Slice(monthlyList, func(i, j int) bool {
		months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
		idxI := indexOf(months, monthlyList[i].Month)
		idxJ := indexOf(months, monthlyList[j].Month)
		return idxI < idxJ
	})

	return monthlyList
}

func getTopCategoryForMonth(transactions []models.ClassifiedTransaction, month string) string {
	categoryMap := make(map[string]float64)

	for _, txn := range transactions {
		if utils.GetMonthName(txn.Date) == month && !txn.IsIncome {
			categoryMap[txn.Category] += txn.WithdrawalAmt
		}
	}

	maxAmount := 0.0
	topCategory := "Other"
	for category, amount := range categoryMap {
		if amount > maxAmount {
			maxAmount = amount
			topCategory = category
		}
	}

	return topCategory
}

func calculateExpenseSpike(monthlyData map[string]*models.MonthlySummary, currentMonth string, currentExpense float64) int {
	// Simplified calculation - compare with average of other months
	totalExpense := 0.0
	count := 0
	for month, data := range monthlyData {
		if month != currentMonth {
			totalExpense += data.Expense
			count++
		}
	}

	if count == 0 {
		return 0
	}

	avgExpense := totalExpense / float64(count)
	if avgExpense == 0 {
		return 0
	}

	spike := ((currentExpense - avgExpense) / avgExpense) * 100
	return int(spike)
}

func indexOf(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}
