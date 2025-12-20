package detectors

// Detector is the interface all anomaly detectors must implement
// This allows pluggable detection: rules, statistics, ML, etc.
// Note: Types are defined in individual detector files to avoid import cycles
type Detector interface {
	// Detect analyzes a transaction and returns anomaly signals
	// Types are defined in each detector's file
	Detect(ctx interface{}, profile interface{}) []interface{}
	
	// Name returns the detector's name for logging/debugging
	Name() string
}

// BaseDetector provides common functionality for all detectors
type BaseDetector struct {
	name string
}

func (b *BaseDetector) Name() string {
	return b.name
}

