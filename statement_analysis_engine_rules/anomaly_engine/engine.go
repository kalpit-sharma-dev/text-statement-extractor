package anomaly_engine

import (
	"classify/statement_analysis_engine_rules/anomaly_engine/detectors"
	"classify/statement_analysis_engine_rules/anomaly_engine/profiles"
	"classify/statement_analysis_engine_rules/anomaly_engine/suppression"
	"classify/statement_analysis_engine_rules/anomaly_engine/types"
	"classify/statement_analysis_engine_rules/models"
)

// DetectorFunc is a function type for detectors
type DetectorFunc func(ctx types.TransactionContext, profile *profiles.UserProfile) []types.AnomalySignal

// Engine is the main anomaly detection orchestrator
// It coordinates all detectors and aggregates results
type Engine struct {
	detectors     []DetectorFunc
	detectorNames []string
	scorer        *Scorer
	profile       *profiles.UserProfile
	history       []models.ClassifiedTransaction // For duplicate detection
	suppressor    *suppression.Suppressor       // Bank-grade suppression rules
}

// EngineConfig holds engine configuration
type EngineConfig struct {
	EnableRuleDetector        bool
	EnableStatisticalDetector bool
	EnableMLDetector          bool
	EnableDuplicateDetector   bool
	EnablePatternDetector     bool // Multi-transaction pattern detection
	EnableIncomeDetector      bool // Income disruption detection
	HistorySize               int // Number of recent transactions to keep for duplicate detection
}

// DefaultEngineConfig returns default engine configuration
func DefaultEngineConfig() *EngineConfig {
	return &EngineConfig{
		EnableRuleDetector:        true,
		EnableStatisticalDetector: true,
		EnableMLDetector:          false, // Disabled by default
		EnableDuplicateDetector:   true,
		EnablePatternDetector:     true, // Enable pattern detection by default
		EnableIncomeDetector:      true, // Enable income disruption detection
		HistorySize:               100,
	}
}

// NewEngine creates a new anomaly detection engine
func NewEngine(config *EngineConfig, transactionHistory []models.ClassifiedTransaction) *Engine {
	if config == nil {
		config = DefaultEngineConfig()
	}

	engine := &Engine{
		detectors:     make([]DetectorFunc, 0),
		detectorNames: make([]string, 0),
		scorer:        NewScorer(),
		history:       transactionHistory,
		suppressor:    suppression.NewSuppressor(),
	}

	// Build user profile from history
	engine.profile = profiles.BuildUserProfile(transactionHistory)

	// Register detectors based on config (using function wrappers to avoid import cycle)
	if config.EnableRuleDetector {
		ruleDet := detectors.NewRuleDetector(nil)
		engine.detectors = append(engine.detectors, func(ctx types.TransactionContext, profile *profiles.UserProfile) []types.AnomalySignal {
			return ruleDet.Detect(ctx, profile)
		})
		engine.detectorNames = append(engine.detectorNames, "RuleDetector")
	}

	if config.EnableStatisticalDetector {
		statDet := detectors.NewStatisticalDetector(nil)
		engine.detectors = append(engine.detectors, func(ctx types.TransactionContext, profile *profiles.UserProfile) []types.AnomalySignal {
			return statDet.Detect(ctx, profile)
		})
		engine.detectorNames = append(engine.detectorNames, "StatisticalDetector")
	}

	if config.EnableMLDetector {
		mlDet := detectors.NewMLDetector(true)
		engine.detectors = append(engine.detectors, func(ctx types.TransactionContext, profile *profiles.UserProfile) []types.AnomalySignal {
			results := mlDet.Detect(ctx, profile)
			signals := make([]types.AnomalySignal, 0, len(results))
			for _, r := range results {
				if sig, ok := r.(types.AnomalySignal); ok {
					signals = append(signals, sig)
				}
			}
			return signals
		})
		engine.detectorNames = append(engine.detectorNames, "MLDetector")
	}

	if config.EnableDuplicateDetector {
		// Use recent history for duplicate detection
		historySize := config.HistorySize
		if historySize > len(transactionHistory) {
			historySize = len(transactionHistory)
		}
		recentHistory := transactionHistory
		if len(transactionHistory) > historySize {
			recentHistory = transactionHistory[len(transactionHistory)-historySize:]
		}
		dupDet := detectors.NewDuplicateDetector(nil, recentHistory)
		engine.detectors = append(engine.detectors, func(ctx types.TransactionContext, profile *profiles.UserProfile) []types.AnomalySignal {
			results := dupDet.DetectInternal(ctx.Txn, profile)
			signals := make([]types.AnomalySignal, 0, len(results))
			for _, r := range results {
				if sigMap, ok := r.(map[string]interface{}); ok {
					signals = append(signals, types.AnomalySignal{
						Code:        getString(sigMap, "Code"),
						Category:    getString(sigMap, "Category"),
						Score:       getFloat(sigMap, "Score"),
						Severity:    types.Severity(getString(sigMap, "Severity")),
						Explanation: getString(sigMap, "Explanation"),
					})
				}
			}
			return signals
		})
		engine.detectorNames = append(engine.detectorNames, "DuplicateDetector")
	}

	if config.EnablePatternDetector {
		// Use recent history for pattern detection
		historySize := config.HistorySize
		if historySize > len(transactionHistory) {
			historySize = len(transactionHistory)
		}
		recentHistory := transactionHistory
		if len(transactionHistory) > historySize {
			recentHistory = transactionHistory[len(transactionHistory)-historySize:]
		}
		patternDet := detectors.NewPatternDetector(nil, recentHistory)
		engine.detectors = append(engine.detectors, func(ctx types.TransactionContext, profile *profiles.UserProfile) []types.AnomalySignal {
			return patternDet.Detect(ctx, profile)
		})
		engine.detectorNames = append(engine.detectorNames, "PatternDetector")
	}

	if config.EnableIncomeDetector {
		// Use full history for income detection (needs monthly patterns)
		incomeDet := detectors.NewIncomeDetector(nil, transactionHistory)
		engine.detectors = append(engine.detectors, func(ctx types.TransactionContext, profile *profiles.UserProfile) []types.AnomalySignal {
			return incomeDet.Detect(ctx, profile)
		})
		engine.detectorNames = append(engine.detectorNames, "IncomeDetector")
	}

	return engine
}

