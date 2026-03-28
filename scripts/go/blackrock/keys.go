package main

import (
	"strings"
	"unicode"
)

// Same rules as scripts/go/statestreet/main.go so JSON keys match when the
// underlying datapoint matches State Street’s three tables (Fund / Index / Fund Market Price).
var keyRenames = map[string]string{
	"price_earnings_ratio_fy1": "price_earnings_fw",
}

func normalizeKey(s string) string {
	var b strings.Builder
	lastUnderscore := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			lastUnderscore = false
		} else if !lastUnderscore {
			b.WriteRune('_')
			lastUnderscore = true
		}
	}
	out := strings.Trim(b.String(), "_")
	if renamed, ok := keyRenames[out]; ok {
		return renamed
	}
	return out
}
