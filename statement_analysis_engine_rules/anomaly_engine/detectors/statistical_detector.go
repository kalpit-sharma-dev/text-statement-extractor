package detectors

import (
	"fmt"
	"math"
	"strings"

	"classify/statement_analysis_engine_rules/anomaly_engine/profiles"
	"classify/statement_analysis_engine_rules/anomaly_engine/types"
)

// StatisticalDetector implements statistical-based anomaly detection
// Uses Z-Score, IQR, percentiles (from existing logic)
type StatisticalDetector struct {
	BaseDetector
	config *StatisticalConfig
}

// StatisticalConfig holds statistical detection thresholds
type StatisticalConfig struct {
	// Z-Score thresholds
	ZScoreCritical float64 // Default: 4.0
	ZScoreHigh     float64 // Default: 3.0
	ZScoreMedium   float64 // Default: 2.5
	ZScoreLow      float64 // Default: 2.0
	
	// IQR multiplier (Tukey's method)
	IQRMultiplier float64 // Default: 1.5
	
	// Percentile thresholds
	UseP99 bool // Use 99th percentile as critical threshold
	UseP95 bool // Use 95th percentile as high threshold
}

// DefaultStatisticalConfig returns default statistical configuration
func DefaultStatisticalConfig() *StatisticalConfig {
	return &StatisticalConfig{
		ZScoreCritical: 4.0,
		ZScoreHigh:     3.0,
		ZScoreMedium:   2.5,
		ZScoreLow:      2.0,
		IQRMultiplier:  1.5,
		UseP99:         true,
		UseP95:         true,
	}
}

// NewStatisticalDetector creates a new statistical detector
func NewStatisticalDetector(config *StatisticalConfig) *StatisticalDetector {
	if config == nil {
		config = DefaultStatisticalConfig()
	}
	return &StatisticalDetector{
		BaseDetector: BaseDetector{name: "StatisticalDetector"},
		config:       config,
	}
}

// Detect implements Detector interface
func (s *StatisticalDetector) Detect(ctx types.TransactionContext, profile *profiles.UserProfile) []types.AnomalySignal {
	signals := make([]types.AnomalySignal, 0)
	
	txn := ctx.Txn
	amount := txn.WithdrawalAmt
	
	if amount == 0 {
		return signals
	}
	
	// Skip small transactions (noise reduction)
	if amount < 1000.0 {
		return signals
	}
	
	category := txn.Category
	if category == "" {
		category = "Other"
	}
	
	// Get category profile
	catProfile, exists := profile.CategoryProfiles[category]
	if !exists || catProfile.Count < 5 {
		// Not enough data for category, use overall profile
		signals = append(signals, s.detectUnusualAmountOverall(amount, profile)...)
	} else {
		// Use category-specific statistics
		signals = append(signals, s.detectUnusualAmountCategory(amount, catProfile, category)...)
	}
	
		// Detect unusual merchant
		signals = append(signals, s.detectUnusualMerchant(ctx, profile)...)
	
	// Detect spending spike
	signals = append(signals, s.detectSpendingSpike(amount, profile)...)
	
	return signals
}

// detectUnusualAmountCategory detects unusual amounts using category statistics
func (s *StatisticalDetector) detectUnusualAmountCategory(amount float64, catProfile *profiles.CategoryProfile, category string) []types.AnomalySignal {
	signals := make([]types.AnomalySignal, 0)
	
	// Method 1: Z-Score
	var zScore float64
	if catProfile.StdDev > 0 {
		zScore = (amount - catProfile.Mean) / catProfile.StdDev
	}
	
	// Method 2: IQR-based (Tukey's method)
	upperBound := catProfile.Q3 + (s.config.IQRMultiplier * catProfile.IQR)
	lowerBound := catProfile.Q1 - (s.config.IQRMultiplier * catProfile.IQR)
	isIQROutlier := amount > upperBound || amount < lowerBound
	
	// Method 3: Percentile-based
	isP99Outlier := s.config.UseP99 && amount > catProfile.P99
	isP95Outlier := s.config.UseP95 && amount > catProfile.P95
	
	// Determine severity and score
	absZScore := math.Abs(zScore)
	
	if isP99Outlier || absZScore >= s.config.ZScoreCritical || amount > upperBound*2 {
		score := 95.0
		if isP99Outlier {
			score = 92.0
		}
		signals = append(signals, types.NewSignal(
			types.SignalUnusualAmount,
			types.CategoryAmount,
			score,
			formatUnusualAmount(amount, catProfile, "exceeds 99th percentile", zScore),
		))
	} else if isP95Outlier || absZScore >= s.config.ZScoreHigh || amount > upperBound*1.5 {
		score := 80.0
		if isP95Outlier {
			score = 75.0
		}
		signals = append(signals, types.NewSignal(
			types.SignalUnusualAmount,
			types.CategoryAmount,
			score,
			formatUnusualAmount(amount, catProfile, "exceeds 95th percentile", zScore),
		))
	} else if isIQROutlier || absZScore >= s.config.ZScoreMedium {
		signals = append(signals, types.NewSignal(
			types.SignalUnusualAmount,
			types.CategoryAmount,
			60.0,
			formatUnusualAmount(amount, catProfile, "significantly above typical range", zScore),
		))
	} else if absZScore >= s.config.ZScoreLow {
		signals = append(signals, types.NewSignal(
			types.SignalUnusualAmount,
			types.CategoryAmount,
			40.0,
			formatUnusualAmount(amount, catProfile, "above average", zScore),
		))
	}
	
	return signals
}

