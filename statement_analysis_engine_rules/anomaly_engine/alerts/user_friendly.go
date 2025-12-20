package alerts

import (
	"classify/statement_analysis_engine_rules/anomaly_engine/types"
	"fmt"
	"strings"
)

// UserFriendlyFormatter formats alerts in Zerodha/HDFC style (friendly, insight-driven)
// Never uses words like "anomaly", "risk", "fraud" in user-facing messages
type UserFriendlyFormatter struct {
	*Formatter
}

// NewUserFriendlyFormatter creates a new user-friendly formatter
func NewUserFriendlyFormatter() *UserFriendlyFormatter {
	return &UserFriendlyFormatter{
		Formatter: NewFormatter(),
	}
}

// Format converts AnomalyResult to user-friendly, insight-driven alert
func (f *UserFriendlyFormatter) Format(result interface{}, transactionAmount float64, merchant string, profileContext map[string]interface{}) Alert {
	// Type assertion
	anomalyResult, ok := result.(struct {
		Signals     []types.AnomalySignal
		FinalScore  float64
		Severity    types.Severity
		TopSignals  []types.AnomalySignal
		Explanation string
		Confidence  float64
	})
	if !ok || len(anomalyResult.Signals) == 0 {
		return Alert{
			Title:      "Transaction processed",
			Message:    "No unusual patterns detected in this transaction.",
			Severity:   types.SeverityInfo,
			Confidence: 1.0,
		}
	}

	topSignal := anomalyResult.TopSignals[0]

	return Alert{
		Title:      f.userFriendlyTitle(topSignal, anomalyResult.Severity),
		Message:    f.userFriendlyMessage(anomalyResult, transactionAmount, merchant, profileContext),
		Severity:   anomalyResult.Severity,
		Confidence: anomalyResult.Confidence,
		Action:     f.userFriendlyAction(anomalyResult.Severity),
	}
}

// userFriendlyTitle generates friendly, insight-driven titles (Zerodha/HDFC style)
func (f *UserFriendlyFormatter) userFriendlyTitle(signal types.AnomalySignal, severity types.Severity) string {
	// Never use "anomaly", "risk", "fraud" in titles
	
	switch severity {
	case types.SeverityCritical:
		return "💡 Large transaction alert"
	case types.SeverityHigh:
		return "💡 Spending insight"
	case types.SeverityMedium:
		return "ℹ️ Pattern change"
	case types.SeverityLow:
		return "ℹ️ Spending update"
	default:
		return "ℹ️ Transaction insight"
	}
}

// userFriendlyMessage generates friendly, explainable messages with context
func (f *UserFriendlyFormatter) userFriendlyMessage(result interface{}, amount float64, merchant string, profileContext map[string]interface{}) string {
	resultTyped := result.(struct {
		Signals     []types.AnomalySignal
		TopSignals  []types.AnomalySignal
		Severity    types.Severity
		Confidence  float64
	})
	
	if len(resultTyped.TopSignals) == 0 {
		return "No unusual patterns detected."
	}
	
	topSignal := resultTyped.TopSignals[0]
	amountStr := formatAmount(amount)
	
	var message strings.Builder
	
	// Get comparison baseline from context
	comparisonBaseline := f.getComparisonBaseline(profileContext, topSignal)
	
	// Build friendly message based on signal type
	switch topSignal.Code {
	case types.SignalHighAmount, types.SignalAmountSpike:
		message.WriteString(fmt.Sprintf("You spent %s at %s, which is %s.", 
			amountStr, merchant, comparisonBaseline))
		message.WriteString(" This is higher than your usual spending pattern.")
		
	case types.SignalUnusualAmount:
		message.WriteString(fmt.Sprintf("Transaction of %s to %s %s.", 
			amountStr, merchant, comparisonBaseline))
		message.WriteString(" This differs from your typical spending in this category.")
		
	case types.SignalNewMerchant:
		message.WriteString(fmt.Sprintf("First transaction with %s for %s.", 
			merchant, amountStr))
		message.WriteString(fmt.Sprintf(" %s", comparisonBaseline))
		
	case types.SignalRareMerchant:
		message.WriteString(fmt.Sprintf("Transaction to %s for %s.", 
			merchant, amountStr))
		message.WriteString(fmt.Sprintf(" You've used this merchant rarely. %s", comparisonBaseline))
		
	case types.SignalSpendingSpike:
		message.WriteString(fmt.Sprintf("Your spending today is %s.", comparisonBaseline))
		message.WriteString(" This is higher than your typical daily average.")
		
	case types.SignalDuplicatePayment:
		message.WriteString(fmt.Sprintf("Similar payment of %s to %s was made recently.", 
			amountStr, merchant))
		message.WriteString(" Just a heads-up in case this was unintentional.")
		
	default:
		// Generic friendly message
		message.WriteString(fmt.Sprintf("Transaction of %s to %s %s.", 
			amountStr, merchant, comparisonBaseline))
	}
	
	// Add confidence indicator only if low
	if resultTyped.Confidence < 0.7 {
		message.WriteString(fmt.Sprintf(" (Confidence: %.0f%%)", resultTyped.Confidence*100))
	}
	
	return message.String()
}

// getComparisonBaseline extracts comparison context from profile
func (f *UserFriendlyFormatter) getComparisonBaseline(profileContext map[string]interface{}, signal types.AnomalySignal) string {
	if profileContext == nil {
		return "unusual compared to your history"
	}
	
	// Try to get specific comparison data
	if avgDaily, ok := profileContext["avgDailySpend"].(float64); ok {
		if signal.Code == types.SignalSpendingSpike {
			return fmt.Sprintf("%.1f× your daily average", 
				profileContext["currentSpend"].(float64)/avgDaily)
		}
	}
	
	if avgTxn, ok := profileContext["avgTxnAmount"].(float64); ok {
		if signal.Code == types.SignalHighAmount || signal.Code == types.SignalUnusualAmount {
			if amount, ok := profileContext["amount"].(float64); ok {
				ratio := amount / avgTxn
				return fmt.Sprintf("%.1f× your average transaction", ratio)
			}
		}
	}
	
	// Default friendly comparison
	if days, ok := profileContext["historyDays"].(int); ok && days > 0 {
		return fmt.Sprintf("unusual compared to your last %d days", days)
	}
	
	return "unusual compared to your spending history"
}

// userFriendlyAction generates friendly action text
func (f *UserFriendlyFormatter) userFriendlyAction(severity types.Severity) string {
	switch severity {
	case types.SeverityCritical:
		return "Please verify if this matches your intent"
	case types.SeverityHigh:
		return "Review if this looks correct"
	case types.SeverityMedium:
		return "Just for your awareness"
	default:
		return ""
	}
}

// FormatBatch formats multiple alerts in user-friendly style
func (f *UserFriendlyFormatter) FormatBatch(results []interface{}, transactions []struct {
	Amount  float64
	Merchant string
}, profileContexts []map[string]interface{}) []Alert {
	alerts := make([]Alert, 0, len(results))
	
	for i, result := range results {
		var amount float64
		var merchant string
		var context map[string]interface{}
		
		if i < len(transactions) {
			amount = transactions[i].Amount
			merchant = transactions[i].Merchant
		}
		if i < len(profileContexts) {
			context = profileContexts[i]
		}
		
		alert := f.Format(result, amount, merchant, context)
		
		// Only include significant alerts
		if alert.Severity != types.SeverityInfo && alert.Confidence > 0.5 {
			alerts = append(alerts, alert)
		}
	}
	
	return alerts
}

