package profiles

import (
	"classify/statement_analysis_engine_rules/models"
	"math"
	"sort"
	"strings"
	"time"
)

// UserProfile represents user's spending patterns and behavior
// This is built from historical transaction data
type UserProfile struct {
	// Amount statistics
	AvgTxnAmount    float64 `json:"avgTxnAmount"`
	StdDevTxnAmount float64 `json:"stdDevTxnAmount"`
	MedianTxnAmount float64 `json:"medianTxnAmount"`
	MinTxnAmount    float64 `json:"minTxnAmount"`
	MaxTxnAmount    float64 `json:"maxTxnAmount"`
	
	// Spending patterns
	DailySpendAvg   float64 `json:"dailySpendAvg"`
	AvgDailySpend   float64 `json:"avgDailySpend"`   // Alias for DailySpendAvg
	WeeklySpendAvg  float64 `json:"weeklySpendAvg"`
	AvgWeeklySpend  float64 `json:"avgWeeklySpend"`  // Alias for WeeklySpendAvg
	MonthlySpendAvg float64 `json:"monthlySpendAvg"`
	AvgMonthlySpend float64 `json:"avgMonthlySpend"` // Alias for MonthlySpendAvg
	
	// Merchant patterns
	KnownMerchants   map[string]int  `json:"knownMerchants"`   // Merchant -> frequency
	MerchantAmounts  map[string][]float64 `json:"-"`            // Internal: amounts per merchant
	
	// Category patterns
	CategoryProfiles map[string]*CategoryProfile `json:"categoryProfiles"`
	
	// Time patterns
	ActiveHours      map[int]int     `json:"activeHours"`      // Hour (0-23) -> transaction count
	ActiveDays       map[int]int     `json:"activeDays"`       // Day of week (0-6) -> transaction count
	
	// Frequency
	AvgTransactionsPerDay float64 `json:"avgTransactionsPerDay"`
	TotalTransactions     int     `json:"totalTransactions"`
	TransactionDays       int     `json:"transactionDays"`
	
	// Percentiles (for robust outlier detection)
	P95Amount float64 `json:"p95Amount"`
	P99Amount float64 `json:"p99Amount"`
}

// CategoryProfile holds statistics for a specific category
type CategoryProfile struct {
	Category   string    `json:"category"`
	Mean       float64   `json:"mean"`
	Median     float64   `json:"median"`
	StdDev     float64   `json:"stdDev"`
	Min        float64   `json:"min"`
	Max        float64   `json:"max"`
	Q1         float64   `json:"q1"`         // 25th percentile
	Q3         float64   `json:"q3"`         // 75th percentile
	IQR        float64   `json:"iqr"`        // Interquartile range
	P95        float64   `json:"p95"`        // 95th percentile
	P99        float64   `json:"p99"`        // 99th percentile
	Count      int       `json:"count"`
	Values     []float64 `json:"-"`          // Internal: for calculations
}

