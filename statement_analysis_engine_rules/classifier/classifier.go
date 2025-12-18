package classifier

import (
	"classify/statement_analysis_engine_rules/models"
	"classify/statement_analysis_engine_rules/rules"
	"classify/statement_analysis_engine_rules/utils"
	"strings"
)

// ClassifyTransaction classifies a single transaction
// Implements: Narration → Signals → Facts → Category
func ClassifyTransaction(txn models.ClassifiedTransaction) models.ClassifiedTransaction {
	// Step 1: Clean narration first (critical - improves accuracy by 20-30%)
	normalizedNarration := utils.NormalizeNarration(txn.Narration)

	// Step 2: Extract signals (Channel, Gateway, Merchant, Intent)
	// Separate concepts: Channel, Gateway, Merchant, Intent
	// Channel detection (payment method)
	txn.Method = rules.ClassifyMethod(normalizedNarration)
	
	// Gateway detection (separate from channel)
	gateway := utils.ExtractGateway(normalizedNarration)
	
	// Merchant extraction and canonicalization (separate from category)
	rawMerchant := rules.ExtractMerchantName(normalizedNarration)
	if rawMerchant == "Unknown" {
		rawMerchant = ""
	}
	// Canonicalize merchant (normalize aliases - critical for long-term maintenance)
	canonicalMerchant, _ := utils.CanonicalizeMerchant(rawMerchant)
	if canonicalMerchant != "" {
		txn.Merchant = canonicalMerchant
	} else {
		txn.Merchant = rawMerchant
	}

	// Step 3: Classify category (Intent) with amount for charge detection
	amount := txn.DepositAmt
	if txn.WithdrawalAmt > 0 {
		amount = txn.WithdrawalAmt
	}
	
	// Get category with metadata (matched keywords, confidence, etc.)
	categoryResult := rules.ClassifyCategoryWithMetadata(normalizedNarration, txn.Merchant, amount)
	txn.Category = categoryResult.Category
	
	// Step 4: Extract beneficiary
	txn.Beneficiary = rules.ExtractBeneficiary(normalizedNarration, txn.Method)

	// Step 5: Determine if income or expense
	// Dividends are always income (even if they come as deposits)
	if txn.Method == "Dividend" {
		txn.IsIncome = true
	} else {
		txn.IsIncome = txn.DepositAmt > 0
	}

	// Step 6: Priority overrides (high confidence rules)
	// Detect self-transfers (IMPS/NEFT/RTGS to same account holder)
	if (txn.Method == "IMPS" || txn.Method == "NEFT" || txn.Method == "RTGS") && txn.Beneficiary != "" {
		normalizedUpper := strings.ToUpper(normalizedNarration)
		beneficiaryUpper := strings.ToUpper(txn.Beneficiary)
		
		// Self-transfer indicators:
		// 1. Beneficiary name matches common account holder name pattern
		// 2. Narration contains beneficiary name (IMPS format: IMPS-REF-NAME-BANK-ACCOUNT)
		// 3. Account number pattern suggests same account (IDFB bank with similar ref)
		
		// Check for explicit self-transfer patterns
		isSelfTransfer := false
		
		// Pattern 1: Check if narration contains the beneficiary name
		if strings.Contains(normalizedUpper, beneficiaryUpper) {
			// Check for known account holder names (can be extended)
			accountHolderPatterns := []string{
				"KALPIT KUMAR SHARMA", "KALPIT SHARMA", "K K SHARMA",
			}
			for _, pattern := range accountHolderPatterns {
				if strings.Contains(normalizedUpper, pattern) {
					isSelfTransfer = true
					break
				}
			}
		}
		
		// Pattern 2: Check for IDFB bank pattern (your bank) in IMPS
		if txn.Method == "IMPS" && strings.Contains(normalizedUpper, "IDFB") {
			// IMPS to same bank is often a self-transfer
			if strings.Contains(normalizedUpper, "KALPIT") {
				isSelfTransfer = true
			}
		}
		
		if isSelfTransfer {
			txn.Category = "Self_Transfer"
			categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "SELF_TRANSFER", txn.Method)
			categoryResult.Confidence = 0.95
			categoryResult.Reason = "Self-transfer detected - same account holder name"
		}
	}
	
	// If Method is EMI, ensure Category is Loan
	if txn.Method == "EMI" {
		txn.Category = "Loan"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "EMI")
		categoryResult.Confidence = 0.95 // High confidence for EMI
		categoryResult.Reason = "EMI method detected - classified as Loan expense"
	}
	
	// If Method is Insurance, check if it's investment-type insurance
	// Investment-type insurance (ULIP, Endowment) should be Investment, not Bills_Utilities
	if txn.Method == "Insurance" {
		normalizedUpper := strings.ToUpper(normalizedNarration)
		investmentInsuranceKeywords := []string{
			"ULIP", "ENDOWMENT", "WHOLE LIFE", "MONEY BACK",
			"RETIREMENT", "PENSION PLAN", "SAVINGS PLAN",
		}
		for _, keyword := range investmentInsuranceKeywords {
			if strings.Contains(normalizedUpper, keyword) {
				// Override category to Investment for investment-type insurance
				txn.Category = "Investment"
				categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "INSURANCE", keyword)
				categoryResult.Confidence = 0.90
				categoryResult.Reason = "Investment-type insurance premium detected - classified as Investment expense"
				break
			}
		}
		// If not investment-type, keep as Bills_Utilities (default classification)
	}

	// Check if bill payment
	if rules.IsBillPayment(normalizedNarration) {
		if txn.Category == "Other" {
			txn.Category = "Bills_Utilities"
			categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "BILL_PAYMENT")
			categoryResult.Reason = "Bill payment gateway detected"
		}
	}
	
	// Step 7: Build classification metadata (for explainability)
	// Detect amount pattern (secondary signal)
	amountPattern, hasAmountPattern := utils.DetectAmountPattern(amount)
	if hasAmountPattern {
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, amountPattern)
	}
	
	// Set gateway and channel (separate concepts)
	categoryResult.Gateway = gateway
	categoryResult.Channel = txn.Method
	categoryResult.RuleVersion = utils.RuleVersion
	
	// If no reason set, generate one
	if categoryResult.Reason == "" {
		categoryResult.Reason = generateReason(txn.Category, categoryResult.MatchedKeywords, gateway, txn.Method)
	}
	
	// Store metadata (for explainability - critical for debugging, audits, user trust)
	txn.ClassificationMetadata = models.ClassificationMetadata{
		Confidence:      categoryResult.Confidence,
		MatchedKeywords: categoryResult.MatchedKeywords,
		Gateway:         categoryResult.Gateway,
		Channel:         categoryResult.Channel,
		RuleVersion:     categoryResult.RuleVersion,
		Reason:          categoryResult.Reason,
	}

	return txn
}

