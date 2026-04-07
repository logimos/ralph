// Package progress provides helpers for bounded progress file handling.
package progress

import (
	"io"
	"os"
	"unicode/utf8"
)

// UTF8AlignStart returns a suffix of b that is valid UTF-8, skipping at most utf8.UTFMax-1
// leading bytes (mid-rune cut). If no alignment works, returns b unchanged (best effort).
func UTF8AlignStart(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	lim := utf8.UTFMax
	if len(b) < lim {
		lim = len(b)
	}
	for i := 0; i < lim; i++ {
		if utf8.Valid(b[i:]) {
			return b[i:]
		}
	}
	return b
}

// UTF8Tail returns the last maxBytes bytes of data, adjusted so the result is valid UTF-8
// (does not start mid-rune). If maxBytes <= 0 or len(data) <= maxBytes, returns a copy of data.
func UTF8Tail(data []byte, maxBytes int) []byte {
	if maxBytes <= 0 || len(data) <= maxBytes {
		out := make([]byte, len(data))
		copy(out, data)
		return out
	}
	raw := data[len(data)-maxBytes:]
	aligned := UTF8AlignStart(raw)
	out := make([]byte, len(aligned))
	copy(out, aligned)
	return out
}

// ReadUTF8TailFromFile reads at most maxBytes from the end of path without loading the whole file.
// Returns nil, nil if the file is missing or empty. The returned slice is a copy.
func ReadUTF8TailFromFile(path string, maxBytes int) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := st.Size()
	if size == 0 {
		return nil, nil
	}

	n := int64(maxBytes)
	if n > size {
		n = size
	}
	if _, err := f.Seek(size-n, io.SeekStart); err != nil {
		return nil, err
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(f, buf); err != nil {
		return nil, err
	}
	aligned := UTF8AlignStart(buf)
	out := make([]byte, len(aligned))
	copy(out, aligned)
	return out, nil
}
