package detectors

// Note: ML detector is a stub for future implementation
// It doesn't import anomaly_engine to avoid import cycles
// Types will be passed via interface{} or function wrappers

// MLDetector is a stub for future ML-based anomaly detection
// This allows plugging in ML models later without refactoring
type MLDetector struct {
	BaseDetector
	enabled bool
	// Future: model interface
	// model MLModel
}

// NewMLDetector creates a new ML detector (disabled by default)
func NewMLDetector(enabled bool) *MLDetector {
	return &MLDetector{
		BaseDetector: BaseDetector{name: "MLDetector"},
		enabled:      enabled,
	}
}

// Detect implements Detector interface
// Currently returns empty (disabled) - ready for ML integration
// Note: Uses interface{} to avoid import cycle
func (m *MLDetector) Detect(ctx interface{}, profile interface{}) []interface{} {
	if !m.enabled {
		return []interface{}{}
	}
	
	// TODO: Future ML integration
	// 1. Extract features from transaction context
	// 2. Call ML model (ONNX, TensorFlow Lite, or custom Go model)
	// 3. Convert ML output to AnomalySignal
	
	// Example future implementation:
	// ctxTyped := ctx.(anomaly_engine.TransactionContext)
	// profileTyped := profile.(*profiles.UserProfile)
	// features := extractMLFeatures(ctxTyped, profileTyped)
	// mlScore := m.model.Predict(features)
	// if mlScore > 0.8 {
	//     return []interface{}{
	//         anomaly_engine.NewSignal(
	//             anomaly_engine.SignalMLAnomaly,
	//             anomaly_engine.CategoryML,
	//             mlScore * 100,
	//             "Transaction deviates from your historical behavior pattern",
	//         ),
	//     }
	// }
	
	return []interface{}{}
}

// extractMLFeatures extracts features for ML model (future)
// This shows how features would be extracted for ML
// Note: Uses interface{} to avoid import cycle
func extractMLFeatures(ctx interface{}, profile interface{}) map[string]float64 {
	features := make(map[string]float64)
	
	// Type assertions (avoiding import cycle)
	// In real implementation, these would be proper types
	// ctxTyped := ctx.(anomaly_engine.TransactionContext)
	// profileTyped := profile.(*profiles.UserProfile)
	
	// Example feature extraction (commented for now):
	// txn := ctxTyped.Txn
	// amount := txn.WithdrawalAmt
	// features["amount_normalized"] = amount / profileTyped.AvgTxnAmount
	// ... etc
	
	return features
}

// MLModel interface for future ML integration
type MLModel interface {
	// Predict returns anomaly score (0-1)
	Predict(features map[string]float64) float64
}

