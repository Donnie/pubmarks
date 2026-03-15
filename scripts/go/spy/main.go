package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode"
)

func main() {
	ticker := os.Getenv("TICKER")
	if ticker == "" {
		fmt.Fprintln(os.Stderr, "TICKER environment variable is required")
		os.Exit(1)
	}

	var (
		holdings []holding
		meta     *metadata
		holdErr  error
		metaErr  error
		wg       sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		xlsxPath, cleanup, err := downloadXLSX(ticker)
		if err != nil {
			holdErr = err
			return
		}
		defer cleanup()
		holdings, holdErr = parseHoldings(xlsxPath)
	}()

	go func() {
		defer wg.Done()
		body, err := fetchPage(ticker)
		if err != nil {
			metaErr = err
			return
		}
		meta, metaErr = parsePage(body)
	}()

	wg.Wait()

	if holdErr != nil {
		fatal(holdErr)
	}
	if metaErr != nil {
		fatal(metaErr)
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
	if out == "price_earnings_ratio_fy1" {
		return "price_earnings_fw"
	}
	return out
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