// Evaluate analyzes a transaction and returns anomaly result
func (e *Engine) Evaluate(ctx types.TransactionContext) AnomalyResult {
	// Step 1: Apply suppression rules (bank-grade filtering)
	suppressionRule := e.suppressor.ShouldSuppress(ctx.Txn)
	
		// If suppressed, return empty result
	if suppressionRule.SkipAnomalyDetection {
		return AnomalyResult{
			Signals:     []types.AnomalySignal{},
			FinalScore:  0,
			Severity:    types.SeverityInfo,
			TopSignals:  []types.AnomalySignal{},
			Explanation: "Transaction follows your usual patterns",
			Confidence:  1.0,
			RiskFlags:   []string{},
		}
	}
	
	// Step 2: Collect signals from all detectors
	allSignals := make([]types.AnomalySignal, 0)

	for _, detector := range e.detectors {
		signals := detector(ctx, e.profile) // detector is a function, call it directly
		allSignals = append(allSignals, signals...)
	}

	// Step 3: Score and aggregate signals
	result := e.scorer.Score(allSignals)
	
	// Step 4: Apply severity cap if needed
	if suppressionRule.MaxSeverity != "CRITICAL" {
		result = e.applySeverityCap(result, suppressionRule.MaxSeverity)
	}

	return result
}

// applySeverityCap caps the severity at a maximum level
func (e *Engine) applySeverityCap(result AnomalyResult, maxSeverity string) AnomalyResult {
	maxSev := types.Severity(maxSeverity)
	
	// If current severity exceeds max, downgrade
	if compareSeverity(result.Severity, maxSev) > 0 {
		result.Severity = maxSev
		
		// Also cap individual signal severities
		for i := range result.Signals {
			if compareSeverity(result.Signals[i].Severity, maxSev) > 0 {
				result.Signals[i].Severity = maxSev
			}
		}
		
		for i := range result.TopSignals {
			if compareSeverity(result.TopSignals[i].Severity, maxSev) > 0 {
				result.TopSignals[i].Severity = maxSev
			}
		}
		
		// Adjust final score to match capped severity
		result.FinalScore = severityToScore(maxSev)
		
		// Update explanation (user-friendly)
		result.Explanation = "Pattern change detected (trusted/recurring transaction): " + result.Explanation
	}
	
	return result
}

// compareSeverity returns: 1 if s1 > s2, -1 if s1 < s2, 0 if equal
func compareSeverity(s1, s2 types.Severity) int {
	severityOrder := map[types.Severity]int{
		types.SeverityInfo:     0,
		types.SeverityLow:      1,
		types.SeverityMedium:   2,
		types.SeverityHigh:     3,
		types.SeverityCritical: 4,
	}
	
	order1 := severityOrder[s1]
	order2 := severityOrder[s2]
	
	if order1 > order2 {
		return 1
	} else if order1 < order2 {
		return -1
	}
	return 0
}

// severityToScore converts severity to approximate score
func severityToScore(sev types.Severity) float64 {
	switch sev {
	case types.SeverityCritical:
		return 95.0
	case types.SeverityHigh:
		return 80.0
	case types.SeverityMedium:
		return 60.0
	case types.SeverityLow:
		return 30.0
	default:
		return 10.0
	}
}

// EvaluateBatch evaluates multiple transactions
func (e *Engine) EvaluateBatch(transactions []models.ClassifiedTransaction, userID string) []AnomalyResult {
	results := make([]AnomalyResult, 0, len(transactions))

	for _, txn := range transactions {
		ctx := NewTransactionContext(txn, userID)
		result := e.Evaluate(ctx)
		results = append(results, result)
	}

	return results
}

// UpdateProfile rebuilds user profile (call after adding new transactions)
func (e *Engine) UpdateProfile(transactionHistory []models.ClassifiedTransaction) {
	e.profile = profiles.BuildUserProfile(transactionHistory)
	e.history = transactionHistory
}

// Helper functions for type conversion
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}
