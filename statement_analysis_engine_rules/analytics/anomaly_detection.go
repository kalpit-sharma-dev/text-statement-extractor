package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"math"
	"sort"
	"strings"
	"time"
)

// AnomalyType represents the type of anomaly detected
type AnomalyType string

const (
	AnomalyUnusualAmount      AnomalyType = "unusual_amount"
	AnomalyUnusualTime        AnomalyType = "unusual_time"
	AnomalyUnusualMerchant    AnomalyType = "unusual_merchant"
	AnomalyUnusualFrequency   AnomalyType = "unusual_frequency"
	AnomalyUnusualCategory    AnomalyType = "unusual_category"
	AnomalyDuplicatePayment   AnomalyType = "duplicate_payment"
	AnomalySuddenSpike        AnomalyType = "sudden_spike"
	AnomalyRoundAmountPattern AnomalyType = "round_amount_pattern"
)

// Anomaly represents a detected anomaly
type Anomaly struct {
	Type        AnomalyType `json:"type"`
	Severity    string      `json:"severity"`    // "low", "medium", "high", "critical"
	Score       float64     `json:"score"`       // 0.0 to 1.0 anomaly score
	Description string      `json:"description"`
	Transaction models.TransactionDetail `json:"transaction"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// AnomalyDetectionResult contains all detected anomalies
type AnomalyDetectionResult struct {
	Anomalies     []Anomaly `json:"anomalies"`
	RiskScore     float64   `json:"riskScore"`     // Overall risk score 0-100
	TotalChecked  int       `json:"totalChecked"`
	AnomalyCount  int       `json:"anomalyCount"`
}

// UserProfile represents learned spending patterns for a user
type UserProfile struct {
	// Amount statistics per category
	CategoryStats map[string]*CategoryStats `json:"categoryStats"`
	
	// Time-based patterns
	UsualTransactionHours []int   `json:"usualTransactionHours"` // Hours of day (0-23)
	UsualTransactionDays  []int   `json:"usualTransactionDays"`  // Days of week (0-6)
	
	// Merchant patterns
	FrequentMerchants map[string]int `json:"frequentMerchants"`
	
	// Overall spending
	AvgDailySpend   float64 `json:"avgDailySpend"`
	AvgWeeklySpend  float64 `json:"avgWeeklySpend"`
	AvgMonthlySpend float64 `json:"avgMonthlySpend"`
	
	// Transaction frequency
	AvgTransactionsPerDay float64 `json:"avgTransactionsPerDay"`
}

// CategoryStats holds statistical data for a category
type CategoryStats struct {
	Mean      float64   `json:"mean"`
	StdDev    float64   `json:"stdDev"`
	Median    float64   `json:"median"`
	Min       float64   `json:"min"`
	Max       float64   `json:"max"`
	Q1        float64   `json:"q1"`        // 25th percentile
	Q3        float64   `json:"q3"`        // 75th percentile
	IQR       float64   `json:"iqr"`       // Interquartile range
	Count     int       `json:"count"`
	AllValues []float64 `json:"-"` // For calculations
}

// =============================================================================
// RULE-BASED ANOMALY DETECTION (Without ML)
// =============================================================================

// DetectAnomalies performs rule-based anomaly detection on transactions
func DetectAnomalies(transactions []models.ClassifiedTransaction) AnomalyDetectionResult {
	result := AnomalyDetectionResult{
		Anomalies:    make([]Anomaly, 0),
		TotalChecked: len(transactions),
	}
	
	if len(transactions) < 10 {
		// Not enough data for meaningful anomaly detection
		return result
	}
	
	// Step 1: Build user profile from transaction history
	profile := BuildUserProfile(transactions)
	
	// Step 2: Convert to TransactionDetail for analysis
	txnDetails := PrepareTransactionsForResponse(transactions)
	
	// Step 3: Run anomaly detection checks
	for i, txn := range transactions {
		if txn.WithdrawalAmt == 0 {
			continue // Only check expenses
		}
		
		detail := txnDetails[i]
		
		// Check 1: Unusual Amount (Z-Score based)
		if anomaly := detectUnusualAmount(txn, detail, profile); anomaly != nil {
			result.Anomalies = append(result.Anomalies, *anomaly)
		}
		
		// Check 2: Unusual Merchant (First-time or rare)
		if anomaly := detectUnusualMerchant(txn, detail, profile); anomaly != nil {
			result.Anomalies = append(result.Anomalies, *anomaly)
		}
		
		// Check 3: Duplicate Payment Detection
		if anomaly := detectDuplicatePayment(txn, detail, transactions, i); anomaly != nil {
			result.Anomalies = append(result.Anomalies, *anomaly)
		}
		
		// Check 4: Round Amount Pattern (potential fraud indicator)
		if anomaly := detectRoundAmountPattern(txn, detail); anomaly != nil {
			result.Anomalies = append(result.Anomalies, *anomaly)
		}
		
		// Check 5: Sudden Spending Spike
		if anomaly := detectSuddenSpike(txn, detail, transactions, i, profile); anomaly != nil {
			result.Anomalies = append(result.Anomalies, *anomaly)
		}
	}
	
	// Calculate overall risk score
	result.AnomalyCount = len(result.Anomalies)
	result.RiskScore = calculateOverallRiskScore(result.Anomalies, len(transactions))
	
	return result
}

// BuildUserProfile creates a spending profile from transaction history
func BuildUserProfile(transactions []models.ClassifiedTransaction) *UserProfile {
	profile := &UserProfile{
		CategoryStats:     make(map[string]*CategoryStats),
		FrequentMerchants: make(map[string]int),
	}
	
	// Collect amounts by category
	categoryAmounts := make(map[string][]float64)
	var totalSpend float64
	var txnCount int
	
	for _, txn := range transactions {
		if txn.WithdrawalAmt > 0 {
			category := txn.Category
			if category == "" {
				category = "Other"
			}
			categoryAmounts[category] = append(categoryAmounts[category], txn.WithdrawalAmt)
			totalSpend += txn.WithdrawalAmt
			txnCount++
			
			// Track merchant frequency
			merchant := strings.ToUpper(txn.Merchant)
			if merchant != "" && merchant != "UNKNOWN" {
				profile.FrequentMerchants[merchant]++
			}
		}
	}
	
	// Calculate statistics for each category
	for category, amounts := range categoryAmounts {
		profile.CategoryStats[category] = calculateStats(amounts)
	}
	
	// Calculate averages
	if txnCount > 0 {
		// Estimate based on transaction period
		days := estimateTransactionDays(transactions)
		if days > 0 {
			profile.AvgDailySpend = totalSpend / float64(days)
			profile.AvgWeeklySpend = profile.AvgDailySpend * 7
			profile.AvgMonthlySpend = profile.AvgDailySpend * 30
			profile.AvgTransactionsPerDay = float64(txnCount) / float64(days)
		}
	}
	
	return profile
}

// calculateStats calculates statistical measures for a set of values
func calculateStats(values []float64) *CategoryStats {
	if len(values) == 0 {
		return &CategoryStats{}
	}
	
	stats := &CategoryStats{
		Count:     len(values),
		AllValues: values,
	}
	
	// Sort for percentile calculations
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	
	// Min, Max
	stats.Min = sorted[0]
	stats.Max = sorted[len(sorted)-1]
	
	// Mean
	var sum float64
	for _, v := range values {
		sum += v
	}
	stats.Mean = sum / float64(len(values))
	
	// Standard Deviation
	var sumSquares float64
	for _, v := range values {
		diff := v - stats.Mean
		sumSquares += diff * diff
	}
	stats.StdDev = math.Sqrt(sumSquares / float64(len(values)))
	
	// Median (Q2)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		stats.Median = (sorted[mid-1] + sorted[mid]) / 2
	} else {
		stats.Median = sorted[mid]
	}
	
	// Q1 (25th percentile)
	q1Idx := len(sorted) / 4
	stats.Q1 = sorted[q1Idx]
	
	// Q3 (75th percentile)
	q3Idx := (len(sorted) * 3) / 4
	stats.Q3 = sorted[q3Idx]
	
	// IQR (Interquartile Range)
	stats.IQR = stats.Q3 - stats.Q1
	
	return stats
}

// =============================================================================
// ANOMALY DETECTION METHODS
// =============================================================================

// detectUnusualAmount uses Z-Score and IQR methods to detect unusual amounts
func detectUnusualAmount(txn models.ClassifiedTransaction, detail models.TransactionDetail, profile *UserProfile) *Anomaly {
	category := txn.Category
	if category == "" {
		category = "Other"
	}
	
	stats, exists := profile.CategoryStats[category]
	if !exists || stats.Count < 5 {
		return nil // Not enough data
	}
	
	amount := txn.WithdrawalAmt
	
	// Method 1: Z-Score (how many standard deviations from mean)
	var zScore float64
	if stats.StdDev > 0 {
		zScore = (amount - stats.Mean) / stats.StdDev
	}
	
	// Method 2: IQR-based outlier detection (Tukey's method)
	lowerBound := stats.Q1 - (1.5 * stats.IQR)
	upperBound := stats.Q3 + (1.5 * stats.IQR)
	isIQROutlier := amount < lowerBound || amount > upperBound
	
	// Combined scoring
	var anomalyScore float64
	var severity string
	
	absZScore := math.Abs(zScore)
	if absZScore > 4 || amount > upperBound*2 {
		anomalyScore = 0.95
		severity = "critical"
	} else if absZScore > 3 || amount > upperBound*1.5 {
		anomalyScore = 0.80
		severity = "high"
	} else if absZScore > 2.5 || isIQROutlier {
		anomalyScore = 0.60
		severity = "medium"
	} else if absZScore > 2 {
		anomalyScore = 0.40
		severity = "low"
	} else {
		return nil // Not anomalous
	}
	
	return &Anomaly{
		Type:        AnomalyUnusualAmount,
		Severity:    severity,
		Score:       anomalyScore,
		Description: formatAmountAnomalyDescription(amount, stats, zScore),
		Transaction: detail,
		Context: map[string]interface{}{
			"zScore":      zScore,
			"mean":        stats.Mean,
			"stdDev":      stats.StdDev,
			"upperBound":  upperBound,
			"categoryAvg": stats.Mean,
		},
	}
}

// detectUnusualMerchant detects transactions with new or rare merchants
func detectUnusualMerchant(txn models.ClassifiedTransaction, detail models.TransactionDetail, profile *UserProfile) *Anomaly {
	merchant := strings.ToUpper(txn.Merchant)
	if merchant == "" || merchant == "UNKNOWN" {
		return nil
	}
	
	frequency, exists := profile.FrequentMerchants[merchant]
	
	// First-time merchant with large amount
	if !exists && txn.WithdrawalAmt > profile.AvgDailySpend*3 {
		return &Anomaly{
			Type:        AnomalyUnusualMerchant,
			Severity:    "medium",
			Score:       0.65,
			Description: "First-time merchant with unusually large transaction amount",
			Transaction: detail,
			Context: map[string]interface{}{
				"merchantFrequency": 0,
				"avgDailySpend":     profile.AvgDailySpend,
			},
		}
	}
	
	// Very rare merchant (used only once before) with large amount
	if frequency == 1 && txn.WithdrawalAmt > profile.AvgDailySpend*2 {
		return &Anomaly{
			Type:        AnomalyUnusualMerchant,
			Severity:    "low",
			Score:       0.45,
			Description: "Rare merchant with above-average transaction amount",
			Transaction: detail,
			Context: map[string]interface{}{
				"merchantFrequency": frequency,
				"avgDailySpend":     profile.AvgDailySpend,
			},
		}
	}
	
	return nil
}

// detectDuplicatePayment detects potential duplicate payments
func detectDuplicatePayment(txn models.ClassifiedTransaction, detail models.TransactionDetail, allTxns []models.ClassifiedTransaction, currentIdx int) *Anomaly {
	// Look at recent transactions (last 7 days window)
	for i := currentIdx - 1; i >= 0 && i >= currentIdx-20; i-- {
		other := allTxns[i]
		
		// Same amount, same merchant, within short time
		if other.WithdrawalAmt == txn.WithdrawalAmt &&
			strings.EqualFold(other.Merchant, txn.Merchant) &&
			other.Merchant != "" {
			
			// Parse dates to check time difference
			date1, _ := parseDate(txn.Date)
			date2, _ := parseDate(other.Date)
			
			daysDiff := math.Abs(date1.Sub(date2).Hours() / 24)
			
			if daysDiff < 1 { // Same day
				return &Anomaly{
					Type:        AnomalyDuplicatePayment,
					Severity:    "high",
					Score:       0.85,
					Description: "Potential duplicate payment - same amount and merchant on same day",
					Transaction: detail,
					Context: map[string]interface{}{
						"duplicateDate": other.Date,
						"merchant":      txn.Merchant,
						"amount":        txn.WithdrawalAmt,
					},
				}
			} else if daysDiff < 3 { // Within 3 days
				return &Anomaly{
					Type:        AnomalyDuplicatePayment,
					Severity:    "medium",
					Score:       0.60,
					Description: "Potential duplicate payment - same amount and merchant within 3 days",
					Transaction: detail,
					Context: map[string]interface{}{
						"duplicateDate": other.Date,
						"daysDiff":      daysDiff,
					},
				}
			}
		}
	}
	
	return nil
}

// detectRoundAmountPattern detects suspicious round amount patterns
func detectRoundAmountPattern(txn models.ClassifiedTransaction, detail models.TransactionDetail) *Anomaly {
	amount := txn.WithdrawalAmt
	
	// Check for suspiciously round amounts (common in fraud)
	isRound := false
	var roundType string
	
	// Exact thousands
	if amount >= 10000 && math.Mod(amount, 1000) == 0 {
		isRound = true
		roundType = "exact_thousand"
	}
	
	// Exact ten-thousands
	if amount >= 50000 && math.Mod(amount, 10000) == 0 {
		isRound = true
		roundType = "exact_ten_thousand"
	}
	
	// Exact lakhs
	if amount >= 100000 && math.Mod(amount, 100000) == 0 {
		isRound = true
		roundType = "exact_lakh"
	}
	
	// Only flag if it's a large round amount AND unusual category
	suspiciousCategories := map[string]bool{
		"Other": true, "Unknown": true, "": true,
	}
	
	if isRound && amount >= 25000 && suspiciousCategories[txn.Category] {
		return &Anomaly{
			Type:        AnomalyRoundAmountPattern,
			Severity:    "medium",
			Score:       0.55,
			Description: "Large round amount to unclassified payee",
			Transaction: detail,
			Context: map[string]interface{}{
				"roundType": roundType,
				"amount":    amount,
			},
		}
	}
	
	return nil
}

// detectSuddenSpike detects sudden spending spikes
func detectSuddenSpike(txn models.ClassifiedTransaction, detail models.TransactionDetail, allTxns []models.ClassifiedTransaction, currentIdx int, profile *UserProfile) *Anomaly {
	// Calculate spending in last 3 days vs this transaction
	txnDate, err := parseDate(txn.Date)
	if err != nil {
		return nil
	}
	
	var last3DaysSpend float64
	var last3DaysCount int
	
	for i := currentIdx - 1; i >= 0 && i >= currentIdx-50; i-- {
		other := allTxns[i]
		otherDate, err := parseDate(other.Date)
		if err != nil {
			continue
		}
		
		daysDiff := txnDate.Sub(otherDate).Hours() / 24
		if daysDiff >= 0 && daysDiff <= 3 && other.WithdrawalAmt > 0 {
			last3DaysSpend += other.WithdrawalAmt
			last3DaysCount++
		}
		if daysDiff > 7 {
			break // Stop looking further back
		}
	}
	
	// Compare with average
	avgDailySpend := profile.AvgDailySpend
	if avgDailySpend == 0 {
		return nil
	}
	
	expected3DaySpend := avgDailySpend * 3
	
	// If this single transaction exceeds 3-day average
	if txn.WithdrawalAmt > expected3DaySpend*2 {
		return &Anomaly{
			Type:        AnomalySuddenSpike,
			Severity:    "high",
			Score:       0.75,
			Description: "Single transaction exceeds 2x your typical 3-day spending",
			Transaction: detail,
			Context: map[string]interface{}{
				"amount":           txn.WithdrawalAmt,
				"expected3DaySpend": expected3DaySpend,
				"avgDailySpend":    avgDailySpend,
			},
		}
	}
	
	return nil
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func formatAmountAnomalyDescription(amount float64, stats *CategoryStats, zScore float64) string {
	if amount > stats.Max {
		return "Transaction amount is the highest ever in this category"
	}
	if zScore > 3 {
		return "Transaction amount is significantly higher than usual for this category"
	}
	if zScore > 2 {
		return "Transaction amount is above typical range for this category"
	}
	return "Transaction amount deviates from normal pattern"
}

func parseDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"02/01/2006",
		"2006-01-02",
		"02-01-2006",
	}
	
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, nil
}

func estimateTransactionDays(transactions []models.ClassifiedTransaction) int {
	if len(transactions) < 2 {
		return 1
	}
	
	firstDate, _ := parseDate(transactions[0].Date)
	lastDate, _ := parseDate(transactions[len(transactions)-1].Date)
	
	days := int(math.Abs(lastDate.Sub(firstDate).Hours() / 24))
	if days == 0 {
		days = 1
	}
	
	return days
}

func calculateOverallRiskScore(anomalies []Anomaly, totalTxns int) float64 {
	if len(anomalies) == 0 {
		return 0
	}
	
	var totalScore float64
	severityWeights := map[string]float64{
		"critical": 4.0,
		"high":     3.0,
		"medium":   2.0,
		"low":      1.0,
	}
	
	for _, a := range anomalies {
		weight := severityWeights[a.Severity]
		totalScore += a.Score * weight
	}
	
	// Normalize to 0-100 scale
	maxPossibleScore := float64(len(anomalies)) * 4.0 // If all were critical with score 1.0
	riskScore := (totalScore / maxPossibleScore) * 100
	
	// Also factor in anomaly rate
	anomalyRate := float64(len(anomalies)) / float64(totalTxns)
	if anomalyRate > 0.1 { // More than 10% anomalies
		riskScore = math.Min(riskScore*1.5, 100)
	}
	
	return math.Round(riskScore*100) / 100
}

