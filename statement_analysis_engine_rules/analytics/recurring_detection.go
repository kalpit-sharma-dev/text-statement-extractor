package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"classify/statement_analysis_engine_rules/utils"
	"math"
	"sort"
	"strings"
)

// RecurringPaymentDetector implements comprehensive recurring payment detection
type RecurringPaymentDetector struct {
	transactions []models.ClassifiedTransaction
}

// NewRecurringPaymentDetector creates a new detector
func NewRecurringPaymentDetector(transactions []models.ClassifiedTransaction) *RecurringPaymentDetector {
	return &RecurringPaymentDetector{
		transactions: transactions,
	}
}

// DetectRecurringPayments detects all recurring payments with confidence scoring
func (d *RecurringPaymentDetector) DetectRecurringPayments() []models.RecurringPayment {
	// Group transactions by counterparty signature
	groups := d.groupByCounterparty()

	result := make([]models.RecurringPayment, 0)

	for signature, txns := range groups {
		if len(txns) < 2 {
			continue
		}

		// Calculate confidence score
		confidence, frequency, firstSeen, lastSeen := d.calculateRecurringConfidence(txns, signature)

		// Threshold: ≥50 confidence = probable recurring, ≥70 = confirmed
		if confidence >= 50 {
			avgAmount, dayOfMonth := d.calculateAverages(txns)

			// Extract human-readable name from transactions
			// Don't use the signature hash directly - extract merchant/beneficiary/narration
			displayName := d.extractDisplayName(txns, signature)

			result = append(result, models.RecurringPayment{
				Name:       displayName,
				Amount:     avgAmount,
				DayOfMonth: dayOfMonth,
				Pattern:    frequency,
				Confidence: confidence,
				Frequency:  frequency,
				FirstSeen:  firstSeen,
				LastSeen:   lastSeen,
				Count:      len(txns),
			})
		}
	}

	// Sort by confidence (highest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Confidence > result[j].Confidence
	})

	return result
}

// groupByCounterparty groups transactions by counterparty signature
// Uses: merchant name, narration fingerprint, or beneficiary identifier
func (d *RecurringPaymentDetector) groupByCounterparty() map[string][]models.ClassifiedTransaction {
	groups := make(map[string][]models.ClassifiedTransaction)

	for _, txn := range d.transactions {
		// Only process expenses (withdrawals) or income (deposits) - not both
		if txn.WithdrawalAmt == 0 && txn.DepositAmt == 0 {
			continue
		}

		// Get counterparty signature
		signature := d.getCounterpartySignature(txn)

		if signature == "" {
			continue
		}

		groups[signature] = append(groups[signature], txn)
	}

	return groups
}

// getCounterpartySignature returns a stable identifier for the counterparty
func (d *RecurringPaymentDetector) getCounterpartySignature(txn models.ClassifiedTransaction) string {
	// Priority 1: Normalized merchant name (if available and meaningful)
	if txn.Merchant != "" && txn.Merchant != "Unknown" {
		merchantUpper := strings.ToUpper(strings.TrimSpace(txn.Merchant))
		// Exclude generic merchants
		if !isGenericMerchant(merchantUpper) {
			return "MERCHANT:" + merchantUpper
		}
	}

	// Priority 2: Narration fingerprint (most stable for recurring payments)
	fingerprint := utils.FingerprintNarration(txn.Narration)
	if fingerprint != "" {
		return "FINGERPRINT:" + fingerprint
	}

	// Priority 3: Beneficiary identifier
	if txn.Beneficiary != "" {
		beneficiaryUpper := strings.ToUpper(strings.TrimSpace(txn.Beneficiary))
		return "BENEFICIARY:" + beneficiaryUpper
	}

	return ""
}

// isGenericMerchant checks if merchant is too generic to use as signature
func isGenericMerchant(merchant string) bool {
	generic := []string{
		"UNKNOWN", "MERCHANT", "PAYMENT", "TRANSACTION",
		"BANK", "ATM", "POS", "UPI", "IMPS", "NEFT",
	}
	for _, g := range generic {
		if strings.Contains(merchant, g) {
			return true
		}
	}
	return false
}

