package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// =============================================================================
// PRODUCTION-READY ANOMALY DETECTION
// =============================================================================
//
// This is a STATISTICAL-BASED approach (not ML model-based) that:
// 1. Requires NO training data
// 2. Works immediately with any transaction history
// 3. Uses proven statistical methods (Z-Score, IQR, Percentiles)
// 4. Is fast, explainable, and production-ready
//
// Why NOT ML Model?
// - Anomaly detection doesn't need training - statistical methods work better
// - No labeled data required
// - Faster execution
// - More explainable for compliance/audit
// - Go can do this perfectly - no Python needed
// =============================================================================

// Note: AnomalyDetection, AnomalyDetail, and AnomalySummary types are defined in models package

// CategoryProfile holds statistical profile for a category
type CategoryProfile struct {
	Mean   float64   `json:"mean"`
	Median float64   `json:"median"`
	StdDev float64   `json:"stdDev"`
	Min    float64   `json:"min"`
	Max    float64   `json:"max"`
	Q1     float64   `json:"q1"`  // 25th percentile
	Q3     float64   `json:"q3"`  // 75th percentile
	IQR    float64   `json:"iqr"` // Interquartile range
	P95    float64   `json:"p95"` // 95th percentile
	P99    float64   `json:"p99"` // 99th percentile
	Count  int       `json:"count"`
	Values []float64 `json:"-"` // Internal: for calculations
}

// SpendingProfile represents user's spending patterns
type SpendingProfile struct {
	CategoryProfiles  map[string]*CategoryProfile `json:"categoryProfiles"`
	MerchantFrequency map[string]int              `json:"merchantFrequency"`
	AvgDailySpend     float64                     `json:"avgDailySpend"`
	AvgWeeklySpend    float64                     `json:"avgWeeklySpend"`
	TotalTransactions int                         `json:"totalTransactions"`
	TransactionDays   int                         `json:"transactionDays"`
}

// =============================================================================
// MAIN ENTRY POINT
// =============================================================================

// CalculateAnomalyDetection performs production-ready anomaly detection
// This is the function you call from analyzer.go
func CalculateAnomalyDetection(transactions []models.ClassifiedTransaction) models.AnomalyDetection {
	result := models.AnomalyDetection{
		Anomalies:    make([]models.AnomalyDetail, 0),
		TotalChecked: 0,
		Summary: models.AnomalySummary{
			ByType:     make(map[string]int),
			BySeverity: make(map[string]int),
		},
	}

	if len(transactions) < 10 {
		// Not enough data for meaningful detection
		result.RiskScore = 0
		return result
	}

	// Step 1: Build spending profile from transaction history
	profile := buildSpendingProfile(transactions)

	// Step 2: Detect anomalies
	for i, txn := range transactions {
		// Only check expenses (withdrawals)
		if txn.WithdrawalAmt == 0 {
			continue
		}

		result.TotalChecked++

		// Check 1: Unusual Amount (Z-Score + IQR)
		if anomaly := detectUnusualAmountAnomaly(txn, i, profile); anomaly != nil {
			result.Anomalies = append(result.Anomalies, *anomaly)
		}

		// Check 2: Unusual Merchant (First-time or rare)
		if anomaly := detectUnusualMerchantAnomaly(txn, i, profile); anomaly != nil {
			result.Anomalies = append(result.Anomalies, *anomaly)
		}

		// Check 3: Duplicate Payment
		if anomaly := detectDuplicatePaymentAnomaly(txn, i, transactions); anomaly != nil {
			result.Anomalies = append(result.Anomalies, *anomaly)
		}

		// Check 4: Round Amount Pattern (fraud indicator)
		if anomaly := detectRoundAmountAnomaly(txn, i); anomaly != nil {
			result.Anomalies = append(result.Anomalies, *anomaly)
		}

		// Check 5: Spending Spike (sudden increase)
		if anomaly := detectSpendingSpikeAnomaly(txn, i, transactions, profile); anomaly != nil {
			result.Anomalies = append(result.Anomalies, *anomaly)
		}
	}

	// Step 3: Calculate summary
	result.AnomalyCount = len(result.Anomalies)
	result.Summary = buildAnomalySummary(result.Anomalies)
	result.RiskScore = calculateRiskScore(result.Anomalies, result.TotalChecked)

	return result
}

