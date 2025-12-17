package analyzer

// Config holds configuration for the analyzer
type Config struct {
	// Thresholds
	BigTicketThreshold float64 // Threshold for big ticket movements
	FraudAlertThreshold float64 // Threshold for fraud alerts

	// Limits
	TopBeneficiariesLimit int // Number of top beneficiaries to return
	TopExpensesLimit      int // Number of top expenses to return

	// Options
	EnablePredictiveInsights bool // Enable predictive analytics
	EnableTaxInsights        bool // Enable tax insights
	EnableFraudDetection     bool // Enable fraud detection
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		BigTicketThreshold:       20000,
		FraudAlertThreshold:      50000,
		TopBeneficiariesLimit:    5,
		TopExpensesLimit:         5,
		EnablePredictiveInsights: true,
		EnableTaxInsights:        true,
		EnableFraudDetection:     true,
	}
}

