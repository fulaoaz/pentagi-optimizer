package searchers

import (
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

const (
	maxSearchResponseBytes           = 4 * 1024 * 1024
	maxSearchSummarizationInputBytes = 256 * 1024
)

// readSearchResponseBody bounds provider responses before they are parsed. A
// provider can omit Content-Length or use chunked transfer, so the limit is
// enforced while reading rather than by headers alone.
func readSearchResponseBody(body io.Reader) ([]byte, error) {
	if body == nil {
		return nil, fmt.Errorf("search response body is nil")
	}

	data, err := io.ReadAll(io.LimitReader(body, int64(maxSearchResponseBytes)+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if len(data) > maxSearchResponseBytes {
		return nil, fmt.Errorf("search response body exceeds the %d-byte limit", maxSearchResponseBytes)
	}
	return data, nil
}

func decodeSearchResponseBody(body io.Reader, target any) error {
	data, err := readSearchResponseBody(body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}
	return nil
}

func truncateSearchText(value string, maxBytes int) string {
	if maxBytes <= 0 || len(value) <= maxBytes {
		return value
	}

	cutoff := maxBytes
	for cutoff > 0 && cutoff < len(value) && !utf8.RuneStart(value[cutoff]) {
		cutoff--
	}
	return value[:cutoff]
}

func boundSearchSummarizationInput(value string) string {
	return truncateSearchText(value, maxSearchSummarizationInputBytes)
}