// =============================================================================
// PROFILE BUILDING
// =============================================================================

// buildSpendingProfile creates statistical profile from transaction history
func buildSpendingProfile(transactions []models.ClassifiedTransaction) *SpendingProfile {
	profile := &SpendingProfile{
		CategoryProfiles:  make(map[string]*CategoryProfile),
		MerchantFrequency: make(map[string]int),
	}

	// Collect amounts by category
	categoryAmounts := make(map[string][]float64)
	var totalSpend float64
	var expenseCount int

	for _, txn := range transactions {
		if txn.WithdrawalAmt > 0 {
			category := txn.Category
			if category == "" {
				category = "Other"
			}
			categoryAmounts[category] = append(categoryAmounts[category], txn.WithdrawalAmt)
			totalSpend += txn.WithdrawalAmt
			expenseCount++

			// Track merchant frequency
			merchant := strings.ToUpper(strings.TrimSpace(txn.Merchant))
			if merchant != "" && merchant != "UNKNOWN" {
				profile.MerchantFrequency[merchant]++
			}
		}
	}

	// Calculate statistics for each category
	for category, amounts := range categoryAmounts {
		profile.CategoryProfiles[category] = calculateCategoryProfile(amounts)
	}

	// Calculate time-based averages
	days := calculateTransactionDays(transactions)
	if days > 0 {
		profile.TransactionDays = days
		profile.AvgDailySpend = totalSpend / float64(days)
		profile.AvgWeeklySpend = profile.AvgDailySpend * 7
	}
	profile.TotalTransactions = expenseCount

	return profile
}

// calculateCategoryProfile calculates comprehensive statistics for a category
func calculateCategoryProfile(values []float64) *CategoryProfile {
	if len(values) == 0 {
		return &CategoryProfile{}
	}

	profile := &CategoryProfile{
		Count:  len(values),
		Values: values,
	}

	// Sort for percentile calculations
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	// Min, Max
	profile.Min = sorted[0]
	profile.Max = sorted[len(sorted)-1]

	// Mean
	var sum float64
	for _, v := range values {
		sum += v
	}
	profile.Mean = sum / float64(len(values))

	// Standard Deviation
	var sumSquares float64
	for _, v := range values {
		diff := v - profile.Mean
		sumSquares += diff * diff
	}
	profile.StdDev = math.Sqrt(sumSquares / float64(len(values)))

	// Median (Q2)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		profile.Median = (sorted[mid-1] + sorted[mid]) / 2
	} else {
		profile.Median = sorted[mid]
	}

	// Q1 (25th percentile)
	q1Idx := len(sorted) / 4
	if q1Idx >= len(sorted) {
		q1Idx = len(sorted) - 1
	}
	profile.Q1 = sorted[q1Idx]

	// Q3 (75th percentile)
	q3Idx := (len(sorted) * 3) / 4
	if q3Idx >= len(sorted) {
		q3Idx = len(sorted) - 1
	}
	profile.Q3 = sorted[q3Idx]

	// IQR
	profile.IQR = profile.Q3 - profile.Q1

	// P95 (95th percentile)
	p95Idx := int(float64(len(sorted)) * 0.95)
	if p95Idx >= len(sorted) {
		p95Idx = len(sorted) - 1
	}
	profile.P95 = sorted[p95Idx]

	// P99 (99th percentile)
	p99Idx := int(float64(len(sorted)) * 0.99)
	if p99Idx >= len(sorted) {
		p99Idx = len(sorted) - 1
	}
	profile.P99 = sorted[p99Idx]

	return profile
}

// =============================================================================
// ANOMALY DETECTION METHODS
// =============================================================================

