package main

import (
	"encoding/json"
	"fmt"
	"log"
	"statement_analysis_engine_rules/analyzer"
	"statement_analysis_engine_rules/classifier"
	"statement_analysis_engine_rules/models"
)

// This example shows how to integrate statement extraction with classification
func main() {
	// Step 1: Extract statement from TXT file (using extract_statement.go)
	statement, err := ReadAccountStatementFromTxt("Acct_Statement_XXXXXXXX1725_17122025.txt")
	if err != nil {
		log.Fatal("Failed to extract statement:", err)
	}

	// Step 2: Convert extracted transactions to classified transactions
	classifiedTransactions := make([]models.ClassifiedTransaction, 0, len(statement.Transactions))
	for _, txn := range statement.Transactions {
		classifiedTxn := classifier.ConvertFromTxtTransaction(
			txn.Date,
			txn.Narration,
			txn.ChequeRefNo,
			txn.ValueDate,
			txn.WithdrawalAmt,
			txn.DepositAmt,
			txn.ClosingBalance,
		)
		classifiedTransactions = append(classifiedTransactions, classifiedTxn)
	}

	// Step 3: Create analyzer instance
	analyzerInstance := analyzer.NewAnalyzer()
	analyzerInstance.AddTransactions(classifiedTransactions)

	// Step 4: Format statement period
	statementPeriod := fmt.Sprintf("%s - %s",
		statement.StatementPeriod.FromDate,
		statement.StatementPeriod.ToDate,
	)

	// Step 5: Run analysis
	response := analyzerInstance.Analyze(
		statement.AccountInfo.AccountNo,
		statement.AccountInfo.AccountHolderName,
		statementPeriod,
		statement.Summary.OpeningBalance,
		statement.Summary.ClosingBalance,
	)

	// Step 6: Output results as JSON
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatal("Failed to marshal response:", err)
	}

	fmt.Println("=== Classification Analysis Complete ===")
	fmt.Println(string(jsonData))

	// Step 7: Access specific data
	fmt.Println("\n=== Summary ===")
	fmt.Printf("Total Transactions: %d\n", len(classifiedTransactions))
	fmt.Printf("Account: %s\n", response.AccountSummary.AccountNumberMasked)
	fmt.Printf("Customer: %s\n", response.AccountSummary.CustomerName)
	fmt.Printf("Opening Balance: %.2f\n", response.AccountSummary.OpeningBalance)
	fmt.Printf("Closing Balance: %.2f\n", response.AccountSummary.ClosingBalance)
	fmt.Printf("Total Income: %.2f\n", response.AccountSummary.TotalIncome)
	fmt.Printf("Total Expense: %.2f\n", response.AccountSummary.TotalExpense)
	fmt.Printf("Net Savings: %.2f\n", response.AccountSummary.NetSavings)
	fmt.Printf("Savings Rate: %.2f%%\n", response.AccountSummary.SavingsRatePercent)

	fmt.Println("\n=== Transaction Breakdown ===")
	fmt.Printf("UPI: %.2f (%d transactions)\n",
		response.TransactionBreakdown.UPI.Amount,
		response.TransactionBreakdown.UPI.Count)
	fmt.Printf("IMPS: %.2f (%d transactions)\n",
		response.TransactionBreakdown.IMPS.Amount,
		response.TransactionBreakdown.IMPS.Count)
	fmt.Printf("EMI: %.2f (%d transactions)\n",
		response.TransactionBreakdown.EMI.Amount,
		response.TransactionBreakdown.EMI.Count)

	fmt.Println("\n=== Top Expenses ===")
	for i, expense := range response.TopExpenses {
		if i >= 5 {
			break
		}
		fmt.Printf("%d. %s - %.2f (%s)\n",
			i+1, expense.Merchant, expense.Amount, expense.Category)
	}
}
