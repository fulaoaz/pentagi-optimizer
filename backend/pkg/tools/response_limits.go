package tools

import (
	"encoding/json"
	"fmt"
	"io"
)

const maxExternalJSONResponseBytes = 8 * 1024 * 1024

func decodeBoundedJSON(body io.Reader, target any, maxBytes int) error {
	if body == nil {
		return fmt.Errorf("response body is nil")
	}
	if maxBytes <= 0 {
		maxBytes = maxExternalJSONResponseBytes
	}

	data, err := io.ReadAll(io.LimitReader(body, int64(maxBytes)+1))
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if len(data) > maxBytes {
		return fmt.Errorf("response body exceeds the %d-byte limit", maxBytes)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}
	return nil
}