// detectUnusualAmountAnomaly uses Z-Score and IQR to detect unusual amounts
func detectUnusualAmountAnomaly(txn models.ClassifiedTransaction, idx int, profile *SpendingProfile) *models.AnomalyDetail {
	category := txn.Category
	if category == "" {
		category = "Other"
	}

	catProfile, exists := profile.CategoryProfiles[category]
	if !exists || catProfile.Count < 5 {
		return nil // Not enough data
	}

	amount := txn.WithdrawalAmt

	// Method 1: Z-Score
	var zScore float64
	if catProfile.StdDev > 0 {
		zScore = (amount - catProfile.Mean) / catProfile.StdDev
	}

	// Method 2: IQR-based (Tukey's method)
	upperBound := catProfile.Q3 + (1.5 * catProfile.IQR)
	lowerBound := catProfile.Q1 - (1.5 * catProfile.IQR)
	isIQROutlier := amount > upperBound || amount < lowerBound

	// Method 3: Percentile-based
	isP99Outlier := amount > catProfile.P99
	isP95Outlier := amount > catProfile.P95

	// Determine severity and score
	var severity string
	var score float64
	var reason string

	absZScore := math.Abs(zScore)
	if isP99Outlier || absZScore > 4 || amount > upperBound*2 {
		severity = "critical"
		score = 0.95
		reason = "Amount exceeds 99th percentile for this category"
	} else if isP95Outlier || absZScore > 3 || amount > upperBound*1.5 {
		severity = "high"
		score = 0.80
		reason = "Amount exceeds 95th percentile for this category"
	} else if isIQROutlier || absZScore > 2.5 {
		severity = "medium"
		score = 0.60
		reason = "Amount is significantly above typical range"
	} else if absZScore > 2 {
		severity = "low"
		score = 0.40
		reason = "Amount is above average for this category"
	} else {
		return nil // Not anomalous
	}

	return &models.AnomalyDetail{
		TransactionID:    idx,
		Type:             "unusual_amount",
		Severity:         severity,
		Score:            score,
		Description:      reason,
		Amount:           amount,
		Merchant:         txn.Merchant,
		Category:         category,
		Date:             txn.Date,
		Reason:           reason,
		StatisticalValue: zScore,
	}
}

// detectUnusualMerchantAnomaly detects first-time or rare merchants
func detectUnusualMerchantAnomaly(txn models.ClassifiedTransaction, idx int, profile *SpendingProfile) *models.AnomalyDetail {
	merchant := strings.ToUpper(strings.TrimSpace(txn.Merchant))
	if merchant == "" || merchant == "UNKNOWN" {
		return nil
	}

	frequency, exists := profile.MerchantFrequency[merchant]
	amount := txn.WithdrawalAmt

	// First-time merchant with large amount
	if !exists && amount > profile.AvgDailySpend*3 {
		return &models.AnomalyDetail{
			TransactionID:    idx,
			Type:             "unusual_merchant",
			Severity:         "medium",
			Score:            0.65,
			Description:      "First-time merchant with unusually large transaction",
			Amount:           amount,
			Merchant:         txn.Merchant,
			Category:         txn.Category,
			Date:             txn.Date,
			Reason:           "Merchant never seen before with amount 3x daily average",
			StatisticalValue: 0,
		}
	}

	// Very rare merchant (used only once) with large amount
	if frequency == 1 && amount > profile.AvgDailySpend*2 {
		return &models.AnomalyDetail{
			TransactionID:    idx,
			Type:             "unusual_merchant",
			Severity:         "low",
			Score:            0.45,
			Description:      "Rare merchant with above-average transaction",
			Amount:           amount,
			Merchant:         txn.Merchant,
			Category:         txn.Category,
			Date:             txn.Date,
			Reason:           "Merchant used only once before with amount 2x daily average",
			StatisticalValue: float64(frequency),
		}
	}

	return nil
}