// calculateRecurringConfidence calculates confidence score (0-100) for recurring payment
func (d *RecurringPaymentDetector) calculateRecurringConfidence(
	txns []models.ClassifiedTransaction,
	signature string,
) (confidence int, frequency string, firstSeen string, lastSeen string) {
	// Sort transactions by date
	sortedTxns := make([]models.ClassifiedTransaction, len(txns))
	copy(sortedTxns, txns)
	sort.Slice(sortedTxns, func(i, j int) bool {
		date1, _ := utils.ParseDate(sortedTxns[i].Date)
		date2, _ := utils.ParseDate(sortedTxns[j].Date)
		return date1.Before(date2)
	})

	// Get first and last seen dates
	if len(sortedTxns) > 0 {
		firstSeen = sortedTxns[0].Date
		lastSeen = sortedTxns[len(sortedTxns)-1].Date
	}

	// Signal 1: Same merchant/fingerprint (+30 points)
	score := 30

	// Signal 2: Repeated occurrence
	count := len(txns)
	if count >= 3 {
		score += 10 // Bonus for 3+ occurrences
	} else if count == 2 {
		// Check if strong signals present (can lower requirement to 2)
		hasStrongSignal := false
		for _, txn := range txns {
			if hasRecurringKeyword(txn.Narration) {
				hasStrongSignal = true
				break
			}
		}
		if hasStrongSignal {
			score += 5 // Allow 2 occurrences with strong signal
		} else {
			return 0, "", firstSeen, lastSeen // Need at least 2 with strong signal or 3+
		}
	} else {
		return 0, "", firstSeen, lastSeen // Need at least 2
	}

	// Signal 3: Time-based periodicity
	periodicityScore, detectedFrequency := d.checkPeriodicity(sortedTxns)
	score += periodicityScore
	frequency = detectedFrequency

	// Signal 4: Amount stability
	amountScore := d.checkAmountStability(sortedTxns)
	score += amountScore

	// Signal 5: Keyword match
	keywordScore := d.checkKeywords(sortedTxns)
	score += keywordScore

	// Signal 6: Day-of-month stability
	dayScore := d.checkDayOfMonthStability(sortedTxns)
	score += dayScore

	// Signal 7: Direction consistency
	if d.checkDirectionConsistency(sortedTxns) {
		score += 5
	}

	// Negative filters: Exclude certain types
	if d.shouldExclude(sortedTxns) {
		score = 0
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score, frequency, firstSeen, lastSeen
}

// checkPeriodicity checks time-based recurrence patterns
// Returns score (0-25) and detected frequency
func (d *RecurringPaymentDetector) checkPeriodicity(
	txns []models.ClassifiedTransaction,
) (score int, frequency string) {
	if len(txns) < 2 {
		return 0, ""
	}

	// Calculate gaps between transactions
	gaps := make([]float64, 0)
	for i := 1; i < len(txns); i++ {
		date1, err1 := utils.ParseDate(txns[i-1].Date)
		date2, err2 := utils.ParseDate(txns[i].Date)
		if err1 != nil || err2 != nil {
			continue
		}
		days := date2.Sub(date1).Hours() / 24
		gaps = append(gaps, days)
	}

	if len(gaps) == 0 {
		return 0, ""
	}

	// Calculate average gap
	avgGap := 0.0
	for _, gap := range gaps {
		avgGap += gap
	}
	avgGap /= float64(len(gaps))

	// Check against known patterns
	// Monthly: 28-35 days
	if avgGap >= 28 && avgGap <= 35 {
		// Check consistency
		consistent := true
		for _, gap := range gaps {
			if gap < 25 || gap > 38 {
				consistent = false
				break
			}
		}
		if consistent {
			return 25, "MONTHLY"
		}
		return 20, "MONTHLY" // Less consistent
	}

	// Quarterly: 85-95 days
	if avgGap >= 85 && avgGap <= 95 {
		consistent := true
		for _, gap := range gaps {
			if gap < 80 || gap > 100 {
				consistent = false
				break
			}
		}
		if consistent {
			return 25, "QUARTERLY"
		}
		return 20, "QUARTERLY"
	}

	// Weekly: 6-8 days
	if avgGap >= 6 && avgGap <= 8 {
		consistent := true
		for _, gap := range gaps {
			if gap < 5 || gap > 10 {
				consistent = false
				break
			}
		}
		if consistent {
			return 25, "WEEKLY"
		}
		return 20, "WEEKLY"
	}

	// If gaps are somewhat consistent but don't match standard patterns
	variance := 0.0
	for _, gap := range gaps {
		variance += math.Abs(gap - avgGap)
	}
	variance /= float64(len(gaps))

	if variance < 5 { // Low variance = consistent pattern
		return 15, "CUSTOM"
	}

	return 0, ""
}

// checkAmountStability checks if amounts are stable (within ±3-5%)
func (d *RecurringPaymentDetector) checkAmountStability(
	txns []models.ClassifiedTransaction,
) int {
	if len(txns) < 2 {
		return 0
	}

	amounts := make([]float64, 0)
	for _, txn := range txns {
		amt := txn.WithdrawalAmt
		if amt == 0 {
			amt = txn.DepositAmt
		}
		if amt > 0 {
			amounts = append(amounts, amt)
		}
	}

	if len(amounts) < 2 {
		return 0
	}

	// Calculate average
	avg := 0.0
	for _, amt := range amounts {
		avg += amt
	}
	avg /= float64(len(amounts))

	// Check variance
	maxVariance := 0.0
	for _, amt := range amounts {
		variance := math.Abs(amt-avg) / avg
		if variance > maxVariance {
			maxVariance = variance
		}
	}

	// ±3% = perfect (20 points), ±5% = good (15 points), ±10% = acceptable (10 points)
	if maxVariance <= 0.03 {
		return 20
	} else if maxVariance <= 0.05 {
		return 15
	} else if maxVariance <= 0.10 {
		return 10
	}

	return 0
}

// checkKeywords checks for recurring payment keywords
func (d *RecurringPaymentDetector) checkKeywords(
	txns []models.ClassifiedTransaction,
) int {
	keywordCount := 0

	for _, txn := range txns {
		if hasRecurringKeyword(txn.Narration) {
			keywordCount++
		}
	}

	// If all transactions have keywords, full score
	if keywordCount == len(txns) {
		return 15
	} else if keywordCount > 0 {
		// Partial keyword match
		return 10
	}

	return 0
}

// hasRecurringKeyword checks if narration contains recurring payment keywords
func hasRecurringKeyword(narration string) bool {
	narrationUpper := strings.ToUpper(narration)

	// High-confidence keywords
	highConfidenceKeywords := []string{
		"EMI", "LOAN", "REPAY", "INSTALLMENT", "INSTALMENT",
		"CC", "CARD", "CREDIT", "BILL", // Credit card
		"SALARY", "PAYROLL", // Salary
		"RENT", "LANDLORD", // Rent
		"NETFLIX", "PRIME", "SPOTIFY", "SUBSCRIPTION", // Subscriptions
		"PREMIUM", "INSURANCE", // Insurance
		"SIP", "NACH", "ECS", "AUTO DEBIT", "AUTODEBIT",
	}

	for _, keyword := range highConfidenceKeywords {
		if strings.Contains(narrationUpper, keyword) {
			return true
		}
	}

	return false
}

// checkDayOfMonthStability checks if payments occur on same day of month (±2 days)
func (d *RecurringPaymentDetector) checkDayOfMonthStability(
	txns []models.ClassifiedTransaction,
) int {
	if len(txns) < 2 {
		return 0
	}

	days := make([]int, 0)
	for _, txn := range txns {
		date, err := utils.ParseDate(txn.Date)
		if err != nil {
			continue
		}
		days = append(days, date.Day())
	}

	if len(days) < 2 {
		return 0
	}

	// Calculate average day
	avgDay := 0
	for _, day := range days {
		avgDay += day
	}
	avgDay /= len(days)

	// Check if all days are within ±2 days of average
	allWithinRange := true
	for _, day := range days {
		diff := int(math.Abs(float64(day - avgDay)))
		if diff > 2 {
			allWithinRange = false
			break
		}
	}

	if allWithinRange {
		return 10
	}

	return 0
}

// checkDirectionConsistency checks if all transactions are same direction
func (d *RecurringPaymentDetector) checkDirectionConsistency(
	txns []models.ClassifiedTransaction,
) bool {
	if len(txns) == 0 {
		return false
	}

	// Check first transaction direction
	firstIsDebit := txns[0].WithdrawalAmt > 0 && txns[0].DepositAmt == 0

	// All should be same direction
	for _, txn := range txns {
		isDebit := txn.WithdrawalAmt > 0 && txn.DepositAmt == 0
		if isDebit != firstIsDebit {
			return false
		}
	}

	return true
}

// shouldExclude checks if transaction group should be excluded from recurring
func (d *RecurringPaymentDetector) shouldExclude(
	txns []models.ClassifiedTransaction,
) bool {
	if len(txns) == 0 {
		return true
	}

	// Exclude one-off UPI merchant spends
	// Exclude high-frequency food delivery (unless it's a subscription)
	// Exclude P2P payments (unless salary)

	for _, txn := range txns {
		narrationUpper := strings.ToUpper(txn.Narration)

		// Exclude P2P unless it's salary
		if txn.Method == "UPI" || txn.Method == "IMPS" {
			if !strings.Contains(narrationUpper, "SALARY") &&
				!strings.Contains(narrationUpper, "PAYROLL") {
				// Check if it looks like P2P (person name, not merchant)
				if utils.IsPersonToPersonTransfer(txn.Narration, txn.Merchant, txn.WithdrawalAmt) {
					return true
				}
			}
		}

		// Exclude food delivery unless subscription
		if strings.Contains(narrationUpper, "SWIGGY") ||
			strings.Contains(narrationUpper, "ZOMATO") ||
			strings.Contains(narrationUpper, "UBER EATS") {
			// Only exclude if not a subscription pattern
			if !strings.Contains(narrationUpper, "SUBSCRIPTION") &&
				!strings.Contains(narrationUpper, "PRO") {
				// Check if amounts vary significantly (not subscription)
				if d.hasHighAmountVariance(txns) {
					return true
				}
			}
		}
	}

	return false
}

// hasHighAmountVariance checks if amounts vary significantly
func (d *RecurringPaymentDetector) hasHighAmountVariance(
	txns []models.ClassifiedTransaction,
) bool {
	if len(txns) < 2 {
		return false
	}

	amounts := make([]float64, 0)
	for _, txn := range txns {
		amt := txn.WithdrawalAmt
		if amt == 0 {
			amt = txn.DepositAmt
		}
		if amt > 0 {
			amounts = append(amounts, amt)
		}
	}

	if len(amounts) < 2 {
		return false
	}

	avg := 0.0
	for _, amt := range amounts {
		avg += amt
	}
	avg /= float64(len(amounts))

	// Check if variance > 30%
	for _, amt := range amounts {
		variance := math.Abs(amt-avg) / avg
		if variance > 0.30 {
			return true
		}
	}

	return false
}

// calculateAverages calculates average amount and day of month
func (d *RecurringPaymentDetector) calculateAverages(
	txns []models.ClassifiedTransaction,
) (avgAmount float64, avgDay int) {
	if len(txns) == 0 {
		return 0, 0
	}

	// Calculate average amount
	totalAmount := 0.0
	count := 0
	for _, txn := range txns {
		amt := txn.WithdrawalAmt
		if amt == 0 {
			amt = txn.DepositAmt
		}
		if amt > 0 {
			totalAmount += amt
			count++
		}
	}
	if count > 0 {
		avgAmount = totalAmount / float64(count)
	}

	// Calculate average day of month
	totalDay := 0
	dayCount := 0
	for _, txn := range txns {
		date, err := utils.ParseDate(txn.Date)
		if err != nil {
			continue
		}
		totalDay += date.Day()
		dayCount++
	}
	if dayCount > 0 {
		avgDay = totalDay / dayCount
	}

	return avgAmount, avgDay
}

// extractDisplayName extracts a human-readable name from transactions
// This replaces signature hashes with meaningful names
func (d *RecurringPaymentDetector) extractDisplayName(
	txns []models.ClassifiedTransaction,
	signature string,
) string {
	if len(txns) == 0 {
		return signature
	}

	// Priority 1: Use merchant name if available and meaningful
	for _, txn := range txns {
		if txn.Merchant != "" && txn.Merchant != "Unknown" {
			merchantUpper := strings.ToUpper(strings.TrimSpace(txn.Merchant))
			if !isGenericMerchant(merchantUpper) {
				return txn.Merchant
			}
		}
	}

	// Priority 2: Use beneficiary name if available
	for _, txn := range txns {
		if txn.Beneficiary != "" {
			return txn.Beneficiary
		}
	}

	// Priority 3: Extract name from normalized narration (without dates/IDs)
	// Use the first transaction's normalized narration as display name
	if len(txns) > 0 {
		normalizedNarration := utils.NormalizeNarrationForFingerprint(txns[0].Narration)
		if normalizedNarration != "" {
			// Clean up the normalized narration to make it more readable
			normalizedNarration = strings.TrimSpace(normalizedNarration)

			// If it's still too long, truncate intelligently
			if len(normalizedNarration) > 60 {
				// Try to find a meaningful break point
				words := strings.Fields(normalizedNarration)
				if len(words) > 0 {
					// Take first few meaningful words
					displayName := ""
					for i, word := range words {
						if i >= 5 { // Max 5 words
							break
						}
						if displayName != "" {
							displayName += " "
						}
						displayName += word
					}
					return displayName
				}
			}
			return normalizedNarration
		}
	}

	// Priority 4: Extract from original narration (first 50 chars)
	if len(txns) > 0 && txns[0].Narration != "" {
		narration := strings.TrimSpace(txns[0].Narration)
		if len(narration) > 50 {
			narration = narration[:50] + "..."
		}
		return narration
	}

	// Fallback: Remove prefix from signature
	if strings.HasPrefix(signature, "MERCHANT:") {
		return strings.TrimPrefix(signature, "MERCHANT:")
	}
	if strings.HasPrefix(signature, "BENEFICIARY:") {
		return strings.TrimPrefix(signature, "BENEFICIARY:")
	}
	if strings.HasPrefix(signature, "FINGERPRINT:") {
		// For fingerprint, we can't reverse it, so use a generic name
		// Try to extract category or method from transactions
		if len(txns) > 0 {
			if txns[0].Category != "" && txns[0].Category != "Other" {
				return txns[0].Category + " Payment"
			}
			if txns[0].Method != "" && txns[0].Method != "Other" {
				return txns[0].Method + " Payment"
			}
		}
		return "Recurring Payment"
	}

	return signature
}

// MatchTransactionToRecurring matches a transaction to recurring payments using pre-computed lookup map
// This is optimized for performance - avoids re-detecting recurring payments for each transaction
func MatchTransactionToRecurring(
	txn models.ClassifiedTransaction,
	detector *RecurringPaymentDetector,
	recurringMap map[string]models.RecurringPayment,
) models.RecurringMetadata {
	// Get counterparty signature for this transaction
	signature := detector.getCounterpartySignature(txn)
	if signature == "" {
		return models.RecurringMetadata{IsRecurring: false}
	}

	// Extract signature key (remove prefix like "MERCHANT:", "FINGERPRINT:", etc.)
	signatureKey := signature
	if idx := strings.Index(signature, ":"); idx > 0 {
		signatureKey = signature[idx+1:]
	}

	// Try direct match first
	if rp, found := recurringMap[signatureKey]; found {
		if rp.Confidence >= 50 {
			return models.RecurringMetadata{
				IsRecurring: true,
				Confidence:  rp.Confidence,
				Frequency:   rp.Frequency,
				FirstSeen:   rp.FirstSeen,
				LastSeen:    rp.LastSeen,
				Count:       rp.Count,
				Pattern:     rp.Pattern,
			}
		}
	}

	// Fallback: Check if merchant/beneficiary matches any recurring payment
	// This handles cases where signature format differs
	txnMerchantUpper := strings.ToUpper(txn.Merchant)
	txnBeneficiaryUpper := strings.ToUpper(txn.Beneficiary)
	txnFingerprint := utils.FingerprintNarration(txn.Narration)

	for _, rp := range recurringMap {
		if rp.Confidence >= 50 {
			rpNameUpper := strings.ToUpper(rp.Name)
			rpFingerprint := utils.FingerprintNarration(rp.Name)

			// Check if merchant/beneficiary matches
			if (txnMerchantUpper != "" && strings.Contains(rpNameUpper, txnMerchantUpper)) ||
				(txnBeneficiaryUpper != "" && strings.Contains(rpNameUpper, txnBeneficiaryUpper)) ||
				(txnFingerprint != "" && rpFingerprint != "" && txnFingerprint == rpFingerprint) {
				return models.RecurringMetadata{
					IsRecurring: true,
					Confidence:  rp.Confidence,
					Frequency:   rp.Frequency,
					FirstSeen:   rp.FirstSeen,
					LastSeen:    rp.LastSeen,
					Count:       rp.Count,
					Pattern:     rp.Pattern,
				}
			}
		}
	}

	return models.RecurringMetadata{IsRecurring: false}
}

// DetectRecurringForTransaction detects if a specific transaction is recurring
// DEPRECATED: Use MatchTransactionToRecurring with pre-computed recurring payments for better performance
// This function is kept for backward compatibility but should not be used in loops
func DetectRecurringForTransaction(
	txn models.ClassifiedTransaction,
	allTransactions []models.ClassifiedTransaction,
) models.RecurringMetadata {
	detector := NewRecurringPaymentDetector(allTransactions)
	recurringPayments := detector.DetectRecurringPayments()

	// Build lookup map
	recurringMap := make(map[string]models.RecurringPayment)
	for _, rp := range recurringPayments {
		recurringMap[rp.Name] = rp
	}

	return MatchTransactionToRecurring(txn, detector, recurringMap)
}
