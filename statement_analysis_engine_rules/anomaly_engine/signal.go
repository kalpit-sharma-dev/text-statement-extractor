package anomaly_engine

import "classify/statement_analysis_engine_rules/anomaly_engine/types"

// Re-export types for convenience
type AnomalySignal = types.AnomalySignal
type Severity = types.Severity

// Re-export constants
const (
	SignalHighAmount       = types.SignalHighAmount
	SignalAmountSpike      = types.SignalAmountSpike
	SignalUnusualAmount    = types.SignalUnusualAmount
	SignalRoundAmount      = types.SignalRoundAmount
	SignalNewMerchant      = types.SignalNewMerchant
	SignalRareMerchant     = types.SignalRareMerchant
	SignalUnknownMerchant  = types.SignalUnknownMerchant
	SignalDuplicatePayment = types.SignalDuplicatePayment
	SignalSpendingSpike    = types.SignalSpendingSpike
	SignalMLAnomaly        = types.SignalMLAnomaly
)

const (
	CategoryAmount    = types.CategoryAmount
	CategoryFrequency = types.CategoryFrequency
	CategoryBehavior  = types.CategoryBehavior
	CategoryMerchant  = types.CategoryMerchant
	CategoryPattern   = types.CategoryPattern
	CategoryML        = types.CategoryML
)

const (
	SeverityInfo     = types.SeverityInfo
	SeverityLow      = types.SeverityLow
	SeverityMedium   = types.SeverityMedium
	SeverityHigh     = types.SeverityHigh
	SeverityCritical = types.SeverityCritical
)

// NewSignal creates a new anomaly signal
func NewSignal(code, category string, score float64, explanation string) AnomalySignal {
	return types.NewSignal(code, category, score, explanation)
}

func severityFromScore(score float64) Severity {
	return types.SeverityFromScore(score)
}