// detectDuplicatePaymentAnomaly detects potential duplicate payments
func detectDuplicatePaymentAnomaly(txn models.ClassifiedTransaction, idx int, allTxns []models.ClassifiedTransaction) *models.AnomalyDetail {
	// Look at recent transactions (last 20 transactions or 7 days)
	lookbackLimit := 20
	if idx < lookbackLimit {
		lookbackLimit = idx
	}

	for i := idx - 1; i >= 0 && i >= idx-lookbackLimit; i-- {
		other := allTxns[i]

		// Same amount, same merchant
		if other.WithdrawalAmt == txn.WithdrawalAmt &&
			strings.EqualFold(strings.TrimSpace(other.Merchant), strings.TrimSpace(txn.Merchant)) &&
			txn.Merchant != "" {

			// Parse dates
			date1, err1 := parseTransactionDate(txn.Date)
			date2, err2 := parseTransactionDate(other.Date)

			if err1 == nil && err2 == nil {
				daysDiff := math.Abs(date1.Sub(date2).Hours() / 24)

				if daysDiff < 1 { // Same day
					return &models.AnomalyDetail{
						TransactionID:    idx,
						Type:             "duplicate_payment",
						Severity:         "high",
						Score:            0.85,
						Description:      "Potential duplicate payment on same day",
						Amount:           txn.WithdrawalAmt,
						Merchant:         txn.Merchant,
						Category:         txn.Category,
						Date:             txn.Date,
						Reason:           "Same amount and merchant as transaction on " + other.Date,
						StatisticalValue: daysDiff,
					}
				} else if daysDiff < 3 { // Within 3 days
					return &models.AnomalyDetail{
						TransactionID:    idx,
						Type:             "duplicate_payment",
						Severity:         "medium",
						Score:            0.60,
						Description:      "Potential duplicate payment within 3 days",
						Amount:           txn.WithdrawalAmt,
						Merchant:         txn.Merchant,
						Category:         txn.Category,
						Date:             txn.Date,
						Reason:           "Same amount and merchant as transaction " + formatDays(daysDiff) + " ago",
						StatisticalValue: daysDiff,
					}
				}
			}
		}
	}

	return nil
}

// detectRoundAmountAnomaly detects suspicious round amount patterns
func detectRoundAmountAnomaly(txn models.ClassifiedTransaction, idx int) *models.AnomalyDetail {
	amount := txn.WithdrawalAmt

	// Check for suspiciously round amounts (common in fraud)
	var isRound bool
	var roundType string

	// Exact thousands (≥10K)
	if amount >= 10000 && math.Mod(amount, 1000) == 0 {
		isRound = true
		roundType = "exact_thousand"
	}

	// Exact ten-thousands (≥50K)
	if amount >= 50000 && math.Mod(amount, 10000) == 0 {
		isRound = true
		roundType = "exact_ten_thousand"
	}

	// Exact lakhs (≥1L)
	if amount >= 100000 && math.Mod(amount, 100000) == 0 {
		isRound = true
		roundType = "exact_lakh"
	}

	// Only flag if it's a large round amount AND unclassified/unknown category
	suspiciousCategories := map[string]bool{
		"Other": true, "Unknown": true, "": true,
	}

	if isRound && amount >= 25000 && suspiciousCategories[txn.Category] {
		return &models.AnomalyDetail{
			TransactionID:    idx,
			Type:             "round_amount_pattern",
			Severity:         "medium",
			Score:            0.55,
			Description:      "Large round amount to unclassified payee",
			Amount:           amount,
			Merchant:         txn.Merchant,
			Category:         txn.Category,
			Date:             txn.Date,
			Reason:           "Round amount (" + roundType + ") to unclassified category",
			StatisticalValue: amount,
		}
	}

	return nil
}

