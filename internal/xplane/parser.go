package xplane

import (
	"encoding/json"
	"fmt"
	"io"
)

func Parse(r io.Reader) (*Resource, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from stdin: %w", err)
	}

	var data *Resource
	if err := json.Unmarshal(input, &data); err != nil {
		return nil, fmt.Errorf("Failed to decode JSON: %w", err)
	}

	return data, nil
}

func MustParse(r io.Reader) *Resource {
	data, err := Parse(r)
	if err != nil {
		panic(err)
	}

	return data
}
