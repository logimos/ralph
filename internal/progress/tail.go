// Package progress provides helpers for bounded progress file handling.
package progress

import "unicode/utf8"

// UTF8Tail returns the last maxBytes bytes of data, adjusted so the result is valid UTF-8
// (does not start mid-rune). If maxBytes <= 0 or len(data) <= maxBytes, returns a copy of data.
func UTF8Tail(data []byte, maxBytes int) []byte {
	if maxBytes <= 0 || len(data) <= maxBytes {
		out := make([]byte, len(data))
		copy(out, data)
		return out
	}
	raw := data[len(data)-maxBytes:]
	for len(raw) > 0 && !utf8.Valid(raw) {
		raw = raw[1:]
	}
	out := make([]byte, len(raw))
	copy(out, raw)
	return out
}