// detectSpendingSpikeAnomaly detects sudden spending spikes
func detectSpendingSpikeAnomaly(txn models.ClassifiedTransaction, idx int, allTxns []models.ClassifiedTransaction, profile *SpendingProfile) *models.AnomalyDetail {
	txnDate, err := parseTransactionDate(txn.Date)
	if err != nil {
		return nil
	}

	// Calculate spending in last 3 days
	var last3DaysSpend float64
	var last3DaysCount int

	for i := idx - 1; i >= 0 && i >= idx-50; i-- {
		other := allTxns[i]
		if other.WithdrawalAmt == 0 {
			continue
		}

		otherDate, err := parseTransactionDate(other.Date)
		if err != nil {
			continue
		}

		daysDiff := txnDate.Sub(otherDate).Hours() / 24
		if daysDiff >= 0 && daysDiff <= 3 {
			last3DaysSpend += other.WithdrawalAmt
			last3DaysCount++
		}
		if daysDiff > 7 {
			break
		}
	}

	// Compare with average
	if profile.AvgDailySpend == 0 {
		return nil
	}

	expected3DaySpend := profile.AvgDailySpend * 3
	amount := txn.WithdrawalAmt

	// If this single transaction exceeds 2x the expected 3-day spend
	if amount > expected3DaySpend*2 {
		return &models.AnomalyDetail{
			TransactionID:    idx,
			Type:             "spending_spike",
			Severity:         "high",
			Score:            0.75,
			Description:      "Single transaction exceeds 2x typical 3-day spending",
			Amount:           amount,
			Merchant:         txn.Merchant,
			Category:         txn.Category,
			Date:             txn.Date,
			Reason:           "Amount is " + formatRatio(amount/expected3DaySpend) + "x your typical 3-day spending",
			StatisticalValue: amount / expected3DaySpend,
		}
	}

	return nil
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func buildAnomalySummary(anomalies []models.AnomalyDetail) models.AnomalySummary {
	summary := models.AnomalySummary{
		ByType:       make(map[string]int),
		BySeverity:   make(map[string]int),
		TopAnomalies: make([]models.AnomalyDetail, 0),
	}

	// Count by type and severity
	for _, a := range anomalies {
		summary.ByType[a.Type]++
		summary.BySeverity[a.Severity]++
	}

	// Get top 5 by score
	sorted := make([]models.AnomalyDetail, len(anomalies))
	copy(sorted, anomalies)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	if len(sorted) > 5 {
		summary.TopAnomalies = sorted[:5]
	} else {
		summary.TopAnomalies = sorted
	}

	return summary
}

func calculateRiskScore(anomalies []models.AnomalyDetail, totalChecked int) float64 {
	if len(anomalies) == 0 {
		return 0
	}

	severityWeights := map[string]float64{
		"critical": 4.0,
		"high":     3.0,
		"medium":   2.0,
		"low":      1.0,
	}

	var totalScore float64
	for _, a := range anomalies {
		weight := severityWeights[a.Severity]
		totalScore += a.Score * weight
	}

	// Normalize to 0-100
	maxPossibleScore := float64(len(anomalies)) * 4.0
	riskScore := (totalScore / maxPossibleScore) * 100

	// Factor in anomaly rate
	anomalyRate := float64(len(anomalies)) / float64(totalChecked)
	if anomalyRate > 0.1 { // More than 10% anomalies
		riskScore = math.Min(riskScore*1.5, 100)
	}

	return math.Round(riskScore*100) / 100
}

func calculateTransactionDays(transactions []models.ClassifiedTransaction) int {
	if len(transactions) < 2 {
		return 1
	}

	firstDate, err1 := parseTransactionDate(transactions[0].Date)
	lastDate, err2 := parseTransactionDate(transactions[len(transactions)-1].Date)

	if err1 != nil || err2 != nil {
		return 30 // Default estimate
	}

	days := int(math.Abs(lastDate.Sub(firstDate).Hours() / 24))
	if days == 0 {
		days = 1
	}

	return days
}

func parseTransactionDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"02/01/2006",
		"2006-01-02",
		"02-01-2006",
		"01/02/2006", // US format fallback
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, nil
}

func formatDays(days float64) string {
	if days < 1 {
		return "today"
	}
	if days == 1 {
		return "1 day"
	}
	return formatFloat(days) + " days"
}

func formatRatio(ratio float64) string {
	return formatFloat(ratio)
}

func formatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", f), "0"), ".")
}
