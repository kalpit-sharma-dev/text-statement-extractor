package models

// ClassificationMetadata stores "why" a classification happened (for explainability)
// This implements the principle: "Store 'why' a classification happened"
type ClassificationMetadata struct {
	Confidence      float64  `json:"confidence"`       // 0.0 to 1.0
	MatchedKeywords []string `json:"matchedKeywords"` // Keywords that matched
	Gateway         string   `json:"gateway"`          // Payment gateway (BillDesk, PayU, etc.) - separate concept
	Channel         string   `json:"channel"`          // Payment channel (UPI, POS, ECS, etc.) - separate concept
	RuleVersion     string   `json:"ruleVersion"`      // Version of rules used
	Reason          string   `json:"reason"`          // Human-readable explanation
}

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

	// Classification fields (separated concepts: Channel, Gateway, Merchant, Intent)
	Method     string // Payment method/channel: UPI, IMPS, NEFT, RTGS, DebitCard, NetBanking, EMI, ACH, etc.
	Category   string // Intent/category: Food_Delivery, Dining, Travel, Shopping, Groceries, Bills_Utilities, Loan, etc.
	Merchant   string // Canonical merchant name (normalized)
	Beneficiary string // Beneficiary name for transfers
	IsIncome   bool   // true if deposit, false if withdrawal
	IsRecurring bool  // true if recurring payment detected
	
	// Recurring payment metadata (if IsRecurring is true)
	RecurringMetadata RecurringMetadata `json:"recurringMetadata,omitempty"`
	
	// Classification metadata (for explainability and debugging)
	ClassificationMetadata ClassificationMetadata `json:"classificationMetadata,omitempty"`
}

// RecurringMetadata stores recurring payment detection details
type RecurringMetadata struct {
	IsRecurring bool    `json:"isRecurring"`
	Confidence  int     `json:"confidence"`  // 0-100 confidence score
	Frequency   string  `json:"frequency"`   // MONTHLY, WEEKLY, QUARTERLY
	FirstSeen   string  `json:"firstSeen"`   // Date of first occurrence
	LastSeen    string  `json:"lastSeen"`    // Date of last occurrence
	Count       int     `json:"count"`       // Number of occurrences
	Pattern     string  `json:"pattern"`     // Pattern description
}

// TransactionList is a collection of classified transactions
type TransactionList []ClassifiedTransaction

