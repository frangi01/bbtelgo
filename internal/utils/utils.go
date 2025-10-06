package utils

import (
	"bytes"
	"encoding/json"
)

func JSON(v any, pretty bool, escapeHTML bool) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(escapeHTML)
	if pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	b := bytes.TrimRight(buf.Bytes(), "\n")
	return b, nil
}