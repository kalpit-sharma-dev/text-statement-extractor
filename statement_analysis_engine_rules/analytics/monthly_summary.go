package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"classify/statement_analysis_engine_rules/utils"
	"sort"
)

// CalculateMonthlySummary calculates monthly summary
func CalculateMonthlySummary(transactions []models.ClassifiedTransaction) []models.MonthlySummary {
	monthlyData := make(map[string]*models.MonthlySummary)

	// Investment categories/methods to exclude from expenses
	investmentCategories := map[string]bool{
		"Investment":    true,
		"Investments":   true,
		"Self_Transfer": true,
	}
	investmentMethods := map[string]bool{
		"RD":         true,
		"FD":         true,
		"SIP":        true,
		"Investment": true,
	}

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

		// Count income (deposits)
		if txn.DepositAmt > 0 && txn.WithdrawalAmt == 0 {
			monthlyData[month].Income += txn.DepositAmt
		}
		
		// Count ONLY operational expenses (withdrawals) - EXCLUDE investments
		if txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
			// Check if it's an investment
			isInvestment := investmentCategories[txn.Category] || investmentMethods[txn.Method]
			
			if !isInvestment {
				// Only count as expense if it's NOT an investment
				monthlyData[month].Expense += txn.WithdrawalAmt
			}
		}
		
		// Handle edge case where both amounts exist (shouldn't happen, but handle it)
		if txn.DepositAmt > 0 && txn.WithdrawalAmt > 0 {
			if txn.DepositAmt > txn.WithdrawalAmt {
				monthlyData[month].Income += (txn.DepositAmt - txn.WithdrawalAmt)
			} else {
				netWithdrawal := txn.WithdrawalAmt - txn.DepositAmt
				isInvestment := investmentCategories[txn.Category] || investmentMethods[txn.Method]
				if !isInvestment {
					monthlyData[month].Expense += netWithdrawal
				}
			}
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
		// Handle unknown months (indexOf returns -1) - put them at the end
		if idxI == -1 {
			return false // Unknown month goes after known months
		}
		if idxJ == -1 {
			return true // Known month goes before unknown months
		}
		return idxI < idxJ
	})

	return monthlyList
}

func getTopCategoryForMonth(transactions []models.ClassifiedTransaction, month string) string {
	categoryMap := make(map[string]float64)

	// Exclude investment categories from top category calculation
	investmentCategories := map[string]bool{
		"Investment":    true,
		"Investments":   true,
		"Self_Transfer": true,
	}
	investmentMethods := map[string]bool{
		"RD":         true,
		"FD":         true,
		"SIP":        true,
		"Investment": true,
	}

	for _, txn := range transactions {
		// Only count operational expenses (withdrawals) for category breakdown
		if utils.GetMonthName(txn.Date) == month && 
		   txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
			
			// Skip investment categories
			isInvestment := investmentCategories[txn.Category] || investmentMethods[txn.Method]
			if isInvestment {
				continue
			}
			
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
