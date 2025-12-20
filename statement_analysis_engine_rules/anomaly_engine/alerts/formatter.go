package alerts

import (
	"classify/statement_analysis_engine_rules/anomaly_engine/types"
	"fmt"
	"strings"
)

// Alert represents a formatted, user-friendly alert (Zerodha/HDFC style)
type Alert struct {
	Title      string       `json:"title"`
	Message    string       `json:"message"`
	Severity   types.Severity `json:"severity"`
	Confidence float64      `json:"confidence"` // 0-1
	Action     string       `json:"action,omitempty"` // Optional action text
}

// Formatter formats anomaly results into user-friendly alerts
type Formatter struct{}

// NewFormatter creates a new alert formatter
func NewFormatter() *Formatter {
	return &Formatter{}
}

// Format converts AnomalyResult to user-friendly Alert
func (f *Formatter) Format(result interface{}, transactionAmount float64, merchant string) Alert {
	// Type assertion (avoiding import cycle)
	anomalyResult, ok := result.(struct {
		Signals     []types.AnomalySignal
		FinalScore  float64
		Severity    types.Severity
		TopSignals  []types.AnomalySignal
		Explanation string
		Confidence  float64
	})
	if !ok {
		return Alert{
			Title:      "Transaction verified",
			Message:    "No anomalies detected.",
			Severity:   types.SeverityInfo,
			Confidence: 1.0,
		}
	}
	if len(anomalyResult.Signals) == 0 {
		return Alert{
			Title:      "Transaction verified",
			Message:    "No anomalies detected in this transaction.",
			Severity:   types.SeverityInfo,
			Confidence: 1.0,
		}
	}
	
	topSignal := anomalyResult.TopSignals[0]
	
	return Alert{
		Title:      f.alertTitle(topSignal, anomalyResult.Severity),
		Message:    f.alertMessage(anomalyResult, transactionAmount, merchant),
		Severity:   anomalyResult.Severity,
		Confidence: anomalyResult.Confidence,
		Action:     f.alertAction(anomalyResult.Severity),
	}
}

// alertTitle generates alert title based on severity and signal
func (f *Formatter) alertTitle(signal types.AnomalySignal, severity types.Severity) string {
	switch severity {
	case types.SeverityCritical:
		return "⚠️ Unusual transaction detected"
	case types.SeverityHigh:
		return "⚠️ High-risk transaction"
	case types.SeverityMedium:
		return "ℹ️ Transaction alert"
	case types.SeverityLow:
		return "ℹ️ Spending pattern change"
	default:
		return "ℹ️ Transaction notification"
	}
}

// alertMessage generates human-readable message (Zerodha/HDFC style)
func (f *Formatter) alertMessage(result interface{}, amount float64, merchant string) string {
	resultTyped := result.(struct {
		Signals     []types.AnomalySignal
		TopSignals  []types.AnomalySignal
		Severity    types.Severity
		Confidence  float64
	})
	
	if len(resultTyped.TopSignals) == 0 {
		return "No issues detected with this transaction."
	}
	
	topSignal := resultTyped.TopSignals[0]
	
	// Build message based on signal type and severity
	var message strings.Builder
	
	// Amount formatting
	amountStr := formatAmount(amount)
	
	// Severity-based messaging
	switch resultTyped.Severity {
	case types.SeverityCritical:
		message.WriteString(fmt.Sprintf("⚠️ %s spent at %s, which is significantly higher than your usual spending.", amountStr, merchant))
		message.WriteString(" If this wasn't you, please contact the bank immediately.")
		
	case types.SeverityHigh:
		message.WriteString(fmt.Sprintf("₹%s spent at %s, which is unusually high compared to your typical transactions.", formatAmountPlain(amount), merchant))
		message.WriteString(" Please verify this transaction.")
		
	case types.SeverityMedium:
		message.WriteString(fmt.Sprintf("Transaction of %s to %s differs from your usual spending pattern.", amountStr, merchant))
		message.WriteString(" " + topSignal.Explanation)
		
	case types.SeverityLow:
		message.WriteString(fmt.Sprintf("Your spending today is higher than your daily average."))
		if topSignal.Explanation != "" {
			message.WriteString(" " + topSignal.Explanation)
		}
		
	default:
		message.WriteString(topSignal.Explanation)
	}
	
	// Add confidence indicator if not high
	if resultTyped.Confidence < 0.8 {
		message.WriteString(fmt.Sprintf(" (Confidence: %.0f%%)", resultTyped.Confidence*100))
	}
	
	return message.String()
}

// alertAction generates action text based on severity
func (f *Formatter) alertAction(severity types.Severity) string {
	switch severity {
	case types.SeverityCritical:
		return "Contact bank immediately if unauthorized"
	case types.SeverityHigh:
		return "Verify transaction"
	case types.SeverityMedium:
		return "Review transaction"
	default:
		return ""
	}
}

// FormatBatch formats multiple alerts
func (f *Formatter) FormatBatch(results []interface{}, transactions []struct {
	Amount  float64
	Merchant string
}) []Alert {
	alerts := make([]Alert, 0, len(results))
	
	for i, result := range results {
		var amount float64
		var merchant string
		
		if i < len(transactions) {
			amount = transactions[i].Amount
			merchant = transactions[i].Merchant
		}
		
		alert := f.Format(result, amount, merchant)
		alerts = append(alerts, alert)
	}
	
	return alerts
}

// Helper functions

func formatAmount(amount float64) string {
	if amount >= 100000 {
		lakhs := amount / 100000
		if lakhs >= 10 {
			crores := lakhs / 100
			return fmt.Sprintf("₹%.1fCr", crores)
		}
		return fmt.Sprintf("₹%.1fL", lakhs)
	}
	if amount >= 1000 {
		thousands := amount / 1000
		return fmt.Sprintf("₹%.1fK", thousands)
	}
	return fmt.Sprintf("₹%.0f", amount)
}

func formatAmountPlain(amount float64) string {
	if amount >= 100000 {
		return fmt.Sprintf("%.0f", amount)
	}
	return fmt.Sprintf("%.0f", amount)
}

