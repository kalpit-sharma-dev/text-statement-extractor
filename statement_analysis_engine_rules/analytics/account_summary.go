package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"classify/statement_analysis_engine_rules/utils"
	"fmt"
)

// CalculateAccountSummary calculates account summary from transactions
// If statementTotalCredits and statementTotalDebits are provided (>0), they will be used instead of calculating from transactions
func CalculateAccountSummary(
	accountNo string,
	customerName string,
	statementPeriod string,
	openingBalance float64,
	closingBalance float64,
	transactions []models.ClassifiedTransaction,
) models.AccountSummary {
	return CalculateAccountSummaryWithTotals(
		accountNo, customerName, statementPeriod,
		openingBalance, closingBalance,
		transactions, 0.0, 0.0,
	)
}

// CalculateAccountSummaryWithTotals calculates account summary with optional statement totals
func CalculateAccountSummaryWithTotals(
	accountNo string,
	customerName string,
	statementPeriod string,
	openingBalance float64,
	closingBalance float64,
	transactions []models.ClassifiedTransaction,
	statementTotalCredits float64, // Use this if > 0 (official bank total)
	statementTotalDebits float64, // Use this if > 0 (official bank total)
) models.AccountSummary {
	totalIncome := 0.0
	totalExpense := 0.0
	totalInvestments := 0.0

	// Investment categories/methods to exclude from expenses
	// These represent wealth accumulation, savings, or money movement (not consumption)
	investmentCategories := map[string]bool{
		"Investment":    true,
		"Investments":   true,
		"Self_Transfer": true, // Transfers to own accounts (savings/investment accounts)
	}
	investmentMethods := map[string]bool{
		"RD":         true,
		"FD":         true,
		"SIP":        true,
		"Investment": true,
	}

	// If statement totals are provided, use them (they're the official bank totals)
	// Check if both are >= 0 (meaning they were explicitly set, even if one is 0)
	// We use a threshold check: if credits > 0 OR debits > 0, assume they were provided
	// This handles cases where one might legitimately be 0
	if statementTotalCredits > 0 || statementTotalDebits > 0 {
		// Use official statement totals for income
		totalIncome = statementTotalCredits
		
		// For expense vs investment breakdown, we need to calculate from transactions
		// because bank statement doesn't separate investments from expenses
		for _, txn := range transactions {
			if txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
				// Check if it's an investment
				isInvestment := investmentCategories[txn.Category] || investmentMethods[txn.Method]
				
				if isInvestment {
					totalInvestments += txn.WithdrawalAmt
				} else {
					totalExpense += txn.WithdrawalAmt
				}
			}
		}
	} else {
		// Calculate from transactions
		for _, txn := range transactions {
			// Count deposits as income
			// Only count if DepositAmt > 0 and WithdrawalAmt == 0 (to avoid double counting)
			if txn.DepositAmt > 0 && txn.WithdrawalAmt == 0 {
				totalIncome += txn.DepositAmt
			}

			// Count withdrawals - separate into expenses vs investments
			// Only count if WithdrawalAmt > 0 and DepositAmt == 0 (to avoid double counting)
			if txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
				// Check if it's an investment
				isInvestment := investmentCategories[txn.Category] || investmentMethods[txn.Method]
				
				if isInvestment {
					totalInvestments += txn.WithdrawalAmt
				} else {
					totalExpense += txn.WithdrawalAmt
				}
			}

			// If a transaction has both (shouldn't happen, but handle it):
			// Use the larger amount and treat it as the transaction type
			if txn.DepositAmt > 0 && txn.WithdrawalAmt > 0 {
				if txn.DepositAmt > txn.WithdrawalAmt {
					// Net deposit
					totalIncome += (txn.DepositAmt - txn.WithdrawalAmt)
				} else {
					// Net withdrawal
					netWithdrawal := txn.WithdrawalAmt - txn.DepositAmt
					isInvestment := investmentCategories[txn.Category] || investmentMethods[txn.Method]
					if isInvestment {
						totalInvestments += netWithdrawal
					} else {
						totalExpense += netWithdrawal
					}
				}
			}
		}
	}

	// Net Savings = Total Income - Total Expense - Total Investments
	// This shows how much money is left after expenses and investments
	// Positive = surplus, Negative = deficit
	// Note: Opening balance is NOT included - it's just the starting point
	netSavings := totalIncome - totalExpense - totalInvestments

	// Savings Rate = ((Total Income - Total Expense) / Total Income) * 100
	// This shows what percentage of income is NOT spent on operational expenses
	// Investments are considered savings (wealth accumulation), not expenses
	// So savings rate = (Income - Operational Expenses) / Income
	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = ((totalIncome - totalExpense) / totalIncome) * 100
	} else if totalIncome == 0 && (totalExpense > 0 || totalInvestments > 0) {
		// If no income but there are expenses/investments, savings rate cannot be calculated meaningfully
		// Set to a large negative number to indicate expenses without income
		savingsRate = -999.0
	}

	// Extract year from statement period
	year := utils.GetYear(statementPeriod)
	if year == "" {
		year = "2024" // Default
	}

	return models.AccountSummary{
		AccountNumberMasked: utils.MaskAccountNumber(accountNo),
		CustomerName:        customerName,
		StatementPeriod:     statementPeriod,
		Year:                year,
		OpeningBalance:      openingBalance,
		ClosingBalance:      closingBalance,
		TotalIncome:         totalIncome,
		TotalExpense:        totalExpense,
		TotalInvestments:    totalInvestments,
		NetSavings:          netSavings,
		SavingsRatePercent:  savingsRate,
	}
}

// FormatStatementPeriod formats statement period string
func FormatStatementPeriod(fromDate, toDate string) string {
	from, _ := utils.ParseDate(fromDate)
	to, _ := utils.ParseDate(toDate)

	if from.IsZero() || to.IsZero() {
		return fmt.Sprintf("%s - %s", fromDate, toDate)
	}

	return fmt.Sprintf("%s - %s", utils.FormatDate(from, "DD/MM/YYYY"), utils.FormatDate(to, "DD/MM/YYYY"))
}