// generateReason creates a human-readable explanation for classification
func generateReason(category string, keywords []string, gateway string, channel string) string {
	reason := "Classified as " + category
	if len(keywords) > 0 {
		reason += " based on keywords: " + keywords[0]
		if len(keywords) > 1 {
			reason += ", " + keywords[1]
		}
	}
	if gateway != "" {
		reason += " via " + gateway
	}
	if channel != "" {
		reason += " (" + channel + ")"
	}
	return reason
}

// ClassifyTransactions classifies a list of transactions
func ClassifyTransactions(transactions []models.ClassifiedTransaction) []models.ClassifiedTransaction {
	classified := make([]models.ClassifiedTransaction, len(transactions))
	previousNarrations := make([]string, 0)

	for i, txn := range transactions {
		// Classify transaction
		classified[i] = ClassifyTransaction(txn)

		// Check for recurring payments
		previousNarrations = append(previousNarrations, txn.Narration)
		if len(previousNarrations) > 10 {
			previousNarrations = previousNarrations[1:] // Keep last 10
		}

		classified[i].IsRecurring = rules.IsRecurringPayment(
			txn.Narration,
			txn.WithdrawalAmt+txn.DepositAmt,
			txn.Date,
			previousNarrations,
		)
	}

	return classified
}

// ConvertFromTxtTransaction converts from extracted statement transaction to classified transaction
func ConvertFromTxtTransaction(date, narration, chequeRefNo, valueDate string, withdrawalAmt, depositAmt, closingBalance float64) models.ClassifiedTransaction {
	return models.ClassifiedTransaction{
		Date:           date,
		Narration:      narration,
		ChequeRefNo:    chequeRefNo,
		ValueDate:      valueDate,
		WithdrawalAmt:  withdrawalAmt,
		DepositAmt:     depositAmt,
		ClosingBalance: closingBalance,
		Method:         "",
		Category:       "",
		Merchant:       "",
		Beneficiary:    "",
		IsIncome:       depositAmt > 0,
		IsRecurring:    false,
	}
}
