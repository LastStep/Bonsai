package redact

// The long random run: 32 or more of [A-Za-z0-9_+/=-] taken out whole when the run, or a piece of 16 or more of it
// between those separators, looks random. It runs last, over what the other rules left. It keeps pure hex (git SHAs,
// GUIDs), `toolu_` and `srvtoolu_` ids, and CamelCase names with two digits or fewer and no two capitals in a row.

import "strings"

func isRunByte(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' ||
		c == '_' || c == '-' || c == '+' || c == '/' || c == '='
}

func isPieceSep(c byte) bool { return c == '_' || c == '-' || c == '+' || c == '/' || c == '=' }

func findRandom(s string, add func(a, b int)) {
	for i := 0; i < len(s); {
		if !isRunByte(s[i]) {
			i++
			continue
		}
		j := i
		for j < len(s) && isRunByte(s[j]) {
			j++
		}
		if j-i >= 32 && randomRun(s[i:j]) {
			add(i, j)
		}
		i = j
	}
}

func randomRun(run string) bool {
	if strings.HasPrefix(run, "toolu_") || strings.HasPrefix(run, "srvtoolu_") {
		return false
	}
	for start, k := 0, 0; k <= len(run); k++ {
		if k == len(run) || isPieceSep(run[k]) {
			if looksRandom(run[start:k]) {
				return true
			}
			start = k + 1
		}
	}
	return false
}

// looksRandom: 16 or more characters, with a capital, a small letter and a digit, not pure hex, and not a CamelCase
// name (two digits or fewer and no two capitals in a row).
func looksRandom(p string) bool {
	if len(p) < 16 {
		return false
	}
	upper, lower, digits, hex, twoCaps := false, false, 0, true, false
	for i := 0; i < len(p); i++ {
		c := p[i]
		switch {
		case c >= 'A' && c <= 'Z':
			upper = true
			if i > 0 && p[i-1] >= 'A' && p[i-1] <= 'Z' {
				twoCaps = true
			}
		case c >= 'a' && c <= 'z':
			lower = true
		case c >= '0' && c <= '9':
			digits++
		}
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
			hex = false
		}
	}
	if !upper || !lower || digits == 0 || hex {
		return false
	}
	return digits > 2 || twoCaps
}
