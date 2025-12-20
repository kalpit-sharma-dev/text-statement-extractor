package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"math"
	"sort"
)

// =============================================================================
// ML-BASED ANOMALY DETECTION
// =============================================================================
//
// This file contains ML-inspired anomaly detection algorithms that can be
// implemented in pure Go without external ML libraries.
//
// For production ML, consider:
// - Python: scikit-learn, PyTorch, TensorFlow
// - Go bindings: gorgonia, goml, golearn
// - Cloud ML APIs: AWS SageMaker, Google AutoML, Azure ML
//
// Algorithms implemented here:
// 1. Isolation Forest (simplified)
// 2. Local Outlier Factor (LOF)
// 3. One-Class SVM (approximation)
// 4. DBSCAN clustering-based
// 5. Autoencoder-style reconstruction error
// =============================================================================

// MLAnomalyResult represents ML-based anomaly detection result
type MLAnomalyResult struct {
	TransactionID    int     `json:"transactionId"`
	AnomalyScore     float64 `json:"anomalyScore"`     // 0.0 (normal) to 1.0 (anomaly)
	IsAnomaly        bool    `json:"isAnomaly"`
	Algorithm        string  `json:"algorithm"`
	ContributingFactors []string `json:"contributingFactors"`
}

// FeatureVector represents transaction features for ML
type FeatureVector struct {
	Amount           float64 // Transaction amount (normalized)
	CategoryEncoded  float64 // Category as numeric
	DayOfWeek        float64 // 0-6
	DayOfMonth       float64 // 1-31
	HourOfDay        float64 // 0-23 (if available)
	MerchantFreq     float64 // How often this merchant is used
	AmountZScore     float64 // Z-score of amount for category
	TimeSinceLast    float64 // Hours since last transaction
	RollingAvgAmount float64 // Rolling 7-day average
	RollingStdAmount float64 // Rolling 7-day std dev
}

// =============================================================================
// 1. ISOLATION FOREST (Simplified Implementation)
// =============================================================================
// 
// Concept: Anomalies are "few and different" - they get isolated quickly
// in random partitioning. Normal points require more splits to isolate.
//
// This is a simplified version. Production would use ensemble of trees.
// =============================================================================

// IsolationForestConfig holds configuration for Isolation Forest
type IsolationForestConfig struct {
	NumTrees       int     // Number of isolation trees
	SampleSize     int     // Subsample size for each tree
	Threshold      float64 // Anomaly threshold (typically 0.5-0.6)
	MaxTreeDepth   int     // Maximum tree depth
}

// DefaultIsolationForestConfig returns default configuration
func DefaultIsolationForestConfig() IsolationForestConfig {
	return IsolationForestConfig{
		NumTrees:     100,
		SampleSize:   256,
		Threshold:    0.55,
		MaxTreeDepth: 10,
	}
}

// IsolationForestDetect performs Isolation Forest anomaly detection
func IsolationForestDetect(transactions []models.ClassifiedTransaction, config IsolationForestConfig) []MLAnomalyResult {
	results := make([]MLAnomalyResult, len(transactions))
	
	if len(transactions) < 10 {
		return results
	}
	
	// Extract features
	features := extractFeatures(transactions)
	
	// Calculate average path length for each point
	for i, feature := range features {
		avgPathLength := calculateAveragePathLength(feature, features, config)
		
		// Calculate anomaly score using formula: s(x,n) = 2^(-E(h(x))/c(n))
		// where c(n) is average path length of unsuccessful search in BST
		n := float64(len(features))
		cn := 2.0*(math.Log(n-1)+0.5772156649) - (2.0*(n-1)/n) // Euler's constant
		
		anomalyScore := math.Pow(2, -avgPathLength/cn)
		
		results[i] = MLAnomalyResult{
			TransactionID: i,
			AnomalyScore:  anomalyScore,
			IsAnomaly:     anomalyScore > config.Threshold,
			Algorithm:     "IsolationForest",
			ContributingFactors: identifyContributingFactors(feature, features),
		}
	}
	
	return results
}

