package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"math"
	"strings"
)

// CalculateTaxInsights calculates tax-related insights
func CalculateTaxInsights(transactions []models.ClassifiedTransaction) models.TaxInsights {
	potentialSave := 0.0
	missedDeductions := make([]string, 0)

	// Check for investment transactions
	hasELSS := false
	hasInsurance := false
	hasNPS := false
	hasPPF := false

	totalInvestment := 0.0
	totalInsurance := 0.0

	for _, txn := range transactions {
		narration := txn.Narration
		// Investments are typically withdrawals (money going out for investment)
		// But could also be deposits in some cases (returns, dividends)
		// For tax purposes, we care about money invested (withdrawals)
		amount := txn.WithdrawalAmt
		
		// Skip if no withdrawal amount (not an investment transaction)
		if amount == 0 {
			continue
		}

		// Check for ELSS
		if contains(narration, "ELSS") || contains(narration, "MUTUAL FUND") {
			hasELSS = true
			totalInvestment += amount
		}

		// Check for NPS
		if contains(narration, "NPS") || contains(narration, "NATIONAL PENSION") {
			hasNPS = true
			totalInvestment += amount
		}

		// Check for PPF
		if contains(narration, "PPF") || contains(narration, "PUBLIC PROVIDENT") {
			hasPPF = true
			totalInvestment += amount
		}

		// Check for insurance
		if contains(narration, "INSURANCE") || contains(narration, "PREMIUM") ||
			contains(narration, "LIC") || contains(narration, "HDFC LIFE") ||
			contains(narration, "MAXLIFE") || contains(narration, "SBI LIFE") {
			hasInsurance = true
			totalInsurance += amount
		}
	}

	// Calculate potential savings
	// Section 80C limit: 1.5L
	// Section 80D limit: 25k (health insurance)
	section80CUsed := math.Min(totalInvestment, 150000)
	section80DUsed := math.Min(totalInsurance, 25000)

	section80CAvailable := 150000 - section80CUsed
	section80DAvailable := 25000 - section80DUsed

	if section80CAvailable > 0 {
		potentialSave += (section80CAvailable * 0.30) // 30% tax bracket assumption
		missedDeductions = append(missedDeductions, "Invest in ELSS (Section 80C)")
	}

	if section80DAvailable > 0 {
		potentialSave += (section80DAvailable * 0.30)
		missedDeductions = append(missedDeductions, "Medical Insurance (Section 80D)")
	}

	if !hasELSS && !hasNPS && !hasPPF {
		missedDeductions = append(missedDeductions, "Start SIP in ELSS funds")
	}

	if !hasInsurance {
		missedDeductions = append(missedDeductions, "Consider term insurance")
	}

	return models.TaxInsights{
		PotentialSave:    potentialSave,
		MissedDeductions: missedDeductions,
	}
}

func contains(str, substr string) bool {
	return strings.Contains(strings.ToUpper(str), strings.ToUpper(substr))
}
