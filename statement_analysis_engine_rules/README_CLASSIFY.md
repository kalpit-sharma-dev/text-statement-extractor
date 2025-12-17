# Transaction Classification Library

A comprehensive Go library for classifying and analyzing bank statement transactions based on narration patterns.

## Package Structure

```
statement_analysis_engine_rules/
├── analyzer/          # Main analyzer interface
│   ├── analyzer.go    # Core analyzer logic
│   └── config.go      # Configuration
│
├── models/            # Data models
│   ├── transaction.go # Transaction models
│   └── response.go    # Response models
│
├── rules/             # Classification rules
│   ├── method_rules.go      # Payment method classification
│   ├── category_rules.go    # Category classification
│   └── beneficiary_rules.go # Beneficiary extraction
│
├── classifier/        # Classification engine
│   └── classifier.go  # Main classifier
│
├── analytics/         # Analytics modules
│   ├── account_summary.go
│   ├── transaction_breakdown.go
│   ├── monthly_summary.go
│   ├── category_summary.go
│   ├── merchant_summary.go
│   ├── top_expenses.go
│   ├── beneficiaries.go
│   ├── recurring.go
│   ├── fraud.go
│   ├── cashflow.go
│   ├── predictive.go
│   └── tax.go
│
└── utils/             # Utility functions
    ├── normalize.go
    ├── date.go
    └── sort.go
```

## Features

- **Transaction Classification**: Automatically classifies transactions by:
  - Payment Method (UPI, IMPS, NEFT, RTGS, EMI, ACH, DebitCard, NetBanking)
  - Category (Food_Delivery, Dining, Travel, Shopping, Groceries, Bills_Utilities, etc.)
  - Merchant extraction
  - Beneficiary identification

- **Comprehensive Analytics**:
  - Account summary with savings rate
  - Transaction breakdown by method
  - Monthly summaries with expense spikes
  - Category and merchant summaries
  - Top expenses and beneficiaries
  - Recurring payment detection
  - Fraud risk assessment
  - Cash flow scoring
  - Predictive insights
  - Tax optimization suggestions

## Usage

### Basic Usage

```go
package main

import (
    "statement_analysis_engine_rules/analyzer"
    "statement_analysis_engine_rules/classifier"
    "statement_analysis_engine_rules/models"
)

func main() {
    // 1. Convert extracted transactions
    transactions := []models.ClassifiedTransaction{
        classifier.ConvertFromTxtTransaction(
            "01/04/24", "UPI-AMAZON-AMAZON@PAYTM-123456-UPI",
            "0000123456", "01/04/24", 5000.00, 0, 100000.00,
        ),
    }

    // 2. Create analyzer
    analyzerInstance := analyzer.NewAnalyzer()
    analyzerInstance.AddTransactions(transactions)

    // 3. Analyze
    response := analyzerInstance.Analyze(
        "08821130001725",
        "Customer Name",
        "01/04/2024 - 31/03/2025",
        379562.39,
        40939.84,
    )

    // 4. Use response
    fmt.Printf("Total Income: %.2f\n", response.AccountSummary.TotalIncome)
    fmt.Printf("Total Expense: %.2f\n", response.AccountSummary.TotalExpense)
}
```

### Integration with Statement Extraction

```go
// Extract statement using extract_statement.go
statement, err := ReadAccountStatementFromTxt("statement.txt")
if err != nil {
    log.Fatal(err)
}

// Convert to classified transactions
classifiedTransactions := make([]models.ClassifiedTransaction, 0)
for _, txn := range statement.Transactions {
    classifiedTxn := classifier.ConvertFromTxtTransaction(
        txn.Date, txn.Narration, txn.ChequeRefNo, txn.ValueDate,
        txn.WithdrawalAmt, txn.DepositAmt, txn.ClosingBalance,
    )
    classifiedTransactions = append(classifiedTransactions, classifiedTxn)
}

// Analyze
analyzerInstance := analyzer.NewAnalyzer()
analyzerInstance.AddTransactions(classifiedTransactions)
response := analyzerInstance.Analyze(
    statement.AccountInfo.AccountNo,
    statement.AccountInfo.AccountHolderName,
    fmt.Sprintf("%s - %s", statement.StatementPeriod.FromDate, statement.StatementPeriod.ToDate),
    statement.Summary.OpeningBalance,
    statement.Summary.ClosingBalance,
)
```

## Classification Rules

### Payment Methods

The library identifies payment methods based on narration patterns:
- **UPI**: Contains "UPI", "PAYTM", "PHONEPE", "GOOGLEPAY", "@YBL", "@PAYTM", etc.
- **IMPS**: Contains "IMPS", "INSTANT PAYMENT"
- **NEFT**: Contains "NEFT", "NATIONAL ELECTRONIC FUND TRANSFER"
- **RTGS**: Contains "RTGS", "REAL TIME GROSS SETTLEMENT"
- **EMI**: Contains "EMI", "LOAN", "INSTALLMENT"
- **ACH**: Contains "ACH", "AUTOMATED CLEARING HOUSE"
- **DebitCard**: Contains "POS", "DEBIT CARD", "ATM", "SWIPE"
- **NetBanking**: Contains "NET BANKING", "ONLINE BANKING", "IB"

### Categories

Categories are identified based on merchant names and narration:
- **Food_Delivery**: Swiggy, Zomato, Uber Eats
- **Dining**: Restaurants, cafes, hotels
- **Travel**: Uber, Ola, MakeMyTrip, fuel stations
- **Shopping**: Amazon, Flipkart, Myntra, electronics stores
- **Groceries**: DMart, Big Bazaar, supermarkets
- **Bills_Utilities**: Electricity, water, phone, insurance

## Response Structure

The `ClassifyResponse` includes:

- `AccountSummary`: Account details, balances, income, expenses, savings rate
- `TransactionBreakdown`: Breakdown by payment method
- `TopBeneficiaries`: Top beneficiaries with amounts
- `TopExpenses`: Largest expenses
- `MonthlySummary`: Month-wise analysis
- `CategorySummary`: Category-wise spending
- `MerchantSummary`: Merchant-wise spending
- `TransactionTrends`: Spending trends
- `RecommendedProducts`: Product recommendations
- `PredictiveInsights`: Future predictions
- `CashFlowScore`: Cash flow health score
- `SalaryUtilization`: Salary spending patterns
- `BehaviourInsights`: Spending behavior insights
- `RecurringPayments`: Recurring payment detection
- `SavingsOpportunities`: Savings suggestions
- `FraudRisk`: Fraud risk assessment
- `BigTicketMovements`: Large transactions
- `TaxInsights`: Tax optimization suggestions

## Customization

You can extend the classification rules by modifying files in the `rules/` package:
- Add new patterns in `method_rules.go`
- Add new categories in `category_rules.go`
- Enhance beneficiary extraction in `beneficiary_rules.go`

## Examples

See `example_integration.go` for a complete integration example with statement extraction.

## License

This library is part of the classify project.