// BuildUserProfile builds user profile from transaction history
func BuildUserProfile(transactions []models.ClassifiedTransaction) *UserProfile {
	profile := &UserProfile{
		KnownMerchants:   make(map[string]int),
		MerchantAmounts:  make(map[string][]float64),
		CategoryProfiles: make(map[string]*CategoryProfile),
		ActiveHours:      make(map[int]int),
		ActiveDays:       make(map[int]int),
	}

	// Collect all amounts
	var allAmounts []float64
	var totalSpend float64
	var expenseCount int

	// Collect amounts by category
	categoryAmounts := make(map[string][]float64)

	for _, txn := range transactions {
		if txn.WithdrawalAmt > 0 {
			amount := txn.WithdrawalAmt
			allAmounts = append(allAmounts, amount)
			totalSpend += amount
			expenseCount++

			// Category statistics
			category := txn.Category
			if category == "" {
				category = "Other"
			}
			categoryAmounts[category] = append(categoryAmounts[category], amount)

			// Merchant tracking
			merchant := strings.ToUpper(strings.TrimSpace(txn.Merchant))
			if merchant != "" && merchant != "UNKNOWN" {
				profile.KnownMerchants[merchant]++
				profile.MerchantAmounts[merchant] = append(profile.MerchantAmounts[merchant], amount)
			}

			// Time patterns
			timestamp, err := parseTransactionTimestamp(txn.Date)
			if err == nil {
				hour := timestamp.Hour()
				dayOfWeek := int(timestamp.Weekday())
				profile.ActiveHours[hour]++
				profile.ActiveDays[dayOfWeek]++
			}
		}
	}

	// Calculate overall statistics
	if len(allAmounts) > 0 {
		profile.TotalTransactions = expenseCount
		profile.AvgTxnAmount = totalSpend / float64(expenseCount)
		profile.StdDevTxnAmount = calculateStdDev(allAmounts, profile.AvgTxnAmount)
		profile.MedianTxnAmount = calculateMedian(allAmounts)
		profile.MinTxnAmount = allAmounts[0]
		profile.MaxTxnAmount = allAmounts[len(allAmounts)-1]
		profile.P95Amount = calculatePercentile(allAmounts, 0.95)
		profile.P99Amount = calculatePercentile(allAmounts, 0.99)
	}

	// Calculate category profiles
	for category, amounts := range categoryAmounts {
		profile.CategoryProfiles[category] = calculateCategoryProfile(category, amounts)
	}

	// Calculate time-based averages
	days := calculateTransactionDays(transactions)
	if days > 0 {
		profile.TransactionDays = days
		profile.DailySpendAvg = totalSpend / float64(days)
		profile.AvgDailySpend = profile.DailySpendAvg
		profile.WeeklySpendAvg = profile.DailySpendAvg * 7
		profile.AvgWeeklySpend = profile.WeeklySpendAvg
		profile.MonthlySpendAvg = profile.DailySpendAvg * 30
		profile.AvgMonthlySpend = profile.MonthlySpendAvg
		profile.AvgTransactionsPerDay = float64(expenseCount) / float64(days)
	}

	return profile
}

// calculateCategoryProfile calculates comprehensive statistics for a category
func calculateCategoryProfile(category string, values []float64) *CategoryProfile {
	if len(values) == 0 {
		return &CategoryProfile{Category: category}
	}

	cp := &CategoryProfile{
		Category: category,
		Count:    len(values),
		Values:   values,
	}

	// Sort for percentile calculations
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	// Min, Max
	cp.Min = sorted[0]
	cp.Max = sorted[len(sorted)-1]

	// Mean
	var sum float64
	for _, v := range values {
		sum += v
	}
	cp.Mean = sum / float64(len(values))

	// Standard Deviation
	cp.StdDev = calculateStdDev(values, cp.Mean)

	// Median (Q2)
	cp.Median = calculateMedian(sorted)

	// Q1 (25th percentile)
	q1Idx := len(sorted) / 4
	if q1Idx >= len(sorted) {
		q1Idx = len(sorted) - 1
	}
	cp.Q1 = sorted[q1Idx]

	// Q3 (75th percentile)
	q3Idx := (len(sorted) * 3) / 4
	if q3Idx >= len(sorted) {
		q3Idx = len(sorted) - 1
	}
	cp.Q3 = sorted[q3Idx]

	// IQR
	cp.IQR = cp.Q3 - cp.Q1

	// P95, P99
	cp.P95 = calculatePercentile(sorted, 0.95)
	cp.P99 = calculatePercentile(sorted, 0.99)

	return cp
}

// Helper functions

func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sumSquares float64
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(values)))
}

func calculateMedian(sorted []float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func calculatePercentile(sorted []float64, percentile float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)) * percentile)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func parseTransactionTimestamp(dateStr string) (time.Time, error) {
	layouts := []string{
		"02/01/2006",
		"2006-01-02",
		"02-01-2006",
		"01/02/2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Now(), nil
}

func calculateTransactionDays(transactions []models.ClassifiedTransaction) int {
	if len(transactions) < 2 {
		return 1
	}

	firstDate, err1 := parseTransactionTimestamp(transactions[0].Date)
	lastDate, err2 := parseTransactionTimestamp(transactions[len(transactions)-1].Date)

	if err1 != nil || err2 != nil {
		return 30 // Default estimate
	}

	days := int(math.Abs(lastDate.Sub(firstDate).Hours() / 24))
	if days == 0 {
		days = 1
	}

	return days
}

