package detectors

import (
	"fmt"
	"math"
	"strings"
	"time"

	"classify/statement_analysis_engine_rules/anomaly_engine/profiles"
	"classify/statement_analysis_engine_rules/models"
)

// DuplicateDetector detects duplicate payments
type DuplicateDetector struct {
	BaseDetector
	config *DuplicateConfig
	history []models.ClassifiedTransaction // Recent transaction history
}

// DuplicateConfig holds duplicate detection configuration
type DuplicateConfig struct {
	LookbackWindow    int     // Number of transactions to look back
	TimeWindowDays    float64 // Time window in days
	AmountTolerance   float64 // Amount difference tolerance (percentage)
}

// DefaultDuplicateConfig returns default duplicate detection config
func DefaultDuplicateConfig() *DuplicateConfig {
	return &DuplicateConfig{
		LookbackWindow:  50,
		TimeWindowDays:  3.0,
		AmountTolerance: 0.01, // 1% tolerance
	}
}

// NewDuplicateDetector creates a new duplicate detector
func NewDuplicateDetector(config *DuplicateConfig, history []models.ClassifiedTransaction) *DuplicateDetector {
	if config == nil {
		config = DefaultDuplicateConfig()
	}
	return &DuplicateDetector{
		BaseDetector: BaseDetector{name: "DuplicateDetector"},
		config:       config,
		history:      history,
	}
}

// Detect implements Detector interface
// Note: We use models.ClassifiedTransaction directly to avoid import cycle
func (d *DuplicateDetector) Detect(ctx interface{}, profile interface{}) []interface{} {
	// Type assertions (avoiding import cycle)
	txnCtx, ok := ctx.(struct {
		Txn models.ClassifiedTransaction
	})
	if !ok {
		return []interface{}{}
	}
	
	prof, ok := profile.(*profiles.UserProfile)
	if !ok {
		return []interface{}{}
	}
	
	return d.DetectInternal(txnCtx.Txn, prof)
}

// DetectInternal does the actual detection (called via wrapper from engine.go)
func (d *DuplicateDetector) DetectInternal(txn models.ClassifiedTransaction, profile *profiles.UserProfile) []interface{} {
	signals := make([]interface{}, 0)
	
	amount := txn.WithdrawalAmt
	
	if amount == 0 || len(d.history) == 0 {
		return signals
	}
	
	// Look through recent history
	lookbackLimit := d.config.LookbackWindow
	if lookbackLimit > len(d.history) {
		lookbackLimit = len(d.history)
	}
	
	txnDate, err := parseTransactionDate(txn.Date)
	if err != nil {
		return signals
	}
	
	merchant := strings.ToUpper(strings.TrimSpace(txn.Merchant))
	
	for i := len(d.history) - 1; i >= 0 && i >= len(d.history)-lookbackLimit; i-- {
		other := d.history[i]
		
		if other.WithdrawalAmt == 0 {
			continue
		}
		
		// Check same merchant
		otherMerchant := strings.ToUpper(strings.TrimSpace(other.Merchant))
		if merchant == "" || otherMerchant == "" || merchant != otherMerchant {
			continue
		}
		
		// Check amount similarity (within tolerance)
		amountDiff := math.Abs(amount - other.WithdrawalAmt)
		amountRatio := amountDiff / amount
		if amountRatio > d.config.AmountTolerance {
			continue
		}
		
		// Check time window
		otherDate, err := parseTransactionDate(other.Date)
		if err != nil {
			continue
		}
		
		daysDiff := math.Abs(txnDate.Sub(otherDate).Hours() / 24)
		if daysDiff > d.config.TimeWindowDays {
			continue
		}
		
		// Duplicate detected!
		var score float64
		var explanation string
		
		if daysDiff < 1 {
			// Same day - high severity
			score = 85.0
			explanation = formatDuplicateSameDay(amount, merchant, other.Date)
		} else {
			// Within time window - medium severity
			score = 60.0
			explanation = formatDuplicateWithinWindow(amount, merchant, daysDiff)
		}
		
		// Create signal as map (will be converted in engine.go)
		signal := map[string]interface{}{
			"Code":        "DUPLICATE_PAYMENT",
			"Category":    "Frequency",
			"Score":       score,
			"Severity":    getSeverityFromScore(score),
			"Explanation": explanation,
		}
		signals = append(signals, signal)
		
		// Only report first duplicate found
		break
	}
	
	return signals
}

func getSeverityFromScore(score float64) string {
	switch {
	case score >= 90:
		return "CRITICAL"
	case score >= 70:
		return "HIGH"
	case score >= 45:
		return "MEDIUM"
	case score >= 20:
		return "LOW"
	default:
		return "INFO"
	}
}

func parseTransactionDate(dateStr string) (time.Time, error) {
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

// User-friendly explanations (Zerodha/HDFC style)
func formatDuplicateSameDay(amount float64, merchant string, otherDate string) string {
	return fmt.Sprintf("Similar payment to %s was made earlier today (%s). Just a heads-up in case this was unintentional.", 
		merchant, otherDate)
}

func formatDuplicateWithinWindow(amount float64, merchant string, daysDiff float64) string {
	days := int(daysDiff)
	if days == 0 {
		return fmt.Sprintf("Similar payment to %s was made within 24 hours. Just a heads-up in case this was unintentional.", 
			merchant)
	}
	return fmt.Sprintf("Similar payment to %s was made %d days ago. Just a heads-up in case this was unintentional.", 
		merchant, days)
}