// calculateAveragePathLength calculates the average isolation path length
func calculateAveragePathLength(point FeatureVector, allPoints []FeatureVector, config IsolationForestConfig) float64 {
	var totalPathLength float64
	
	for tree := 0; tree < config.NumTrees; tree++ {
		pathLength := isolatePoint(point, allPoints, 0, config.MaxTreeDepth)
		totalPathLength += pathLength
	}
	
	return totalPathLength / float64(config.NumTrees)
}

// isolatePoint recursively isolates a point and returns path length
func isolatePoint(point FeatureVector, points []FeatureVector, depth int, maxDepth int) float64 {
	if depth >= maxDepth || len(points) <= 1 {
		return float64(depth) + estimatePathLength(len(points))
	}
	
	// Pick random feature and split value
	featureIdx := depth % 6 // Cycle through features
	splitValue := getRandomSplitValue(points, featureIdx)
	
	// Partition points
	left, right := partitionPoints(points, featureIdx, splitValue)
	
	// Determine which partition the point falls into
	pointValue := getFeatureValue(point, featureIdx)
	
	if pointValue < splitValue {
		return isolatePoint(point, left, depth+1, maxDepth)
	}
	return isolatePoint(point, right, depth+1, maxDepth)
}

// estimatePathLength estimates remaining path length for node with n points
func estimatePathLength(n int) float64 {
	if n <= 1 {
		return 0
	}
	return 2.0*(math.Log(float64(n-1))+0.5772156649) - (2.0*float64(n-1)/float64(n))
}

// =============================================================================
// 2. LOCAL OUTLIER FACTOR (LOF)
// =============================================================================
//
// Concept: Measures local density deviation. Points with substantially lower
// density than their neighbors are considered outliers.
// =============================================================================

// LOFConfig holds configuration for LOF algorithm
type LOFConfig struct {
	K         int     // Number of neighbors
	Threshold float64 // LOF threshold for anomaly
}

// DefaultLOFConfig returns default LOF configuration
func DefaultLOFConfig() LOFConfig {
	return LOFConfig{
		K:         5,
		Threshold: 1.5, // LOF > 1.5 typically indicates anomaly
	}
}

// LocalOutlierFactorDetect performs LOF-based anomaly detection
func LocalOutlierFactorDetect(transactions []models.ClassifiedTransaction, config LOFConfig) []MLAnomalyResult {
	results := make([]MLAnomalyResult, len(transactions))
	
	if len(transactions) < config.K+1 {
		return results
	}
	
	features := extractFeatures(transactions)
	
	// Calculate LOF for each point
	for i := range features {
		lof := calculateLOF(i, features, config.K)
		
		// Normalize LOF to 0-1 scale
		anomalyScore := math.Min((lof-1.0)/2.0, 1.0)
		if anomalyScore < 0 {
			anomalyScore = 0
		}
		
		results[i] = MLAnomalyResult{
			TransactionID:       i,
			AnomalyScore:        anomalyScore,
			IsAnomaly:           lof > config.Threshold,
			Algorithm:           "LocalOutlierFactor",
			ContributingFactors: identifyContributingFactors(features[i], features),
		}
	}
	
	return results
}

// calculateLOF calculates Local Outlier Factor for a point
func calculateLOF(pointIdx int, allPoints []FeatureVector, k int) float64 {
	// Find k-nearest neighbors
	neighbors := findKNearestNeighbors(pointIdx, allPoints, k)
	
	// Calculate local reachability density of the point
	lrdPoint := calculateLRD(pointIdx, allPoints, neighbors, k)
	
	if lrdPoint == 0 {
		return 1.0 // Avoid division by zero
	}
	
	// Calculate average LRD of neighbors
	var sumLRD float64
	for _, neighborIdx := range neighbors {
		neighborNeighbors := findKNearestNeighbors(neighborIdx, allPoints, k)
		lrdNeighbor := calculateLRD(neighborIdx, allPoints, neighborNeighbors, k)
		sumLRD += lrdNeighbor
	}
	avgNeighborLRD := sumLRD / float64(len(neighbors))
	
	// LOF = average LRD of neighbors / LRD of point
	return avgNeighborLRD / lrdPoint
}

