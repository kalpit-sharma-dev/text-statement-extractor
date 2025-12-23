package pagination

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// GenerateETag generates an ETag from pagination result data.
// ETags are used for HTTP cache validation.
//
// Example:
//   etag := GenerateETag(result)
//   w.Header().Set("ETag", etag)
func GenerateETag[T any](result PaginationResult[T]) (string, error) {
	// Serialize the result to JSON
	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	// Generate SHA256 hash
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	// Return ETag with weak validator prefix (W/)
	// Weak validators allow semantic equivalence
	return fmt.Sprintf(`W/"%s"`, hashStr[:32]), nil
}

// ValidateETag validates an ETag against pagination result data.
//
// Example:
//   if ValidateETag(etag, result) {
//       w.WriteHeader(http.StatusNotModified)
//       return
//   }
func ValidateETag[T any](etag string, result PaginationResult[T]) (bool, error) {
	generatedETag, err := GenerateETag(result)
	if err != nil {
		return false, err
	}

	// Remove quotes and weak validator prefix for comparison
	normalizeETag := func(s string) string {
		s = strings.Trim(s, `"`)
		if strings.HasPrefix(s, "W/") {
			s = s[2:]
		}
		return s
	}

	return normalizeETag(etag) == normalizeETag(generatedETag), nil
}

// GenerateETagFromData generates an ETag from raw data.
func GenerateETagFromData(data []byte) string {
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])
	return fmt.Sprintf(`W/"%s"`, hashStr[:32])
}

