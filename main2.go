package main

import (
	"encoding/json"
	"log"
	"net/http"
)

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
	UPI        TransactionType `json:"UPI"`
	IMPS       TransactionType `json:"IMPS"`
	EMI        TransactionType `json:"EMI"`
	BillPaid   TransactionType `json:"BillPaid"`
	DebitCard  TransactionType `json:"DebitCard"`
	NetBanking TransactionType `json:"NetBanking"`
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
	FoodDelivery   float64 `json:"Food_Delivery"`
	Dining         float64 `json:"Dining"`
	Travel         float64 `json:"Travel"`
	Shopping       float64 `json:"Shopping"`
	Groceries      float64 `json:"Groceries"`
	BillsUtilities float64 `json:"Bills_Utilities"`
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

// enableCORS sets CORS headers to allow cross-origin requests
func enableCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	// Allow requests from localhost (development)
	if origin == "http://localhost:5173" || origin == "http://127.0.0.1:5173" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		// For production, you might want to restrict this to specific origins
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "3600")
}

// classifyHandler handles POST requests to /classify
func classifyHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	enableCORS(w, r)

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Create the response data
	response := ClassifyResponse{
		AccountSummary: AccountSummary{
			AccountNumberMasked: "XXXXXX9218",
			CustomerName:        "Raviraj Desai",
			StatementPeriod:     "1 Jan 2024 - 31 Dec 2024",
			Year:                "2024",
			OpeningBalance:      25340.50,
			ClosingBalance:      32890.25,
			TotalIncome:         1384500.00,
			TotalExpense:        1332950.25,
			NetSavings:          51549.75,
			SavingsRatePercent:  3.7,
		},
		TransactionBreakdown: TransactionBreakdown{
			UPI:        TransactionType{Amount: 198300, Count: 450},
			IMPS:       TransactionType{Amount: 150000, Count: 12},
			EMI:        TransactionType{Amount: 360000, Count: 24},
			BillPaid:   TransactionType{Amount: 184200, Count: 36},
			DebitCard:  TransactionType{Amount: 125000, Count: 45},
			NetBanking: TransactionType{Amount: 315450, Count: 15},
		},
		TopBeneficiaries: []TopBeneficiary{
			{Name: "Ramesh Kumar (Rent)", Amount: 240000, Type: "IMPS"},
			{Name: "HDFC Home Loan", Amount: 300000, Type: "EMI"},
			{Name: "LIC India", Amount: 50000, Type: "BillPay"},
			{Name: "Suresh Mechanic", Amount: 15000, Type: "UPI"},
			{Name: "Jio Fiber", Amount: 12000, Type: "BillPay"},
		},
		TopExpenses: []TopExpense{
			{Merchant: "Makemytrip", Date: "2024-05-15", Amount: 45000, Category: "Travel"},
			{Merchant: "Croma Electronics", Date: "2024-09-10", Amount: 32000, Category: "Electronics"},
			{Merchant: "Tanishq Jewellers", Date: "2024-10-25", Amount: 28000, Category: "Shopping"},
			{Merchant: "Apollo Hospital", Date: "2024-07-05", Amount: 15000, Category: "Health"},
			{Merchant: "Amazon India", Date: "2024-01-12", Amount: 12500, Category: "Shopping"},
		},
		MonthlySummary: []MonthlySummary{
			{Month: "Jan", Income: 115000, Expense: 108200, ClosingBalance: 30140, TopCategory: "Groceries", ExpenseSpikePercent: 0},
			{Month: "Feb", Income: 115000, Expense: 85400, ClosingBalance: 59890, TopCategory: "Investments", ExpenseSpikePercent: -21},
			{Month: "Mar", Income: 150000, Expense: 112300, ClosingBalance: 97590, TopCategory: "Shopping", ExpenseSpikePercent: 5},
			{Month: "Apr", Income: 115000, Expense: 92100, ClosingBalance: 120490, TopCategory: "Travel", ExpenseSpikePercent: -18},
			{Month: "May", Income: 115000, Expense: 135000, ClosingBalance: 100490, TopCategory: "Travel", ExpenseSpikePercent: 25},
			{Month: "Jun", Income: 115000, Expense: 105000, ClosingBalance: 110490, TopCategory: "Education", ExpenseSpikePercent: 0},
			{Month: "Jul", Income: 115000, Expense: 114000, ClosingBalance: 111490, TopCategory: "Health", ExpenseSpikePercent: 2},
			{Month: "Aug", Income: 115000, Expense: 88500, ClosingBalance: 137990, TopCategory: "Investments", ExpenseSpikePercent: -20},
			{Month: "Sep", Income: 115000, Expense: 116000, ClosingBalance: 136990, TopCategory: "Electronics", ExpenseSpikePercent: 3},
			{Month: "Oct", Income: 115000, Expense: 128000, ClosingBalance: 123990, TopCategory: "Shopping", ExpenseSpikePercent: 15},
			{Month: "Nov", Income: 115000, Expense: 109000, ClosingBalance: 129990, TopCategory: "Dining", ExpenseSpikePercent: 0},
			{Month: "Dec", Income: 125000, Expense: 142000, ClosingBalance: 112990, TopCategory: "Gifts", ExpenseSpikePercent: 30},
		},
		CategorySummary: CategorySummary{
			FoodDelivery:   15000,
			Dining:         12000,
			Travel:         45000,
			Shopping:       35000,
			Groceries:      25000,
			BillsUtilities: 18000,
		},
		MerchantSummary: MerchantSummary{
			Amazon:   25000,
			Flipkart: 15000,
			Swiggy:   8000,
			Zomato:   7000,
			Uber:     5000,
		},
		TransactionTrends: TransactionTrends{
			HighestSpendMonth: "May",
			LargestCategory:   "Travel",
		},
		RecommendedProducts: []RecommendedProduct{
			{ID: 1, ProductName: "IndianOil HDFC Bank Credit Card", Type: "Credit Card", Reason: "You spent ₹15,000 on Fuel last month. Save 5% on fuel spends.", Icon: "Fuel", ActionLink: "#"},
			{ID: 2, ProductName: "HDFC Life Click 2 Protect", Type: "Insurance", Reason: "No active term insurance detected. Secure your family's future.", Icon: "Shield", ActionLink: "#"},
			{ID: 3, ProductName: "HDFC Sky Demat Account", Type: "Investment", Reason: "You have a healthy savings balance. Start investing in stocks & MFs.", Icon: "TrendingUp", ActionLink: "#"},
		},
		PredictiveInsights: PredictiveInsights{
			Projected30DaySpend:     112500,
			PredictedLowBalanceDate: "25th Dec",
			UpcomingEMIImpact:       15000,
			SavingsRecommendation:   "Move ₹10k to FD to earn 7% interest",
		},
		CashFlowScore: CashFlowScore{
			Score:   78,
			Status:  "Healthy",
			Insight: "Your cash flow is positive. You are saving 20% of your income.",
		},
		SalaryUtilization: SalaryUtilization{
			SpentFirst3Days:  45,
			SpentFirst7Days:  60,
			SpentFirst15Days: 75,
			DaysSalaryLasts:  22,
			FixedExpenses:    40,
			VariableExpenses: 35,
		},
		BehaviourInsights: []BehaviourInsight{
			{Type: "Weekend Spender", Insight: "You spend 40% more on weekends compared to weekdays."},
			{Type: "Impulse Buying", Insight: "Great job! Impulse purchases are down by 5% this month."},
		},
		RecurringPayments: []RecurringPayment{
			{Name: "Netflix", Amount: 649, DayOfMonth: 5, Pattern: "Monthly"},
			{Name: "Spotify", Amount: 119, DayOfMonth: 10, Pattern: "Monthly"},
		},
		SavingsOpportunities: []SavingsOpportunity{
			{Category: "Switch to Annual Plan", PotentialSave: 1200, Action: "Switch", Difficulty: "Easy", Impact: "High"},
			{Category: "Reduce Dining Out", PotentialSave: 3000, Action: "Limit", Difficulty: "Medium", Impact: "Medium"},
		},
		FraudRisk: FraudRisk{
			RiskLevel: "Low",
			RecentAlerts: []FraudAlert{
				{Amount: 45000, Merchant: "Unknown Vendor"},
			},
		},
		BigTicketMovements: []BigTicketMovement{
			{Description: "Laptop Purchase", Amount: 85000, Date: "2024-11-15", Type: "Debit", Category: "Electronics", Impact: "High Impact"},
		},
		TaxInsights: TaxInsights{
			PotentialSave:    15000,
			MissedDeductions: []string{"Invest in ELSS", "Medical Insurance"},
		},
	}

	// Encode and send JSON response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

func main2() {
	// Register the handler
	http.HandleFunc("/classify", classifyHandler)

	// Start the server
	log.Println("Server starting on :8080")
	log.Println("POST endpoint available at: http://localhost:8080/classify")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