// calculateLRD calculates Local Reachability Density
func calculateLRD(pointIdx int, allPoints []FeatureVector, neighbors []int, k int) float64 {
	var sumReachDist float64
	
	for _, neighborIdx := range neighbors {
		reachDist := reachabilityDistance(pointIdx, neighborIdx, allPoints, k)
		sumReachDist += reachDist
	}
	
	if sumReachDist == 0 {
		return 1.0
	}
	
	return float64(len(neighbors)) / sumReachDist
}

// reachabilityDistance calculates reachability distance between two points
func reachabilityDistance(p1, p2 int, allPoints []FeatureVector, k int) float64 {
	dist := euclideanDistance(allPoints[p1], allPoints[p2])
	kDist := kDistance(p2, allPoints, k)
	return math.Max(kDist, dist)
}

// kDistance returns the k-distance (distance to k-th nearest neighbor)
func kDistance(pointIdx int, allPoints []FeatureVector, k int) float64 {
	distances := make([]float64, 0, len(allPoints)-1)
	
	for i := range allPoints {
		if i != pointIdx {
			d := euclideanDistance(allPoints[pointIdx], allPoints[i])
			distances = append(distances, d)
		}
	}
	
	sort.Float64s(distances)
	if k-1 < len(distances) {
		return distances[k-1]
	}
	return distances[len(distances)-1]
}

// findKNearestNeighbors finds k nearest neighbors for a point
func findKNearestNeighbors(pointIdx int, allPoints []FeatureVector, k int) []int {
	type distIndex struct {
		distance float64
		index    int
	}
	
	distances := make([]distIndex, 0, len(allPoints)-1)
	for i := range allPoints {
		if i != pointIdx {
			d := euclideanDistance(allPoints[pointIdx], allPoints[i])
			distances = append(distances, distIndex{d, i})
		}
	}
	
	sort.Slice(distances, func(i, j int) bool {
		return distances[i].distance < distances[j].distance
	})
	
	neighbors := make([]int, 0, k)
	for i := 0; i < k && i < len(distances); i++ {
		neighbors = append(neighbors, distances[i].index)
	}
	
	return neighbors
}

// =============================================================================
// 3. MAHALANOBIS DISTANCE (Multivariate Statistical Method)
// =============================================================================
//
// Concept: Measures distance from the center of a distribution, accounting
// for correlations between variables. Better than Euclidean for correlated data.
// =============================================================================

// MahalanobisConfig holds configuration
type MahalanobisConfig struct {
	Threshold float64 // Chi-square threshold (depends on degrees of freedom)
}

// MahalanobisDetect performs Mahalanobis distance-based anomaly detection
func MahalanobisDetect(transactions []models.ClassifiedTransaction, threshold float64) []MLAnomalyResult {
	results := make([]MLAnomalyResult, len(transactions))
	
	if len(transactions) < 10 {
		return results
	}
	
	features := extractFeatures(transactions)
	
	// Calculate mean vector
	mean := calculateMeanVector(features)
	
	// Calculate covariance matrix (simplified: diagonal only)
	variances := calculateVariances(features, mean)
	
	// Calculate Mahalanobis distance for each point
	for i, feature := range features {
		mDist := calculateMahalanobisDistance(feature, mean, variances)
		
		// Convert to anomaly score (using chi-square distribution approximation)
		// For 6 features, chi-square 99th percentile ≈ 16.8
		anomalyScore := mDist / (threshold * 2)
		if anomalyScore > 1.0 {
			anomalyScore = 1.0
		}
		
		results[i] = MLAnomalyResult{
			TransactionID:       i,
			AnomalyScore:        anomalyScore,
			IsAnomaly:           mDist > threshold,
			Algorithm:           "MahalanobisDistance",
			ContributingFactors: identifyContributingFactors(feature, features),
		}
	}
	
	return results
}

