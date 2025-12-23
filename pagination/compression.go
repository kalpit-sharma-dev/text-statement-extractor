package pagination

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
)

// CompressResult compresses a pagination result using gzip.
// Returns the compressed data and the original size.
//
// Example:
//   compressed, originalSize, err := CompressResult(result)
//   w.Header().Set("Content-Encoding", "gzip")
//   w.Write(compressed)
func CompressResult[T any](result PaginationResult[T]) ([]byte, int, error) {
	// Marshal to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal result: %w", err)
	}

	originalSize := len(jsonData)

	// Compress using gzip
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write(jsonData); err != nil {
		return nil, 0, fmt.Errorf("failed to write compressed data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, 0, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), originalSize, nil
}

// DecompressResult decompresses gzip-compressed pagination result data.
func DecompressResult[T any](compressedData []byte) (PaginationResult[T], error) {
	var result PaginationResult[T]

	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return result, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return result, fmt.Errorf("failed to decompress data: %w", err)
	}

	if err := json.Unmarshal(decompressed, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal decompressed data: %w", err)
	}

	return result, nil
}

// CompressionRatio calculates the compression ratio.
func CompressionRatio(originalSize, compressedSize int) float64 {
	if originalSize == 0 {
		return 0
	}
	return float64(compressedSize) / float64(originalSize)
}

