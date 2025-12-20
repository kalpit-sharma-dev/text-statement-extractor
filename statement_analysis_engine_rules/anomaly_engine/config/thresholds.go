package config

// Thresholds holds all configurable thresholds for anomaly detection
// These can be tuned based on your data and requirements
type Thresholds struct {
	// Amount thresholds
	LargeAmountThreshold    float64
	VeryLargeAmountThreshold float64
	
	// Statistical thresholds
	ZScoreCritical float64
	ZScoreHigh     float64
	ZScoreMedium   float64
	ZScoreLow      float64
	
	// IQR multiplier
	IQRMultiplier float64
	
	// Duplicate detection
	DuplicateTimeWindowDays float64
	DuplicateAmountTolerance float64
	
	// Round amount detection
	RoundAmountMin float64
	
	// Merchant thresholds
	UnknownMerchantThreshold float64
	NewMerchantMultiplier    float64 // Multiplier of daily average
}

// DefaultThresholds returns default threshold values
// These are industry-standard values used by banks
// Tuned to reduce false positives while maintaining detection accuracy
func DefaultThresholds() *Thresholds {
	return &Thresholds{
		LargeAmountThreshold:     50000,
		VeryLargeAmountThreshold:  100000,
		ZScoreCritical:           4.5,  // Increased from 4.0 to reduce false positives
		ZScoreHigh:               3.5,  // Increased from 3.0
		ZScoreMedium:             3.0,  // Increased from 2.5
		ZScoreLow:                2.5,  // Increased from 2.0
		IQRMultiplier:            1.5, // Tukey's method
		DuplicateTimeWindowDays:  3.0,
		DuplicateAmountTolerance: 0.01, // 1%
		RoundAmountMin:           50000, // Increased from 25000 to reduce noise
		UnknownMerchantThreshold: 20000, // Increased from 10000
		NewMerchantMultiplier:    5.0,  // Increased from 3.0 to 5x daily average
	}
}

// ConservativeThresholds returns more conservative (stricter) thresholds
// Use this to reduce false positives
func ConservativeThresholds() *Thresholds {
	t := DefaultThresholds()
	t.ZScoreCritical = 4.5
	t.ZScoreHigh = 3.5
	t.ZScoreMedium = 3.0
	t.ZScoreLow = 2.5
	t.LargeAmountThreshold = 75000
	t.VeryLargeAmountThreshold = 150000
	return t
}

// AggressiveThresholds returns more aggressive (looser) thresholds
// Use this to catch more anomalies (may have more false positives)
func AggressiveThresholds() *Thresholds {
	t := DefaultThresholds()
	t.ZScoreCritical = 3.5
	t.ZScoreHigh = 2.5
	t.ZScoreMedium = 2.0
	t.ZScoreLow = 1.5
	t.LargeAmountThreshold = 30000
	t.VeryLargeAmountThreshold = 75000
	return t
}

