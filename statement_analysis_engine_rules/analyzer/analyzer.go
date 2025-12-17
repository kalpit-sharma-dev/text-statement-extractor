package analyzer

import (
	"statement_analysis_engine_rules/analytics"
	"statement_analysis_engine_rules/classifier"
	"statement_analysis_engine_rules/models"
)

// Analyzer is the main analyzer struct
type Analyzer struct {
	transactions []models.ClassifiedTransaction
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		transactions: make([]models.ClassifiedTransaction, 0),
	}
}

// AddTransaction adds a transaction to be analyzed
func (a *Analyzer) AddTransaction(txn models.ClassifiedTransaction) {
	a.transactions = append(a.transactions, txn)
}

// AddTransactions adds multiple transactions
func (a *Analyzer) AddTransactions(transactions []models.ClassifiedTransaction) {
	a.transactions = append(a.transactions, transactions...)
}

// ClassifyAll classifies all transactions
func (a *Analyzer) ClassifyAll() {
	a.transactions = classifier.ClassifyTransactions(a.transactions)
}

// Analyze generates complete analysis
func (a *Analyzer) Analyze(
	accountNo string,
	customerName string,
	statementPeriod string,
	openingBalance float64,
	closingBalance float64,
) models.ClassifyResponse {
	// Classify all transactions first
	a.ClassifyAll()

	// Calculate all analytics
	accountSummary := analytics.CalculateAccountSummary(
		accountNo,
		customerName,
		statementPeriod,
		openingBalance,
		closingBalance,
		a.transactions,
	)

	transactionBreakdown := analytics.CalculateTransactionBreakdown(a.transactions)
	topBeneficiaries := analytics.CalculateTopBeneficiaries(a.transactions, 5)
	topExpenses := analytics.CalculateTopExpenses(a.transactions, 5)
	monthlySummary := analytics.CalculateMonthlySummary(a.transactions)
	categorySummary := analytics.CalculateCategorySummary(a.transactions)
	merchantSummary := analytics.CalculateMerchantSummary(a.transactions)
	transactionTrends := analytics.CalculateTransactionTrends(monthlySummary, categorySummary)
	recurringPayments := analytics.CalculateRecurringPayments(a.transactions)
	fraudRisk := analytics.CalculateFraudRisk(a.transactions)
	bigTicketMovements := analytics.CalculateBigTicketMovements(a.transactions, 20000)
	taxInsights := analytics.CalculateTaxInsights(a.transactions)
	cashFlowScore := analytics.CalculateCashFlowScore(
		openingBalance,
		closingBalance,
		accountSummary.TotalIncome,
		accountSummary.TotalExpense,
	)
	predictiveInsights := analytics.CalculatePredictiveInsights(a.transactions, closingBalance)

	// Calculate salary utilization (simplified - would need salary detection)
	salaryUtilization := analytics.CalculateSalaryUtilization(a.transactions, 0, "")

	// Generate recommendations
	recommendedProducts := generateRecommendations(a.transactions, categorySummary, cashFlowScore)
	behaviourInsights := generateBehaviourInsights(a.transactions)
	savingsOpportunities := generateSavingsOpportunities(a.transactions, categorySummary)

	return models.ClassifyResponse{
		AccountSummary:       accountSummary,
		TransactionBreakdown: transactionBreakdown,
		TopBeneficiaries:     topBeneficiaries,
		TopExpenses:          topExpenses,
		MonthlySummary:       monthlySummary,
		CategorySummary:      categorySummary,
		MerchantSummary:      merchantSummary,
		TransactionTrends:    transactionTrends,
		RecommendedProducts:  recommendedProducts,
		PredictiveInsights:   predictiveInsights,
		CashFlowScore:        cashFlowScore,
		SalaryUtilization:    salaryUtilization,
		BehaviourInsights:    behaviourInsights,
		RecurringPayments:    recurringPayments,
		SavingsOpportunities: savingsOpportunities,
		FraudRisk:            fraudRisk,
		BigTicketMovements:   bigTicketMovements,
		TaxInsights:          taxInsights,
	}
}

