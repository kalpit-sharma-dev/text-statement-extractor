package main

import (
	"fmt"
	"log"
)

func main() {
	// Extract from TXT file
	statement, err := ReadAccountStatementFromTxt("Acct_Statement_XXXXXXXX1725_17122025.txt")
	if err != nil {
		log.Fatal(err)
	}

	// Use the extracted data
	fmt.Printf("Account: %s\n", statement.AccountInfo.AccountNo)
	fmt.Printf("Transactions: %d\n", len(statement.Transactions))
	fmt.Printf("Opening Balance: %.2f\n", statement.Summary.OpeningBalance)
	fmt.Printf("Closing Balance: %.2f\n", statement.Summary.ClosingBalance)
}
