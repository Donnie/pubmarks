package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"
)

var keyRenames = map[string]string{
	"price_earnings_ratio_fy1": "price_earnings_fw",
}

func main() {
	pageURL := os.Getenv("PAGE_URL")
	if pageURL == "" {
		fmt.Fprintln(os.Stderr, "PAGE_URL is not set")
		os.Exit(1)
	}

	doc, err := fetchPage(pageURL)
	if err != nil {
		fatal(err)
	}

	xlsxURL, err := holdingsXLSXURL(doc, pageURL)
	if err != nil {
		fatal(err)
	}

	xlsxPath, cleanup, err := downloadXLSX(xlsxURL)
	if err != nil {
		fatal(err)
	}
	defer cleanup()

	holdings, err := parseHoldings(xlsxPath)
	if err != nil {
		fatal(err)
	}

	meta, err := parsePage(doc)
	if err != nil {
		fatal(err)
	}

	out := make(map[string]any)
	out["date"] = meta.Date.Format("2006-01-02")
	out["holdings"] = holdings
	for k, v := range meta.FundCharacteristics {
		out[normalizeKey(k)] = v
	}
	for k, v := range meta.IndexCharacteristics {
		out[normalizeKey(k)] = v
	}
	for k, v := range meta.FundMarketPrice {
		out[normalizeKey(k)] = v
	}

	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fatal(err)
	}
}

// normalizeKey lowercases and replaces any non-alphanumeric rune (including space) with _.
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

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
