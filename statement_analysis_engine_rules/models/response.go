package models

// AccountSummary represents the account summary information
type AccountSummary struct {
	AccountNumberMasked string  `json:"accountNumberMasked"`
	CustomerName        string  `json:"customerName"`
	StatementPeriod     string  `json:"statementPeriod"`
	Year                string  `json:"year"`
	OpeningBalance      float64 `json:"openingBalance"`
	ClosingBalance      float64 `json:"closingBalance"`
	TotalIncome         float64 `json:"totalIncome"`
	TotalExpense        float64 `json:"totalExpense"`
	NetSavings          float64 `json:"netSavings"`
	SavingsRatePercent  float64 `json:"savingsRatePercent"`
}

// TransactionType represents transaction breakdown by type
type TransactionType struct {
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
}

// TransactionBreakdown represents breakdown by transaction type
type TransactionBreakdown struct {
	UPI           TransactionType `json:"UPI"`
	IMPS          TransactionType `json:"IMPS"`
	NEFT          TransactionType `json:"NEFT"`
	RTGS          TransactionType `json:"RTGS"`
	EMI           TransactionType `json:"EMI"`
	BillPaid      TransactionType `json:"BillPaid"`
	DebitCard     TransactionType `json:"DebitCard"`
	ATMWithdrawal TransactionType `json:"ATMWithdrawal"`
	NetBanking    TransactionType `json:"NetBanking"`
	Salary        TransactionType `json:"Salary"`
	RD            TransactionType `json:"RD"`
	FD            TransactionType `json:"FD"`
	SIP           TransactionType `json:"SIP"`
	Interest      TransactionType `json:"Interest"`
	Cheque        TransactionType `json:"Cheque"`
	Dividend      TransactionType `json:"Dividend"`
	Investment    TransactionType `json:"Investment"`
	Other         TransactionType `json:"Other"`
}

// TopBeneficiary represents a top beneficiary
type TopBeneficiary struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Type   string  `json:"type"`
}