// detectUnusualAmountOverall detects unusual amounts using overall profile
func (s *StatisticalDetector) detectUnusualAmountOverall(amount float64, profile *profiles.UserProfile) []types.AnomalySignal {
	signals := make([]types.AnomalySignal, 0)
	
	if profile.StdDevTxnAmount == 0 {
		return signals
	}
	
	zScore := (amount - profile.AvgTxnAmount) / profile.StdDevTxnAmount
	absZScore := math.Abs(zScore)
	
	if absZScore >= s.config.ZScoreCritical || amount > profile.P99Amount {
		signals = append(signals, types.NewSignal(
			types.SignalUnusualAmount,
			types.CategoryAmount,
			90.0,
			formatUnusualAmountOverall(amount, profile, zScore),
		))
	} else if absZScore >= s.config.ZScoreHigh || amount > profile.P95Amount {
		signals = append(signals, types.NewSignal(
			types.SignalUnusualAmount,
			types.CategoryAmount,
			75.0,
			formatUnusualAmountOverall(amount, profile, zScore),
		))
	}
	
	return signals
}

// detectUnusualMerchant detects first-time or rare merchants
func (s *StatisticalDetector) detectUnusualMerchant(ctx types.TransactionContext, profile *profiles.UserProfile) []types.AnomalySignal {
	signals := make([]types.AnomalySignal, 0)
	
	merchant := strings.ToUpper(strings.TrimSpace(ctx.Txn.Merchant))
	if merchant == "" || merchant == "UNKNOWN" {
		return signals
	}
	
	frequency, exists := profile.KnownMerchants[merchant]
	amount := ctx.Txn.WithdrawalAmt
	
	// First-time merchant with large amount
	if !exists && amount > profile.AvgDailySpend*3 {
		signals = append(signals, types.NewSignal(
			types.SignalNewMerchant,
			types.CategoryMerchant,
			65.0,
			formatNewMerchant(amount, profile.AvgDailySpend),
		))
	} else if frequency == 1 && amount > profile.AvgDailySpend*2 {
		signals = append(signals, types.NewSignal(
			types.SignalRareMerchant,
			types.CategoryMerchant,
			45.0,
			formatRareMerchant(amount, profile.AvgDailySpend),
		))
	}
	
	return signals
}

// detectSpendingSpike detects sudden spending spikes
func (s *StatisticalDetector) detectSpendingSpike(amount float64, profile *profiles.UserProfile) []types.AnomalySignal {
	signals := make([]types.AnomalySignal, 0)
	
	if profile.AvgDailySpend == 0 {
		return signals
	}
	
	expected3DaySpend := profile.AvgDailySpend * 3
	
	// If single transaction exceeds 2x expected 3-day spend
	if amount > expected3DaySpend*2 {
		ratio := amount / expected3DaySpend
		score := math.Min(85.0, 60.0 + (ratio-2.0)*10)
		
		signals = append(signals, types.NewSignal(
			types.SignalSpendingSpike,
			types.CategoryBehavior,
			score,
			formatSpendingSpike(amount, expected3DaySpend, ratio),
		))
	}
	
	return signals
}

// Formatting helpers

// User-friendly explanations (Zerodha/HDFC style - no "anomaly" word)
func formatUnusualAmount(amount float64, profile *profiles.CategoryProfile, reason string, zScore float64) string {
	// Friendly explanation with comparison context
	if strings.Contains(reason, "99th percentile") {
		return fmt.Sprintf("significantly higher than usual for this category (top 1%% of your transactions)")
	} else if strings.Contains(reason, "95th percentile") {
		return fmt.Sprintf("higher than usual for this category (top 5%% of your transactions)")
	}
	return fmt.Sprintf("unusual for this category compared to your spending history")
}

func formatUnusualAmountOverall(amount float64, profile *profiles.UserProfile, zScore float64) string {
	// Friendly explanation with context
	if zScore >= 3.0 {
		return fmt.Sprintf("significantly higher than your average transaction (%.1f× your typical amount)", 
			amount/profile.AvgTxnAmount)
	}
	return fmt.Sprintf("higher than your average transaction (%.1f× your typical amount)", 
		amount/profile.AvgTxnAmount)
}

func formatNewMerchant(amount, avgDaily float64) string {
	ratio := amount / avgDaily
	return fmt.Sprintf("This is %.1f× your daily average spending", ratio)
}

func formatRareMerchant(amount, avgDaily float64) string {
	ratio := amount / avgDaily
	return fmt.Sprintf("This is %.1f× your daily average spending", ratio)
}

func formatSpendingSpike(amount, expected3Day, ratio float64) string {
	return fmt.Sprintf("This single transaction equals %.1f× your typical 3-day spending", ratio)
}