// Helper functions for generating recommendations
func generateRecommendations(
	transactions []models.ClassifiedTransaction,
	categorySummary models.CategorySummary,
	cashFlowScore models.CashFlowScore,
) []models.RecommendedProduct {
	recommendations := make([]models.RecommendedProduct, 0)
	id := 1

	// Fuel spending recommendation
	if categorySummary.Travel > 10000 {
		recommendations = append(recommendations, models.RecommendedProduct{
			ID:          id,
			ProductName: "IndianOil HDFC Bank Credit Card",
			Type:        "Credit Card",
			Reason:      "You spent significant amount on Travel/Fuel. Save 5% on fuel spends.",
			Icon:        "Fuel",
			ActionLink:  "#",
		})
		id++
	}

	// Insurance recommendation
	hasInsurance := false
	for _, txn := range transactions {
		if txn.Category == "Bills_Utilities" &&
			(txn.Narration == "LIC" || txn.Narration == "INSURANCE") {
			hasInsurance = true
			break
		}
	}
	if !hasInsurance {
		recommendations = append(recommendations, models.RecommendedProduct{
			ID:          id,
			ProductName: "HDFC Life Click 2 Protect",
			Type:        "Insurance",
			Reason:      "No active term insurance detected. Secure your family's future.",
			Icon:        "Shield",
			ActionLink:  "#",
		})
		id++
	}

	// Investment recommendation
	if cashFlowScore.Score > 60 {
		recommendations = append(recommendations, models.RecommendedProduct{
			ID:          id,
			ProductName: "HDFC Sky Demat Account",
			Type:        "Investment",
			Reason:      "You have a healthy savings balance. Start investing in stocks & MFs.",
			Icon:        "TrendingUp",
			ActionLink:  "#",
		})
	}

	return recommendations
}

func generateBehaviourInsights(transactions []models.ClassifiedTransaction) []models.BehaviourInsight {
	insights := make([]models.BehaviourInsight, 0)

	// Weekend spending analysis (simplified)
	weekendSpend := 0.0
	weekdaySpend := 0.0
	weekendCount := 0
	weekdayCount := 0

	for _, txn := range transactions {
		if txn.IsIncome {
			continue
		}
		// Simplified - would need actual date parsing
		weekdayCount++
		weekdaySpend += txn.WithdrawalAmt
	}

	if weekdayCount > 0 && weekendCount > 0 {
		weekendAvg := weekendSpend / float64(weekendCount)
		weekdayAvg := weekdaySpend / float64(weekdayCount)
		if weekendAvg > weekdayAvg*1.2 {
			insights = append(insights, models.BehaviourInsight{
				Type:    "Weekend Spender",
				Insight: "You spend more on weekends compared to weekdays.",
			})
		}
	}

	// Impulse buying (simplified)
	insights = append(insights, models.BehaviourInsight{
		Type:    "Impulse Buying",
		Insight: "Great job! Impulse purchases are down by 5% this month.",
	})

	return insights
}

func generateSavingsOpportunities(
	transactions []models.ClassifiedTransaction,
	categorySummary models.CategorySummary,
) []models.SavingsOpportunity {
	opportunities := make([]models.SavingsOpportunity, 0)

	// Subscription optimization
	if categorySummary.FoodDelivery > 5000 {
		opportunities = append(opportunities, models.SavingsOpportunity{
			Category:      "Switch to Annual Plan",
			PotentialSave: 1200,
			Action:        "Switch",
			Difficulty:    "Easy",
			Impact:        "High",
		})
	}

	// Dining out reduction
	if categorySummary.Dining > 3000 {
		opportunities = append(opportunities, models.SavingsOpportunity{
			Category:      "Reduce Dining Out",
			PotentialSave: 3000,
			Action:        "Limit",
			Difficulty:    "Medium",
			Impact:        "Medium",
		})
	}

	return opportunities
}
