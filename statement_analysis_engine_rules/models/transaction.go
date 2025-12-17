package models

// ClassifiedTransaction represents a transaction with classification information
type ClassifiedTransaction struct {
	// Original transaction data
	Date           string
	Narration      string
	ChequeRefNo    string
	ValueDate      string
	WithdrawalAmt  float64
	DepositAmt     float64
	ClosingBalance float64

	// Classification fields
	Method     string // UPI, IMPS, NEFT, RTGS, DebitCard, NetBanking, EMI, ACH, etc.
	Category   string // Food_Delivery, Dining, Travel, Shopping, Groceries, Bills_Utilities, etc.
	Merchant   string // Extracted merchant name
	Beneficiary string // Beneficiary name for transfers
	IsIncome   bool   // true if deposit, false if withdrawal
	IsRecurring bool  // true if recurring payment detected
}

// TransactionList is a collection of classified transactions
type TransactionList []ClassifiedTransaction

