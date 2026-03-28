package main

import (
	"strconv"
	"strings"
)

// applyNAVPriceExtras sets base_currency and price (float64) from the fund header NAV
// string, e.g. "USD 12.34" or "$12.34".
func applyNAVPriceExtras(out map[string]any) {
	raw, ok := out["nav"].(string)
	if !ok {
		return
	}
	cur, price, ok := parseDisplayMoney(raw)
	if !ok {
		return
	}
	out["base_currency"] = cur
	out["price"] = price
}

func parseDisplayMoney(display string) (currency string, price float64, ok bool) {
	s := strings.TrimSpace(display)
	if s == "" || s == "-" {
		return "", 0, false
	}

	if cur, f, ok := parseISOPrefixAmount(s); ok {
		return cur, f, true
	}
	if cur, f, ok := parseAmountISOTrailing(s); ok {
		return cur, f, true
	}
	if cur, f, ok := parseSymbolPrefixAmount(s); ok {
		return cur, f, true
	}
	return "", 0, false
}

func parseISOPrefixAmount(s string) (string, float64, bool) {
	if len(s) < 5 {
		return "", 0, false
	}
	if !isASCIIAlpha(s[0]) || !isASCIIAlpha(s[1]) || !isASCIIAlpha(s[2]) || s[3] != ' ' {
		return "", 0, false
	}
	iso := strings.ToUpper(s[:3])
	rest := strings.TrimSpace(s[4:])
	f, ok := parseAmount(rest)
	if !ok {
		return "", 0, false
	}
	return iso, f, true
}

func parseAmountISOTrailing(s string) (string, float64, bool) {
	parts := strings.Fields(s)
	if len(parts) < 2 {
		return "", 0, false
	}
	last := parts[len(parts)-1]
	if len(last) != 3 || !isAllASCIIAlpha(last) {
		return "", 0, false
	}
	iso := strings.ToUpper(last)
	num := strings.TrimSpace(strings.Join(parts[:len(parts)-1], " "))
	f, ok := parseAmount(num)
	if !ok {
		return "", 0, false
	}
	return iso, f, true
}

var symbolPrefixes = []struct {
	sym string
	iso string
}{
	{"US$", "USD"},
	{"CA$", "CAD"},
	{"C$", "CAD"},
	{"HK$", "HKD"},
	{"A$", "AUD"},
	{"MX$", "MXN"},
	{"$", "USD"},
	{"£", "GBP"},
	{"€", "EUR"},
	{"¥", "JPY"},
}

func parseSymbolPrefixAmount(s string) (string, float64, bool) {
	for _, sp := range symbolPrefixes {
		if strings.HasPrefix(s, sp.sym) {
			rest := strings.TrimSpace(strings.TrimPrefix(s, sp.sym))
			f, ok := parseAmount(rest)
			if ok {
				return sp.iso, f, true
			}
		}
	}
	return "", 0, false
}

func parseAmount(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	if s == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(s, 64)
	return f, err == nil
}

func isASCIIAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isAllASCIIAlpha(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isASCIIAlpha(s[i]) {
			return false
		}
	}
	return true
}
