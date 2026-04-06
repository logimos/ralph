package layers

import "strings"

// parseCommandLine splits a command string into the executable name and remaining arguments.
// Double-quoted segments preserve spaces (e.g. node "/path with spaces/main.js").
func parseCommandLine(s string) (name string, rest []string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	var parts []string
	var b strings.Builder
	inQuote := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuote = !inQuote
			continue
		}
		if !inQuote && c == ' ' {
			if b.Len() > 0 {
				parts = append(parts, b.String())
				b.Reset()
			}
			continue
		}
		b.WriteByte(c)
	}
	if b.Len() > 0 {
		parts = append(parts, b.String())
	}
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}
