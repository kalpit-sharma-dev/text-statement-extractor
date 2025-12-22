package utils

import (
	"strings"
)

// MatchNames compares two names by checking if all words in one name appear in the other
// This handles cases where:
// - Full name vs partial name (e.g., "KALPIT KUMAR SHARMA" vs "KALPIT SHARMA")
// - Different order (e.g., "KUMAR SHARMA" vs "SHARMA KUMAR")
// - Missing middle name (e.g., "KALPIT KUMAR SHARMA" vs "KALPIT SHARMA")
// Returns true if all words in shorterName appear in longerName (case-insensitive)
func MatchNames(name1, name2 string) bool {
	if name1 == "" || name2 == "" {
		return false
	}

	// Normalize names: remove common prefixes and extra spaces
	name1 = normalizeName(name1)
	name2 = normalizeName(name2)

	if name1 == "" || name2 == "" {
		return false
	}

	// Split into words
	words1 := strings.Fields(strings.ToUpper(name1))
	words2 := strings.Fields(strings.ToUpper(name2))

	// Remove common words that don't help with matching
	words1 = removeCommonWords(words1)
	words2 = removeCommonWords(words2)

	if len(words1) == 0 || len(words2) == 0 {
		return false
	}

	// Check if all words in shorter name appear in longer name
	// This handles: "KALPIT SHARMA" matching "KALPIT KUMAR SHARMA"
	if len(words1) <= len(words2) {
		return allWordsMatch(words1, words2)
	}
	return allWordsMatch(words2, words1)
}

// normalizeName removes common prefixes and normalizes the name
func normalizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToUpper(name)

	// Remove common prefixes
	prefixes := []string{"MR.", "MR ", "MRS.", "MRS ", "MS.", "MS ", "DR.", "DR ", "PROF.", "PROF "}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			name = strings.TrimPrefix(name, prefix)
			name = strings.TrimSpace(name)
			break
		}
	}

	return strings.TrimSpace(name)
}

// removeCommonWords removes common words that don't help with name matching
func removeCommonWords(words []string) []string {
	commonWords := map[string]bool{
		"AND": true,
		"THE": true,
		"OF":  true,
		"TO":  true,
		"FOR": true,
	}

	filtered := make([]string, 0, len(words))
	for _, word := range words {
		if !commonWords[word] && len(word) > 1 { // Ignore single character words
			filtered = append(filtered, word)
		}
	}
	return filtered
}

// allWordsMatch checks if all words in shorter list appear in longer list
func allWordsMatch(shorter, longer []string) bool {
	if len(shorter) == 0 {
		return false
	}

	// Create a map of words in longer name for quick lookup
	longerMap := make(map[string]bool)
	for _, word := range longer {
		longerMap[word] = true
	}

	// Check if all words in shorter name are in longer name
	for _, word := range shorter {
		if !longerMap[word] {
			return false
		}
	}

	return true
}