// TopExpense represents a top expense
type TopExpense struct {
	Merchant string  `json:"merchant"`
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

// MonthlySummary represents monthly summary data
type MonthlySummary struct {
	Month               string  `json:"month"`
	Income              float64 `json:"income"`
	Expense             float64 `json:"expense"`
	ClosingBalance      float64 `json:"closingBalance"`
	TopCategory         string  `json:"topCategory"`
	ExpenseSpikePercent int     `json:"expenseSpikePercent"`
}

// CategorySummary represents category-wise summary
type CategorySummary struct {
	Shopping       float64 `json:"Shopping"`
	BillsUtilities float64 `json:"Bills_Utilities"`
	Travel         float64 `json:"Travel"`
	Dining         float64 `json:"Dining"`
	Groceries      float64 `json:"Groceries"`
	FoodDelivery   float64 `json:"Food_Delivery"`
	Fuel           float64 `json:"Fuel"`
	Investments    float64 `json:"Investments"`
}

// MerchantSummary represents merchant-wise summary
type MerchantSummary struct {
	Amazon   float64 `json:"Amazon"`
	Flipkart float64 `json:"Flipkart"`
	Swiggy   float64 `json:"Swiggy"`
	Zomato   float64 `json:"Zomato"`
	Uber     float64 `json:"Uber"`
}

// TransactionTrends represents transaction trends
type TransactionTrends struct {
	HighestSpendMonth string `json:"highestSpendMonth"`
	LargestCategory   string `json:"largestCategory"`
}

// RecommendedProduct represents a recommended product
type RecommendedProduct struct {
	ID          int    `json:"id"`
	ProductName string `json:"productName"`
	Type        string `json:"type"`
	Reason      string `json:"reason"`
	Icon        string `json:"icon"`
	ActionLink  string `json:"actionLink"`
}

// PredictiveInsights represents predictive insights
type PredictiveInsights struct {
	Projected30DaySpend     float64 `json:"projected30DaySpend"`
	PredictedLowBalanceDate string  `json:"predictedLowBalanceDate"`
	UpcomingEMIImpact       float64 `json:"upcomingEMIImpact"`
	SavingsRecommendation   string  `json:"savingsRecommendation"`
}

// CashFlowScore represents cash flow score
type CashFlowScore struct {
	Score   int    `json:"score"`
	Status  string `json:"status"`
	Insight string `json:"insight"`
}

// SalaryUtilization represents salary utilization metrics
type SalaryUtilization struct {
	SpentFirst3Days  float64 `json:"spentFirst3Days"`
	SpentFirst7Days  float64 `json:"spentFirst7Days"`
	SpentFirst15Days float64 `json:"spentFirst15Days"`
	DaysSalaryLasts  int     `json:"daysSalaryLasts"`
	FixedExpenses    float64 `json:"fixedExpenses"`
	VariableExpenses float64 `json:"variableExpenses"`
}

// BehaviourInsight represents a behavior insight
type BehaviourInsight struct {
	Type    string `json:"type"`
	Insight string `json:"insight"`
}

// RecurringPayment represents a recurring payment
type RecurringPayment struct {
	Name       string  `json:"name"`
	Amount     float64 `json:"amount"`
	DayOfMonth int     `json:"dayOfMonth"`
	Pattern    string  `json:"pattern"`
}

// SavingsOpportunity represents a savings opportunity
type SavingsOpportunity struct {
	Category      string  `json:"category"`
	PotentialSave float64 `json:"potentialSave"`
	Action        string  `json:"action"`
	Difficulty    string  `json:"difficulty"`
	Impact        string  `json:"impact"`
}

// FraudAlert represents a fraud alert
type FraudAlert struct {
	Amount   float64 `json:"amount"`
	Merchant string  `json:"merchant"`
}

// FraudRisk represents fraud risk information
type FraudRisk struct {
	RiskLevel    string       `json:"riskLevel"`
	RecentAlerts []FraudAlert `json:"recentAlerts"`
}

// BigTicketMovement represents a big ticket movement
type BigTicketMovement struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Impact      string  `json:"impact"`
}

// TaxInsights represents tax insights
type TaxInsights struct {
	PotentialSave    float64  `json:"potentialSave"`
	MissedDeductions []string `json:"missedDeductions"`
}

// ClassifyResponse represents the complete response structure
type ClassifyResponse struct {
	AccountSummary       AccountSummary       `json:"accountSummary"`
	TransactionBreakdown TransactionBreakdown `json:"transactionBreakdown"`
	TopBeneficiaries     []TopBeneficiary     `json:"topBeneficiaries"`
	TopExpenses          []TopExpense         `json:"topExpenses"`
	MonthlySummary       []MonthlySummary     `json:"monthlySummary"`
	CategorySummary      CategorySummary      `json:"categorySummary"`
	MerchantSummary      MerchantSummary      `json:"merchantSummary"`
	TransactionTrends    TransactionTrends    `json:"transactionTrends"`
	RecommendedProducts  []RecommendedProduct `json:"recommendedProducts"`
	PredictiveInsights   PredictiveInsights   `json:"predictiveInsights"`
	CashFlowScore        CashFlowScore        `json:"cashFlowScore"`
	SalaryUtilization    SalaryUtilization    `json:"salaryUtilization"`
	BehaviourInsights    []BehaviourInsight   `json:"behaviourInsights"`
	RecurringPayments    []RecurringPayment   `json:"recurringPayments"`
	SavingsOpportunities []SavingsOpportunity `json:"savingsOpportunities"`
	FraudRisk            FraudRisk            `json:"fraudRisk"`
	BigTicketMovements   []BigTicketMovement  `json:"bigTicketMovements"`
	TaxInsights          TaxInsights          `json:"taxInsights"`
}
