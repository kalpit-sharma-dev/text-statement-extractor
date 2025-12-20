package types

import (
	"classify/statement_analysis_engine_rules/models"
	"time"
)

// TransactionContext holds transaction and user context
type TransactionContext struct {
	Txn       models.ClassifiedTransaction
	UserID    string
	Timestamp time.Time
	Location  string
	DeviceID  string
}

// AnomalySignal represents an atomic anomaly signal
type AnomalySignal struct {
	Code        string
	Category    string
	Score       float64
	Severity    Severity
	Explanation string
	Metadata    map[string]interface{}
}

// Severity levels
type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// Signal codes
const (
	SignalHighAmount        = "HIGH_AMOUNT"
	SignalAmountSpike       = "AMOUNT_SPIKE"
	SignalUnusualAmount     = "UNUSUAL_AMOUNT"
	SignalRoundAmount       = "ROUND_AMOUNT"
	SignalNewMerchant       = "NEW_MERCHANT"
	SignalRareMerchant      = "RARE_MERCHANT"
	SignalUnknownMerchant   = "UNKNOWN_MERCHANT"
	SignalDuplicatePayment      = "DUPLICATE_PAYMENT"
	SignalSpendingSpike         = "SPENDING_SPIKE"
	SignalMultipleLargeTransfers = "MULTIPLE_LARGE_TRANSFERS"
	SignalHighValueRecurring    = "HIGH_VALUE_RECURRING"
	SignalLargeBillPayment      = "LARGE_BILL_PAYMENT"
	SignalIncomeDisruption      = "INCOME_DISRUPTION"
	SignalMLAnomaly             = "ML_ANOMALY"
)

// Signal categories
const (
	CategoryAmount    = "Amount"
	CategoryFrequency = "Frequency"
	CategoryBehavior  = "Behavior"
	CategoryMerchant  = "Merchant"
	CategoryPattern   = "Pattern"
	CategoryML        = "ML"
)

// NewSignal creates a new anomaly signal
func NewSignal(code, category string, score float64, explanation string) AnomalySignal {
	return AnomalySignal{
		Code:        code,
		Category:    category,
		Score:       score,
		Severity:    SeverityFromScore(score),
		Explanation: explanation,
		Metadata:    make(map[string]interface{}),
	}
}

// SeverityFromScore converts score to severity
func SeverityFromScore(score float64) Severity {
	switch {
	case score >= 90:
		return SeverityCritical
	case score >= 70:
		return SeverityHigh
	case score >= 45:
		return SeverityMedium
	case score >= 20:
		return SeverityLow
	default:
		return SeverityInfo
	}
}

