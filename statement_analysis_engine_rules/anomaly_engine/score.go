package anomaly_engine

import (
	"math"
	"sort"
)

import "classify/statement_analysis_engine_rules/anomaly_engine/types"

// AnomalyResult represents the final anomaly detection result
type AnomalyResult struct {
	Signals        []types.AnomalySignal `json:"signals"`
	FinalScore     float64               `json:"finalScore"`     // 0-100
	Severity       types.Severity        `json:"severity"`
	TopSignals     []types.AnomalySignal `json:"topSignals"`     // Top 3 by score
	Explanation    string                `json:"explanation"`   // Human-readable summary
	Confidence     float64               `json:"confidence"`     // 0-1
	RiskFlags      []string              `json:"riskFlags"`     // RBI AML-style flags
}

// Scorer aggregates signals into final anomaly score
type Scorer struct {
	// Aggregation weights
	MaxSignalWeight    float64 // Weight for max signal (default: 0.6)
	SumSignalWeight    float64 // Weight for sum of signals (default: 0.3)
	CountSignalWeight  float64 // Weight for signal count (default: 0.1)
}

// NewScorer creates a new scorer with default weights
func NewScorer() *Scorer {
	return &Scorer{
		MaxSignalWeight:   0.6,
		SumSignalWeight:   0.3,
		CountSignalWeight: 0.1,
	}
}

// Score aggregates signals into final anomaly result
// Uses bank-grade aggregation: max(signal) + log(sum(signals)) * weight
func (s *Scorer) Score(signals []types.AnomalySignal) AnomalyResult {
	if len(signals) == 0 {
		return AnomalyResult{
			Signals:    signals,
			FinalScore: 0,
			Severity:   SeverityInfo,
			Explanation: "No anomalies detected",
			Confidence: 1.0,
		}
	}

	// Sort signals by score (descending)
	sorted := make([]types.AnomalySignal, len(signals))
	copy(sorted, signals)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	// Get top signal (max)
	maxSignal := sorted[0]
	maxScore := maxSignal.Score

	// Calculate sum of all signal scores
	var sumScore float64
	for _, sig := range signals {
		sumScore += sig.Score
	}

	// Calculate signal count contribution (diminishing returns)
	countScore := math.Min(float64(len(signals))*10, 50) // Max 50 points for count

	// Aggregation formula: max + log(sum) * weight + count * weight
	// Using log to prevent dilution by many weak signals
	logSum := 0.0
	if sumScore > 0 {
		logSum = math.Log(sumScore+1) * 20 // Scale log appropriately
	}

	finalScore := (maxScore * s.MaxSignalWeight) +
		(logSum * s.SumSignalWeight) +
		(countScore * s.CountSignalWeight)

	// Cap at 100
	finalScore = math.Min(finalScore, 100)

	// Get top 3 signals
	topCount := 3
	if len(sorted) < topCount {
		topCount = len(sorted)
	}
	topSignals := sorted[:topCount]

	// Generate explanation
	explanation := generateExplanation(sorted, finalScore)

	// Calculate confidence (based on signal strength and count)
	confidence := calculateConfidence(sorted)

	// Generate risk flags (RBI AML-style)
	riskFlags := generateRiskFlags(sorted)

	return AnomalyResult{
		Signals:     signals,
		FinalScore:  math.Round(finalScore*100) / 100,
		Severity:    types.SeverityFromScore(finalScore),
		TopSignals:   topSignals,
		Explanation:  explanation,
		Confidence:   confidence,
		RiskFlags:    riskFlags,
	}
}

// generateExplanation creates user-friendly, insight-driven explanation
// Never uses words like "anomaly", "risk", "fraud" - Zerodha/HDFC style
func generateExplanation(signals []types.AnomalySignal, finalScore float64) string {
	if len(signals) == 0 {
		return "No unusual patterns detected"
	}

	top := signals[0]
	
	// Single strong signal
	if len(signals) == 1 {
		return top.Explanation
	}

	// Multiple signals - summarize in friendly way
	if finalScore >= 90 {
		return "Multiple spending patterns detected: " + top.Explanation
	} else if finalScore >= 70 {
		return "Unusual spending pattern: " + top.Explanation
	} else if len(signals) > 3 {
		return "Multiple pattern changes detected across different categories"
	}

	return top.Explanation
}

