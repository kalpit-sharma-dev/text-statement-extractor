package main

import (
	"fmt"
	"log"
)

// Example usage of the extraction library
func main() {
	// Extract statement from TXT file
	statement, err := ReadAccountStatementFromTxt("Acct_Statement_XXXXXXXX1725_17122025.txt")
	if err != nil {
		log.Fatal(err)
	}

	// Access extracted data
	fmt.Printf("Account Number: %s\n", statement.AccountInfo.AccountNo)
	fmt.Printf("Account Holder: %s\n", statement.AccountInfo.AccountHolderName)
	fmt.Printf("Total Transactions: %d\n", len(statement.Transactions))
	fmt.Printf("Opening Balance: %.2f\n", statement.Summary.OpeningBalance)
	fmt.Printf("Closing Balance: %.2f\n", statement.Summary.ClosingBalance)
	fmt.Printf("Total Debits: %.2f\n", statement.Summary.TotalDebits)
	fmt.Printf("Total Credits: %.2f\n", statement.Summary.TotalCredits)
	fmt.Printf("Debit Count: %d\n", statement.Summary.DebitCount)
	fmt.Printf("Credit Count: %d\n", statement.Summary.CreditCount)
}
