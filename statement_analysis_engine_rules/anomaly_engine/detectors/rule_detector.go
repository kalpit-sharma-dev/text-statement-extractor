package detectors

import (
	"fmt"
	"math"
	"strings"

	"classify/statement_analysis_engine_rules/anomaly_engine/profiles"
	"classify/statement_analysis_engine_rules/anomaly_engine/types"
)

// RuleDetector implements rule-based anomaly detection
// Uses business rules and thresholds (bank-grade rules)
type RuleDetector struct {
	BaseDetector
	config *RuleConfig
}

// RuleConfig holds rule-based detection thresholds
type RuleConfig struct {
	// Amount thresholds
	LargeAmountThreshold     float64 // Default: 50000
	VeryLargeAmountThreshold float64 // Default: 100000

	// Round amount detection
	RoundAmountMin float64 // Default: 25000

	// Unknown merchant threshold
	UnknownMerchantThreshold float64 // Default: 10000

	// Duplicate payment window (days)
	DuplicateWindowDays float64 // Default: 3
}

// DefaultRuleConfig returns default rule configuration
func DefaultRuleConfig() *RuleConfig {
	return &RuleConfig{
		LargeAmountThreshold:     50000,
		VeryLargeAmountThreshold: 100000,
		RoundAmountMin:           25000,
		UnknownMerchantThreshold: 10000,
		DuplicateWindowDays:      3,
	}
}

// NewRuleDetector creates a new rule-based detector
func NewRuleDetector(config *RuleConfig) *RuleDetector {
	if config == nil {
		config = DefaultRuleConfig()
	}
	return &RuleDetector{
		BaseDetector: BaseDetector{name: "RuleDetector"},
		config:       config,
	}
}

// Detect implements Detector interface
func (r *RuleDetector) Detect(ctx types.TransactionContext, profile *profiles.UserProfile) []types.AnomalySignal {
	signals := make([]types.AnomalySignal, 0)
	
	txn := ctx.Txn
	amount := txn.WithdrawalAmt
	
	if amount == 0 {
		return signals // Only check expenses
	}
	
	// Skip if amount is too small (noise reduction)
	if amount < 1000.0 {
		return signals
	}

	// Rule 1: Large amount detection
	if amount >= r.config.VeryLargeAmountThreshold {
		score := math.Min(90, 60+(amount/r.config.VeryLargeAmountThreshold)*10)
		signals = append(signals, types.NewSignal(
			types.SignalHighAmount,
			types.CategoryAmount,
			score,
			formatLargeAmount(amount, r.config.VeryLargeAmountThreshold),
		))
	} else if amount >= r.config.LargeAmountThreshold {
		score := math.Min(70, 40+(amount/r.config.LargeAmountThreshold)*15)
		signals = append(signals, types.NewSignal(
			types.SignalHighAmount,
			types.CategoryAmount,
			score,
			formatLargeAmount(amount, r.config.LargeAmountThreshold),
		))
	}

	// Rule 2: Round amount pattern (fraud indicator)
	if amount >= r.config.RoundAmountMin {
		if isRoundAmount(amount) {
			// Only flag if category is suspicious
			suspiciousCategories := map[string]bool{
				"Other": true, "Unknown": true, "": true,
			}
			if suspiciousCategories[txn.Category] {
				signals = append(signals, types.NewSignal(
					types.SignalRoundAmount,
					types.CategoryPattern,
					55,
					formatRoundAmount(amount),
				))
			}
		}
	}

	// Rule 3: Unknown merchant with large amount
	merchant := strings.ToUpper(strings.TrimSpace(txn.Merchant))
	if (merchant == "" || merchant == "UNKNOWN") && amount >= r.config.UnknownMerchantThreshold {
		score := math.Min(65, 30+(amount/r.config.UnknownMerchantThreshold)*10)
		signals = append(signals, types.NewSignal(
			types.SignalUnknownMerchant,
			types.CategoryMerchant,
			score,
			formatUnknownMerchant(amount),
		))
	}

	return signals
}

// isRoundAmount checks if amount is suspiciously round
func isRoundAmount(amount float64) bool {
	// Exact thousands (≥10K)
	if amount >= 10000 && math.Mod(amount, 1000) == 0 {
		return true
	}
	// Exact ten-thousands (≥50K)
	if amount >= 50000 && math.Mod(amount, 10000) == 0 {
		return true
	}
	// Exact lakhs (≥1L)
	if amount >= 100000 && math.Mod(amount, 100000) == 0 {
		return true
	}
	return false
}

// User-friendly explanations (Zerodha/HDFC style)
func formatLargeAmount(amount, threshold float64) string {
	ratio := amount / threshold
	return fmt.Sprintf("This is %.1f× your typical large transaction threshold", ratio)
}

func formatRoundAmount(amount float64) string {
	// Friendly, non-alarming explanation
	return fmt.Sprintf("Round amount transaction - just for your awareness")
}

func formatUnknownMerchant(amount float64) string {
	// Friendly explanation
	return fmt.Sprintf("Transaction to a merchant not in your usual spending patterns")
}

// formatAmount and formatRatio use utils package
