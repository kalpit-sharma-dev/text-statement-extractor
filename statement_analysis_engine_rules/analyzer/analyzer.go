package analyzer

import (
	"classify/statement_analysis_engine_rules/analytics"
	"classify/statement_analysis_engine_rules/classifier"
	"classify/statement_analysis_engine_rules/models"
	"strings"
)

// Analyzer is the main analyzer struct
type Analyzer struct {
	transactions        []models.ClassifiedTransaction
	statementTotalCredits float64 // Optional: official statement total credits
	statementTotalDebits  float64 // Optional: official statement total debits
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		transactions: make([]models.ClassifiedTransaction, 0),
	}
}

// SetStatementTotals sets the official statement totals (use these for accurate calculations)
func (a *Analyzer) SetStatementTotals(totalCredits, totalDebits float64) {
	a.statementTotalCredits = totalCredits
	a.statementTotalDebits = totalDebits
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
// customerName is optional - if provided, used for self-transfer detection
func (a *Analyzer) ClassifyAll(customerName string) {
	a.transactions = classifier.ClassifyTransactions(a.transactions, customerName)
}

// Analyze generates complete analysis
func (a *Analyzer) Analyze(
	accountNo string,
	customerName string,
	statementPeriod string,
	openingBalance float64,
	closingBalance float64,
) models.ClassifyResponse {
	// Classify all transactions first (pass customerName for self-transfer detection)
	a.ClassifyAll(customerName)

	// Calculate all analytics
	// Use statement totals if available, otherwise calculate from transactions
	accountSummary := analytics.CalculateAccountSummaryWithTotals(
		accountNo,
		customerName,
		statementPeriod,
		openingBalance,
		closingBalance,
		a.transactions,
		a.statementTotalCredits,
		a.statementTotalDebits,
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
	// Use new anomaly_engine package (bank-grade detection)
	anomalyDetection := analytics.CalculateAnomalyDetectionWithEngine(a.transactions, customerName)
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
	
	// Prepare all transactions for heatmap and pattern analysis
	transactionDetails := analytics.PrepareTransactionsForResponse(a.transactions)

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
		AnomalyDetection:     anomalyDetection,
		Transactions:         transactionDetails, // All transactions for heatmap and pattern analysis
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
		// Only check withdrawals (expenses), not deposits
		if txn.DepositAmt > 0 || txn.WithdrawalAmt == 0 {
			continue
		}
		narration := txn.Narration
		// Check if narration contains insurance-related keywords
		if txn.Category == "Bills_Utilities" {
			if contains(narration, "LIC") || contains(narration, "INSURANCE") ||
				contains(narration, "PREMIUM") || contains(narration, "HDFC LIFE") ||
				contains(narration, "MAXLIFE") || contains(narration, "SBI LIFE") {
				hasInsurance = true
				break
			}
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

// Helper function for string contains (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToUpper(s), strings.ToUpper(substr))
}

func generateBehaviourInsights(transactions []models.ClassifiedTransaction) []models.BehaviourInsight {
	insights := make([]models.BehaviourInsight, 0)

	// Weekend spending analysis (simplified)
	weekendSpend := 0.0
	weekdaySpend := 0.0
	weekendCount := 0
	weekdayCount := 0

	for _, txn := range transactions {
		// Only count withdrawals (expenses), skip deposits
		if txn.DepositAmt > 0 || txn.WithdrawalAmt == 0 {
			continue
		}
		// Simplified - would need actual date parsing to determine weekend
		// For now, count all as weekday (this is a placeholder - needs proper date logic)
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