// calculateConfidence calculates confidence score (0-1)
func calculateConfidence(signals []types.AnomalySignal) float64 {
	if len(signals) == 0 {
		return 1.0
	}

	// Confidence increases with:
	// 1. Higher signal scores
	// 2. More signals (up to a point)
	// 3. Consistent signal categories

	var avgScore float64
	for _, sig := range signals {
		avgScore += sig.Score
	}
	avgScore /= float64(len(signals))

	// Base confidence from average score
	confidence := avgScore / 100.0

	// Boost if multiple signals agree
	if len(signals) >= 2 {
		confidence = math.Min(confidence*1.2, 1.0)
	}

	// Boost if signals are from same category (stronger agreement)
	categoryCount := make(map[string]int)
	for _, sig := range signals {
		categoryCount[sig.Category]++
	}
	maxCategoryCount := 0
	for _, count := range categoryCount {
		if count > maxCategoryCount {
			maxCategoryCount = count
		}
	}
	if maxCategoryCount >= 2 {
		confidence = math.Min(confidence*1.1, 1.0)
	}

	return math.Round(confidence*1000) / 1000
}

// generateRiskFlags generates RBI AML-style risk flags
func generateRiskFlags(signals []types.AnomalySignal) []string {
	flags := make([]string, 0)
	flagSet := make(map[string]bool)

	for _, sig := range signals {
		// Critical/High severity signals generate flags
		if sig.Severity == types.SeverityCritical || sig.Severity == types.SeverityHigh {
			switch sig.Code {
			case types.SignalHighAmount, types.SignalAmountSpike:
				if !flagSet["LARGE_AMOUNT"] {
					flags = append(flags, "LARGE_AMOUNT")
					flagSet["LARGE_AMOUNT"] = true
				}
			case types.SignalNewMerchant, types.SignalUnknownMerchant:
				if !flagSet["UNKNOWN_MERCHANT"] {
					flags = append(flags, "UNKNOWN_MERCHANT")
					flagSet["UNKNOWN_MERCHANT"] = true
				}
			case types.SignalDuplicatePayment:
				if !flagSet["DUPLICATE_TRANSACTION"] {
					flags = append(flags, "DUPLICATE_TRANSACTION")
					flagSet["DUPLICATE_TRANSACTION"] = true
				}
			case types.SignalRoundAmount:
				if !flagSet["SUSPICIOUS_PATTERN"] {
					flags = append(flags, "SUSPICIOUS_PATTERN")
					flagSet["SUSPICIOUS_PATTERN"] = true
				}
			case types.SignalSpendingSpike:
				if !flagSet["UNUSUAL_SPENDING"] {
					flags = append(flags, "UNUSUAL_SPENDING")
					flagSet["UNUSUAL_SPENDING"] = true
				}
			case types.SignalMultipleLargeTransfers:
				if !flagSet["BENEFICIARY_CONCENTRATION"] {
					flags = append(flags, "BENEFICIARY_CONCENTRATION")
					flagSet["BENEFICIARY_CONCENTRATION"] = true
				}
			case types.SignalHighValueRecurring:
				if !flagSet["HIGH_RECURRING_COMMITMENT"] {
					flags = append(flags, "HIGH_RECURRING_COMMITMENT")
					flagSet["HIGH_RECURRING_COMMITMENT"] = true
				}
			case types.SignalLargeBillPayment:
				if !flagSet["UNUSUAL_BILL_AMOUNT"] {
					flags = append(flags, "UNUSUAL_BILL_AMOUNT")
					flagSet["UNUSUAL_BILL_AMOUNT"] = true
				}
			case types.SignalIncomeDisruption:
				if !flagSet["INCOME_DISRUPTION"] {
					flags = append(flags, "INCOME_DISRUPTION")
					flagSet["INCOME_DISRUPTION"] = true
				}
			}
		}
	}

	return flags
}

