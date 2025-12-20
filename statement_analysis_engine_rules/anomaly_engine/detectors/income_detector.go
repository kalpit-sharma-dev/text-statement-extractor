package detectors

import (
	"fmt"
	"time"

	"classify/statement_analysis_engine_rules/anomaly_engine/profiles"
	"classify/statement_analysis_engine_rules/anomaly_engine/types"
	"classify/statement_analysis_engine_rules/models"
)

// IncomeDetector detects income disruption and anomalies
type IncomeDetector struct {
	BaseDetector
	config *IncomeConfig
	history []models.ClassifiedTransaction
}

// IncomeConfig holds income detection configuration
type IncomeConfig struct {
	// Income disruption thresholds
	IncomeDropThreshold    float64 // Percentage drop to flag (default: 70%)
	MinDaysForAlert        int     // Minimum days in month to alert (default: 15)
	ExpectedSalaryDays     []int   // Expected salary days (e.g., [1, 2, 3, 25, 26, 27])
}

// DefaultIncomeConfig returns default income detection config
func DefaultIncomeConfig() *IncomeConfig {
	return &IncomeConfig{
		IncomeDropThreshold: 0.70, // 70% drop
		MinDaysForAlert:     15,   // Alert if 15+ days passed
		ExpectedSalaryDays:   []int{1, 2, 3, 25, 26, 27, 28, 29, 30, 31}, // Common salary days
	}
}

// NewIncomeDetector creates a new income detector
func NewIncomeDetector(config *IncomeConfig, history []models.ClassifiedTransaction) *IncomeDetector {
	if config == nil {
		config = DefaultIncomeConfig()
	}
	return &IncomeDetector{
		BaseDetector: BaseDetector{name: "IncomeDetector"},
		config:       config,
		history:      history,
	}
}

// Detect implements Detector interface
func (i *IncomeDetector) Detect(ctx types.TransactionContext, profile *profiles.UserProfile) []types.AnomalySignal {
	signals := make([]types.AnomalySignal, 0)
	
	// Only check income/credit transactions
	if ctx.Txn.DepositAmt == 0 {
		return signals
	}
	
	// Detect income disruption
	signals = append(signals, i.detectIncomeDisruption(ctx, profile)...)
	
	return signals
}

// detectIncomeDisruption detects if salary/income is missing or significantly reduced
func (i *IncomeDetector) detectIncomeDisruption(ctx types.TransactionContext, profile *profiles.UserProfile) []types.AnomalySignal {
	signals := make([]types.AnomalySignal, 0)
	
	txn := ctx.Txn
	currentDate := ctx.Timestamp
	
	// Only check income transactions
	if txn.DepositAmt == 0 {
		return signals
	}
	
	// Calculate current month income
	currentMonthIncome := i.calculateCurrentMonthIncome(currentDate)
	
	// Calculate average monthly income from history
	avgMonthlyIncome := i.calculateAverageMonthlyIncome()
	
	if avgMonthlyIncome == 0 {
		return signals // Not enough history
	}
	
	// Check if current month income is significantly lower
	daysInMonth := currentDate.Day()
	expectedIncomeByNow := (avgMonthlyIncome / 30.0) * float64(daysInMonth)
	
	// If we're past minimum days and income is way below expected
	if daysInMonth >= i.config.MinDaysForAlert {
		incomeRatio := currentMonthIncome / expectedIncomeByNow
		
		if incomeRatio < i.config.IncomeDropThreshold {
			score := 80.0
			if incomeRatio < 0.1 {
				score = 90.0 // Critical if < 10% of expected
			}
			
			explanation := fmt.Sprintf("Income for this month (₹%.0f) is significantly lower than expected (₹%.0f). This is %.0f%% of your typical monthly income. Please verify if salary has been received.", 
				currentMonthIncome, expectedIncomeByNow, incomeRatio*100)
			
			signals = append(signals, types.NewSignal(
				"INCOME_DISRUPTION",
				types.CategoryBehavior,
				score,
				explanation,
			))
		}
	}
	
	return signals
}

// calculateCurrentMonthIncome calculates total income for current month
func (i *IncomeDetector) calculateCurrentMonthIncome(currentDate time.Time) float64 {
	var total float64
	
	currentYear, currentMonth := currentDate.Year(), currentDate.Month()
	
	for _, txn := range i.history {
		if txn.DepositAmt == 0 {
			continue
		}
		
		// Only count income, not refunds
		if txn.Category == "Refund" {
			continue
		}
		
		txnDate, err := parseTransactionDateForPattern(txn.Date)
		if err != nil {
			continue
		}
		
		if txnDate.Year() == currentYear && txnDate.Month() == currentMonth {
			total += txn.DepositAmt
		}
	}
	
	return total
}

// calculateAverageMonthlyIncome calculates average monthly income from history
func (i *IncomeDetector) calculateAverageMonthlyIncome() float64 {
	if len(i.history) == 0 {
		return 0
	}
	
	// Group income by month
	monthlyIncome := make(map[string]float64)
	
	for _, txn := range i.history {
		if txn.DepositAmt == 0 {
			continue
		}
		
		// Only count income, not refunds
		if txn.Category == "Refund" {
			continue
		}
		
		txnDate, err := parseTransactionDateForPattern(txn.Date)
		if err != nil {
			continue
		}
		
		monthKey := fmt.Sprintf("%d-%02d", txnDate.Year(), txnDate.Month())
		monthlyIncome[monthKey] += txn.DepositAmt
	}
	
	if len(monthlyIncome) == 0 {
		return 0
	}
	
	// Calculate average
	var total float64
	for _, income := range monthlyIncome {
		total += income
	}
	
	return total / float64(len(monthlyIncome))
}

