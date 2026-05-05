package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"combine/pubmarks/gitfs"

	"pe5yr/internal/pe"
)

func main() {
	log.SetFlags(0)

	args := os.Args[1:]
	switch {
	case len(args) == 0:
		if err := regenerateAll(); err != nil {
			log.Fatal(err)
		}
	case len(args) == 1 && (args[0] == "-h" || args[0] == "--help"):
		fmt.Fprintln(os.Stderr, "usage: pe5yr [TICKER]")
		fmt.Fprintln(os.Stderr, "  With no arguments, writes pe-averages.json next to each datasets/stocks/*/combined.csv.")
		fmt.Fprintln(os.Stderr, "  With TICKER, writes only that ticker’s pe-averages.json.")
		os.Exit(2)
	case len(args) == 1:
		if err := regenerateOne(args[0]); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Fprintln(os.Stderr, "usage: pe5yr [TICKER]")
		os.Exit(2)
	}
}

func regenerateAll() error {
	ds, err := gitfs.ResolveDatasetsDir()
	if err != nil {
		return err
	}
	root := filepath.Join(ds, "stocks")
	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("read %q: %w", root, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		dir := filepath.Join(root, name)
		csvPath := filepath.Join(dir, "combined.csv")
		if _, err := os.Stat(csvPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("%s: %w", csvPath, err)
		}
		if err := writePeAverages(dir, strings.ToUpper(name)); err != nil {
			return err
		}
	}
	return nil
}

func regenerateOne(ticker string) error {
	ticker = strings.TrimSpace(ticker)
	if ticker == "" {
		return fmt.Errorf("empty ticker")
	}
	dir, err := gitfs.TickerStockDir(ticker)
	if err != nil {
		return err
	}
	return writePeAverages(dir, strings.ToUpper(ticker))
}

func writePeAverages(stockDir string, tickerUpper string) error {
	csvPath := filepath.Join(stockDir, "combined.csv")
	b, err := os.ReadFile(csvPath)
	if err != nil {
		return fmt.Errorf("%s: %w", csvPath, err)
	}
	res, err := pe.FiveYearAveragePeFromCombinedCSV(tickerUpper, string(b))
	if err != nil {
		return fmt.Errorf("%s: %w", tickerUpper, err)
	}
	outPath := filepath.Join(stockDir, "pe-averages.json")
	tmp := outPath + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create %q: %w", tmp, err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(toPayload(res)); err != nil {
		f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("encode json: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, outPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename to %q: %w", outPath, err)
	}
	log.Printf("wrote %s", outPath)
	return nil
}

// payload is the JSON object written to pe-averages.json (same shape as the former HTTP handler).
type payload struct {
	Ticker          string  `json:"ticker"`
	StartDate       string  `json:"start_date"`
	EndDate         string  `json:"end_date"`
	MinPe           float64 `json:"p_e_min"`
	MinPeDate       string  `json:"p_e_min_date"`
	MaxPe           float64 `json:"p_e_max"`
	MaxPeDate       string  `json:"p_e_max_date"`
	Mean5yrPe       float64 `json:"p_e_mean_5yr"`
	Median5yrPe     float64 `json:"p_e_median_5yr"`
	Mode5yrPe       float64 `json:"p_e_mode_5yr"`
	Avg5yrPe        float64 `json:"p_e_avg_5yr"`
	Ey5yrPe         float64 `json:"p_e_earningsyield_5yr"`
	LatestPe        float64 `json:"p_e_last"`
	Shiller5yrPe    float64 `json:"p_e_shiller_5yr"`
	Profitable5yrPe float64 `json:"p_e_profitable_5yr"`
	Lossy5yrPe      float64 `json:"p_e_lossy_5yr"`
	LastPrice       float64 `json:"price_last"`
	LastEps         float64 `json:"eps_last"`
}

func toPayload(r pe.Result) payload {
	return payload{
		Ticker:          r.Ticker,
		StartDate:       r.StartDate.Format("2006-01-02"),
		EndDate:         r.EndDate.Format("2006-01-02"),
		MinPe:           round4(r.MinPe),
		MinPeDate:       r.MinPeDate.Format("2006-01-02"),
		MaxPe:           round4(r.MaxPe),
		MaxPeDate:       r.MaxPeDate.Format("2006-01-02"),
		Mean5yrPe:       round4(r.Mean5yrPe),
		Median5yrPe:     round4(r.Median5yrPe),
		Mode5yrPe:       round4(math.Round(r.ModePe)),
		Avg5yrPe:        round4(r.Mean5yrPe),
		Ey5yrPe:         round4(r.Ey5yrPe),
		LatestPe:        round4(r.LatestPe),
		Shiller5yrPe:    round4(r.Shiller5yrPe),
		Profitable5yrPe: round4(r.Profitable5yrPe),
		Lossy5yrPe:      round4(r.Lossy5yrPe),
		LastPrice:       round2(r.LastPrice),
		LastEps:         round4(r.LastEps),
	}
}

func round4(v float64) float64 { return math.Round(v*10000) / 10000 }
func round2(v float64) float64 { return math.Round(v*100) / 100 }
