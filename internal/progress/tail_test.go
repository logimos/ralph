package progress

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"
)

func TestUTF8Tail_empty(t *testing.T) {
	got := UTF8Tail(nil, 100)
	if len(got) != 0 {
		t.Fatalf("got %q", got)
	}
}

func TestUTF8Tail_noTrim(t *testing.T) {
	data := []byte("hello")
	got := UTF8Tail(data, 100)
	if !bytes.Equal(got, data) {
		t.Fatalf("got %q want %q", got, data)
	}
}

func TestUTF8Tail_ascii(t *testing.T) {
	data := []byte("0123456789")
	got := UTF8Tail(data, 4)
	if string(got) != "6789" {
		t.Fatalf("got %q", got)
	}
}

func TestUTF8Tail_multibyteBoundary(t *testing.T) {
	// "€" = 3 bytes; take last 2 bytes of string → invalid; should drop until valid
	s := []byte("abc€xyz")
	// Last 4 bytes might cut € — tail should skip to valid UTF-8
	got := UTF8Tail(s, 4)
	if !utf8.Valid(got) {
		t.Fatalf("invalid UTF-8: %q", got)
	}
}

func TestReadUTF8TailFromFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "p.txt")
	content := []byte("0123456789abcdefghij")
	if err := os.WriteFile(p, content, 0644); err != nil {
		t.Fatal(err)
	}
	got, err := ReadUTF8TailFromFile(p, 5)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "fghij" {
		t.Fatalf("got %q", got)
	}
}
