package analytics

import (
	"fmt"
	"statement_analysis_engine_rules/models"
	"statement_analysis_engine_rules/utils"
)

// CalculateAccountSummary calculates account summary from transactions
func CalculateAccountSummary(
	accountNo string,
	customerName string,
	statementPeriod string,
	openingBalance float64,
	closingBalance float64,
	transactions []models.ClassifiedTransaction,
) models.AccountSummary {
	totalIncome := 0.0
	totalExpense := 0.0

	for _, txn := range transactions {
		if txn.IsIncome {
			totalIncome += txn.DepositAmt
		} else {
			totalExpense += txn.WithdrawalAmt
		}
	}

	netSavings := totalIncome - totalExpense
	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = (netSavings / totalIncome) * 100
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
