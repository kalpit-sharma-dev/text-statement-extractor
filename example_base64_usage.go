package main

import (
	"fmt"
	"log"
	"os"
)

// Example usage with base64 encoded string
func main() {
	// Option 1: If you have a base64 string directly
	base64String := "your_base64_encoded_string_here"

	statement, err := ReadAccountStatementFromBase64(base64String)
	if err != nil {
		log.Fatal(err)
	}

	// Access extracted data
	fmt.Printf("Account Number: %s\n", statement.AccountInfo.AccountNo)
	fmt.Printf("Account Holder: %s\n", statement.AccountInfo.AccountHolderName)
	fmt.Printf("Total Transactions: %d\n", len(statement.Transactions))
	fmt.Printf("Opening Balance: %.2f\n", statement.Summary.OpeningBalance)
	fmt.Printf("Closing Balance: %.2f\n", statement.Summary.ClosingBalance)

	// Option 2: If you have a file containing base64 string
	// Read the base64 string from a file
	base64Bytes, err := os.ReadFile("statement_base64.txt")
	if err != nil {
		log.Fatal(err)
	}

	statement2, err := ReadAccountStatementFromBase64(string(base64Bytes))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nSecond statement - Account: %s\n", statement2.AccountInfo.AccountNo)
}
