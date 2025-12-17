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

	// If statement totals are provided, use them (they're the official bank totals)
	// Check if both are >= 0 (meaning they were explicitly set, even if one is 0)
	// We use a threshold check: if credits > 0 OR debits > 0, assume they were provided
	// This handles cases where one might legitimately be 0
	if statementTotalCredits > 0 || statementTotalDebits > 0 {
		// Use official statement totals (even if one is 0, that's valid)
		totalIncome = statementTotalCredits
		totalExpense = statementTotalDebits
	} else {
		// Calculate from transactions
		for _, txn := range transactions {
			// Count deposits as income
			// Only count if DepositAmt > 0 and WithdrawalAmt == 0 (to avoid double counting)
			if txn.DepositAmt > 0 && txn.WithdrawalAmt == 0 {
				totalIncome += txn.DepositAmt
			}

			// Count withdrawals as expenses
			// Only count if WithdrawalAmt > 0 and DepositAmt == 0 (to avoid double counting)
			if txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
				totalExpense += txn.WithdrawalAmt
			}

			// If a transaction has both (shouldn't happen, but handle it):
			// Use the larger amount and treat it as the transaction type
			if txn.DepositAmt > 0 && txn.WithdrawalAmt > 0 {
				if txn.DepositAmt > txn.WithdrawalAmt {
					// Net deposit
					totalIncome += (txn.DepositAmt - txn.WithdrawalAmt)
				} else {
					// Net withdrawal
					totalExpense += (txn.WithdrawalAmt - txn.DepositAmt)
				}
			}
		}
	}

	// Net Savings = Total Income - Total Expense
	// Opening balance is NOT included in this calculation
	// Opening balance is just the starting point, not part of income/expense
	netSavings := totalIncome - totalExpense

	// Savings Rate = (Net Savings / Total Income) * 100
	// This shows what percentage of income is saved
	// If expenses > income, savings rate will be negative (which is valid)
	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = (netSavings / totalIncome) * 100
	} else if totalIncome == 0 && totalExpense > 0 {
		// If no income but there are expenses, savings rate cannot be calculated meaningfully
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