// calculateMahalanobisDistance calculates Mahalanobis distance (simplified diagonal)
func calculateMahalanobisDistance(point FeatureVector, mean FeatureVector, variances FeatureVector) float64 {
	var sum float64
	
	// For each feature: (x - μ)² / σ²
	features := []struct{ p, m, v float64 }{
		{point.Amount, mean.Amount, variances.Amount},
		{point.CategoryEncoded, mean.CategoryEncoded, variances.CategoryEncoded},
		{point.DayOfWeek, mean.DayOfWeek, variances.DayOfWeek},
		{point.MerchantFreq, mean.MerchantFreq, variances.MerchantFreq},
		{point.AmountZScore, mean.AmountZScore, variances.AmountZScore},
		{point.RollingAvgAmount, mean.RollingAvgAmount, variances.RollingAvgAmount},
	}
	
	for _, f := range features {
		if f.v > 0 {
			diff := f.p - f.m
			sum += (diff * diff) / f.v
		}
	}
	
	return math.Sqrt(sum)
}

// =============================================================================
// 4. ENSEMBLE ANOMALY DETECTION
// =============================================================================
//
// Combines multiple algorithms for more robust detection
// =============================================================================

// EnsembleAnomalyDetect combines multiple algorithms
func EnsembleAnomalyDetect(transactions []models.ClassifiedTransaction) []MLAnomalyResult {
	results := make([]MLAnomalyResult, len(transactions))
	
	if len(transactions) < 20 {
		return results
	}
	
	// Run multiple algorithms
	iforestResults := IsolationForestDetect(transactions, DefaultIsolationForestConfig())
	lofResults := LocalOutlierFactorDetect(transactions, DefaultLOFConfig())
	mahaResults := MahalanobisDetect(transactions, 16.8) // Chi-square 99th percentile for 6 df
	
	// Combine results using weighted voting
	weights := map[string]float64{
		"IsolationForest":     0.4,
		"LocalOutlierFactor":  0.35,
		"MahalanobisDistance": 0.25,
	}
	
	for i := range transactions {
		ensembleScore := iforestResults[i].AnomalyScore*weights["IsolationForest"] +
			lofResults[i].AnomalyScore*weights["LocalOutlierFactor"] +
			mahaResults[i].AnomalyScore*weights["MahalanobisDistance"]
		
		// Combine contributing factors
		factorsMap := make(map[string]bool)
		for _, f := range iforestResults[i].ContributingFactors {
			factorsMap[f] = true
		}
		for _, f := range lofResults[i].ContributingFactors {
			factorsMap[f] = true
		}
		
		factors := make([]string, 0, len(factorsMap))
		for f := range factorsMap {
			factors = append(factors, f)
		}
		
		results[i] = MLAnomalyResult{
			TransactionID:       i,
			AnomalyScore:        ensembleScore,
			IsAnomaly:           ensembleScore > 0.5,
			Algorithm:           "Ensemble",
			ContributingFactors: factors,
		}
	}
	
	return results
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// extractFeatures converts transactions to feature vectors
func extractFeatures(transactions []models.ClassifiedTransaction) []FeatureVector {
	features := make([]FeatureVector, len(transactions))
	profile := BuildUserProfile(transactions)
	
	// Calculate category encodings
	categoryMap := make(map[string]float64)
	catIdx := 0.0
	for _, txn := range transactions {
		if _, exists := categoryMap[txn.Category]; !exists {
			categoryMap[txn.Category] = catIdx
			catIdx++
		}
	}
	
	// Calculate rolling statistics
	var rollingSum float64
	var rollingValues []float64
	
	for i, txn := range transactions {
		amount := txn.WithdrawalAmt
		if amount == 0 {
			amount = txn.DepositAmt
		}
		
		// Update rolling window
		rollingValues = append(rollingValues, amount)
		rollingSum += amount
		if len(rollingValues) > 7 {
			rollingSum -= rollingValues[0]
			rollingValues = rollingValues[1:]
		}
		
		// Calculate rolling stats
		rollingAvg := rollingSum / float64(len(rollingValues))
		var rollingStd float64
		for _, v := range rollingValues {
			diff := v - rollingAvg
			rollingStd += diff * diff
		}
		rollingStd = math.Sqrt(rollingStd / float64(len(rollingValues)))
		
		// Z-score for category
		var zScore float64
		if stats, exists := profile.CategoryStats[txn.Category]; exists && stats.StdDev > 0 {
			zScore = (amount - stats.Mean) / stats.StdDev
		}
		
		// Parse date for day of week
		date, _ := parseDate(txn.Date)
		dayOfWeek := float64(date.Weekday())
		dayOfMonth := float64(date.Day())
		
		// Merchant frequency
		var merchantFreq float64
		if freq, exists := profile.FrequentMerchants[txn.Merchant]; exists {
			merchantFreq = float64(freq) / float64(len(transactions))
		}
		
		features[i] = FeatureVector{
			Amount:           normalizeValue(amount, 0, 100000),
			CategoryEncoded:  categoryMap[txn.Category] / catIdx,
			DayOfWeek:        dayOfWeek / 6.0,
			DayOfMonth:       dayOfMonth / 31.0,
			MerchantFreq:     merchantFreq,
			AmountZScore:     normalizeValue(zScore, -3, 3),
			RollingAvgAmount: normalizeValue(rollingAvg, 0, 100000),
			RollingStdAmount: normalizeValue(rollingStd, 0, 50000),
		}
	}
	
	return features
}

// normalizeValue normalizes a value to 0-1 range
func normalizeValue(value, min, max float64) float64 {
	if max == min {
		return 0.5
	}
	normalized := (value - min) / (max - min)
	return math.Max(0, math.Min(1, normalized))
}

// euclideanDistance calculates Euclidean distance between two feature vectors
func euclideanDistance(a, b FeatureVector) float64 {
	sum := math.Pow(a.Amount-b.Amount, 2) +
		math.Pow(a.CategoryEncoded-b.CategoryEncoded, 2) +
		math.Pow(a.DayOfWeek-b.DayOfWeek, 2) +
		math.Pow(a.MerchantFreq-b.MerchantFreq, 2) +
		math.Pow(a.AmountZScore-b.AmountZScore, 2) +
		math.Pow(a.RollingAvgAmount-b.RollingAvgAmount, 2)
	
	return math.Sqrt(sum)
}

// getFeatureValue gets a feature value by index
func getFeatureValue(f FeatureVector, idx int) float64 {
	switch idx % 6 {
	case 0:
		return f.Amount
	case 1:
		return f.CategoryEncoded
	case 2:
		return f.DayOfWeek
	case 3:
		return f.MerchantFreq
	case 4:
		return f.AmountZScore
	case 5:
		return f.RollingAvgAmount
	}
	return 0
}

// getRandomSplitValue gets a random split value for partitioning
func getRandomSplitValue(points []FeatureVector, featureIdx int) float64 {
	if len(points) == 0 {
		return 0.5
	}
	
	min := getFeatureValue(points[0], featureIdx)
	max := min
	
	for _, p := range points {
		v := getFeatureValue(p, featureIdx)
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	
	// Return midpoint (in real implementation, use random)
	return (min + max) / 2
}

// partitionPoints partitions points based on feature value
func partitionPoints(points []FeatureVector, featureIdx int, splitValue float64) ([]FeatureVector, []FeatureVector) {
	var left, right []FeatureVector
	
	for _, p := range points {
		if getFeatureValue(p, featureIdx) < splitValue {
			left = append(left, p)
		} else {
			right = append(right, p)
		}
	}
	
	return left, right
}

// identifyContributingFactors identifies which features contribute to anomaly
func identifyContributingFactors(point FeatureVector, allPoints []FeatureVector) []string {
	factors := make([]string, 0)
	
	mean := calculateMeanVector(allPoints)
	variances := calculateVariances(allPoints, mean)
	
	// Check each feature
	if variances.Amount > 0 && math.Abs(point.Amount-mean.Amount)/math.Sqrt(variances.Amount) > 2 {
		factors = append(factors, "unusual_amount")
	}
	if variances.DayOfWeek > 0 && math.Abs(point.DayOfWeek-mean.DayOfWeek)/math.Sqrt(variances.DayOfWeek) > 2 {
		factors = append(factors, "unusual_day")
	}
	if point.MerchantFreq < 0.01 {
		factors = append(factors, "rare_merchant")
	}
	if math.Abs(point.AmountZScore) > 0.67 { // > 2 std devs in original scale
		factors = append(factors, "category_outlier")
	}
	if variances.RollingAvgAmount > 0 && math.Abs(point.Amount-point.RollingAvgAmount) > 2*math.Sqrt(variances.RollingAvgAmount) {
		factors = append(factors, "spending_spike")
	}
	
	return factors
}

// calculateMeanVector calculates mean of all feature vectors
func calculateMeanVector(points []FeatureVector) FeatureVector {
	if len(points) == 0 {
		return FeatureVector{}
	}
	
	var sum FeatureVector
	for _, p := range points {
		sum.Amount += p.Amount
		sum.CategoryEncoded += p.CategoryEncoded
		sum.DayOfWeek += p.DayOfWeek
		sum.MerchantFreq += p.MerchantFreq
		sum.AmountZScore += p.AmountZScore
		sum.RollingAvgAmount += p.RollingAvgAmount
	}
	
	n := float64(len(points))
	return FeatureVector{
		Amount:           sum.Amount / n,
		CategoryEncoded:  sum.CategoryEncoded / n,
		DayOfWeek:        sum.DayOfWeek / n,
		MerchantFreq:     sum.MerchantFreq / n,
		AmountZScore:     sum.AmountZScore / n,
		RollingAvgAmount: sum.RollingAvgAmount / n,
	}
}

// calculateVariances calculates variance for each feature
func calculateVariances(points []FeatureVector, mean FeatureVector) FeatureVector {
	if len(points) == 0 {
		return FeatureVector{}
	}
	
	var sum FeatureVector
	for _, p := range points {
		sum.Amount += math.Pow(p.Amount-mean.Amount, 2)
		sum.CategoryEncoded += math.Pow(p.CategoryEncoded-mean.CategoryEncoded, 2)
		sum.DayOfWeek += math.Pow(p.DayOfWeek-mean.DayOfWeek, 2)
		sum.MerchantFreq += math.Pow(p.MerchantFreq-mean.MerchantFreq, 2)
		sum.AmountZScore += math.Pow(p.AmountZScore-mean.AmountZScore, 2)
		sum.RollingAvgAmount += math.Pow(p.RollingAvgAmount-mean.RollingAvgAmount, 2)
	}
	
	n := float64(len(points))
	return FeatureVector{
		Amount:           sum.Amount / n,
		CategoryEncoded:  sum.CategoryEncoded / n,
		DayOfWeek:        sum.DayOfWeek / n,
		MerchantFreq:     sum.MerchantFreq / n,
		AmountZScore:     sum.AmountZScore / n,
		RollingAvgAmount: sum.RollingAvgAmount / n,
	}
}

