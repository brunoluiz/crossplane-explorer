package xplane

import (
	"encoding/json"
	"fmt"
	"io"
)

// Parse a stream into a crossplane resource (usually from stdin or os.Exec).
func Parse(r io.Reader) (*Resource, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}

	var data *Resource
	if err := json.Unmarshal(input, &data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return data, nil
}
