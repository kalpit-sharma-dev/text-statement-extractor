package analytics

import (
	"classify/statement_analysis_engine_rules/anomaly_engine"
	"classify/statement_analysis_engine_rules/anomaly_engine/alerts"
	"classify/statement_analysis_engine_rules/anomaly_engine/profiles"
	"classify/statement_analysis_engine_rules/models"
)

// CalculateAnomalyDetectionWithEngine uses the new anomaly_engine package
// This integrates the bank-grade anomaly detection engine
func CalculateAnomalyDetectionWithEngine(transactions []models.ClassifiedTransaction, userID string) models.AnomalyDetection {
	if len(transactions) < 10 {
		return models.AnomalyDetection{
			Anomalies:    make([]models.AnomalyDetail, 0),
			TotalChecked: len(transactions),
			Summary: models.AnomalySummary{
				ByType:     make(map[string]int),
				BySeverity: make(map[string]int),
			},
		}
	}

	// Create engine with default config
	engineConfig := anomaly_engine.DefaultEngineConfig()
	engine := anomaly_engine.NewEngine(engineConfig, transactions)

	// Evaluate all transactions
	results := engine.EvaluateBatch(transactions, userID)

	// Convert to response format
	return convertAnomalyResultsToResponse(results, transactions)
}

// convertAnomalyResultsToResponse converts engine results to API response format
func convertAnomalyResultsToResponse(results []anomaly_engine.AnomalyResult, transactions []models.ClassifiedTransaction) models.AnomalyDetection {
	response := models.AnomalyDetection{
		Anomalies:    make([]models.AnomalyDetail, 0),
		TotalChecked: len(transactions),
		Summary: models.AnomalySummary{
			ByType:     make(map[string]int),
			BySeverity: make(map[string]int),
		},
	}

	// Collect all anomalies
	for i, result := range results {
		if result.FinalScore == 0 || len(result.Signals) == 0 {
			continue
		}

		// Create anomaly detail for each significant signal
		for _, signal := range result.TopSignals {
			if signal.Score < 20 {
				continue // Skip low-score signals
			}

			txn := transactions[i]
			detail := models.AnomalyDetail{
				TransactionID:    i,
				Type:             signal.Code,
				Severity:         string(signal.Severity),
				Score:            signal.Score,
				Description:      signal.Explanation,
				Amount:           txn.WithdrawalAmt,
				Merchant:         txn.Merchant,
				Category:         txn.Category,
				Date:             txn.Date,
				Reason:           signal.Explanation,
				StatisticalValue: signal.Score,
			}

			response.Anomalies = append(response.Anomalies, detail)
			response.Summary.ByType[signal.Code]++
			response.Summary.BySeverity[string(signal.Severity)]++
		}

		// Update risk score (use highest)
		if result.FinalScore > response.RiskScore {
			response.RiskScore = result.FinalScore
		}
	}

	response.AnomalyCount = len(response.Anomalies)

	// Get top anomalies by score
	if len(response.Anomalies) > 5 {
		// Sort by score and take top 5
		sorted := make([]models.AnomalyDetail, len(response.Anomalies))
		copy(sorted, response.Anomalies)
		// Simple sort by score (descending)
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i].Score < sorted[j].Score {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		response.Summary.TopAnomalies = sorted[:5]
	} else {
		response.Summary.TopAnomalies = response.Anomalies
	}

	return response
}

// GenerateAnomalyAlerts generates user-friendly alerts (Zerodha/HDFC style)
// Uses UserFriendlyFormatter for insight-driven, non-alarming messages
func GenerateAnomalyAlerts(transactions []models.ClassifiedTransaction, userID string) []alerts.Alert {
	if len(transactions) < 10 {
		return []alerts.Alert{}
	}

	// Create engine
	engineConfig := anomaly_engine.DefaultEngineConfig()
	engine := anomaly_engine.NewEngine(engineConfig, transactions)

	// Build profile for context
	profile := profiles.BuildUserProfile(transactions)

	// Evaluate all transactions
	results := engine.EvaluateBatch(transactions, userID)

	// Use user-friendly formatter (Zerodha/HDFC style)
	formatter := alerts.NewUserFriendlyFormatter()
	
	alertData := make([]struct {
		Amount  float64
		Merchant string
	}, len(transactions))
	
	profileContexts := make([]map[string]interface{}, len(transactions))

	for i, txn := range transactions {
		alertData[i].Amount = txn.WithdrawalAmt
		alertData[i].Merchant = txn.Merchant
		
		// Build context for comparison baseline
		profileContexts[i] = map[string]interface{}{
			"amount":        txn.WithdrawalAmt,
			"avgTxnAmount":  profile.AvgTxnAmount,
			"avgDailySpend": profile.AvgDailySpend,
			"historyDays":   30, // Default to 30 days
		}
	}

	// Convert results to interface{} slice for FormatBatch
	resultsInterface := make([]interface{}, len(results))
	for i := range results {
		resultsInterface[i] = results[i]
	}
	
	// Use user-friendly formatter
	alertsList := formatter.FormatBatch(resultsInterface, alertData, profileContexts)

	return alertsList
}

