package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// NormalizeNarrationForFingerprint normalizes narration by removing dates, numbers, and reference IDs
// This creates a stable fingerprint for recurring payment detection
// Example: "ACH D STAFF LOAN EMI REC 12-DEC-2024" → "ACH D STAFF LOAN EMI REC"
func NormalizeNarrationForFingerprint(narration string) string {
	if narration == "" {
		return ""
	}

	// Convert to uppercase for consistency
	normalized := strings.ToUpper(strings.TrimSpace(narration))

	// Remove dates (various formats)
	// Pattern: DD-MM-YYYY, DD/MM/YYYY, DD-MMM-YYYY, etc.
	datePatterns := []*regexp.Regexp{
		regexp.MustCompile(`\d{1,2}[-/]\d{1,2}[-/]\d{2,4}`),           // DD-MM-YYYY or DD/MM/YYYY
		regexp.MustCompile(`\d{1,2}[-/]\w{3}[-/]\d{2,4}`),            // DD-MMM-YYYY
		regexp.MustCompile(`\w{3}\s+\d{1,2},?\s+\d{4}`),              // MMM DD, YYYY
		regexp.MustCompile(`\d{4}[-/]\d{1,2}[-/]\d{1,2}`),            // YYYY-MM-DD
	}

	for _, pattern := range datePatterns {
		normalized = pattern.ReplaceAllString(normalized, "")
	}

	// Remove reference numbers and IDs (long sequences of digits)
	// Pattern: sequences of 8+ digits (likely reference numbers)
	refPattern := regexp.MustCompile(`\d{8,}`)
	normalized = refPattern.ReplaceAllString(normalized, "")

	// Remove account numbers (masked: XXXXXXXXXXXX1234)
	maskedAccountPattern := regexp.MustCompile(`X{6,}\d{4}`)
	normalized = maskedAccountPattern.ReplaceAllString(normalized, "")

	// Remove transaction IDs (alphanumeric codes)
	// Pattern: sequences like A54152, K16675, etc.
	txnIdPattern := regexp.MustCompile(`\b[A-Z]\d{4,}\b`)
	normalized = txnIdPattern.ReplaceAllString(normalized, "")

	// Clean up multiple spaces
	spacePattern := regexp.MustCompile(`\s+`)
	normalized = spacePattern.ReplaceAllString(normalized, " ")

	// Trim and return
	return strings.TrimSpace(normalized)
}

// FingerprintNarration creates a hash fingerprint of normalized narration
// This is used to identify transactions with same narration pattern
func FingerprintNarration(narration string) string {
	normalized := NormalizeNarrationForFingerprint(narration)
	if normalized == "" {
		return ""
	}

	// Create SHA256 hash
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// MatchNarrationFingerprints checks if two narrations have the same fingerprint
func MatchNarrationFingerprints(narration1, narration2 string) bool {
	fp1 := FingerprintNarration(narration1)
	fp2 := FingerprintNarration(narration2)
	return fp1 != "" && fp1 == fp2
}

