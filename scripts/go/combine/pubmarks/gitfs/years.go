package gitfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// TickerNotInDatasetsError is returned when datasets/stocks/<ticker> is missing or not a directory.
type TickerNotInDatasetsError struct {
	Ticker string
	Path   string
}

func (e TickerNotInDatasetsError) Error() string {
	return fmt.Sprintf("gitfs: ticker %q has no datasets directory at %q", e.Ticker, e.Path)
}

func resolveStockTickerDir(ticker string) (string, error) {
	ds, err := ResolveDatasetsDir()
	if err != nil {
		return "", err
	}
	stockDir := filepath.Join(ds, "stocks", strings.ToLower(strings.TrimSpace(ticker)))
	fi, err := os.Stat(stockDir)
	if err != nil || !fi.IsDir() {
		return "", TickerNotInDatasetsError{Ticker: strings.ToUpper(strings.TrimSpace(ticker)), Path: stockDir}
	}
	return stockDir, nil
}

// TickerStockDir returns the path to datasets/stocks/<ticker> if present.
func TickerStockDir(ticker string) (string, error) {
	return resolveStockTickerDir(ticker)
}

// YearsWithCSV lists calendar-year subdirectory names under datasets/stocks/<ticker> where <year>/<basename>.csv exists.
func YearsWithCSV(ticker, basename string) ([]int, error) {
	stockDir, err := resolveStockTickerDir(ticker)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(stockDir)
	if err != nil {
		return nil, fmt.Errorf("gitfs: read ticker dir %q: %w", stockDir, err)
	}

	var years []int
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		name := ent.Name()
		y, err := strconv.Atoi(name)
		if err != nil || y < 1000 || y > 9999 {
			continue
		}
		csvPath := filepath.Join(stockDir, name, basename+".csv")
		if _, err := os.Stat(csvPath); err != nil {
			continue
		}
		years = append(years, y)
	}
	if len(years) == 0 {
		return nil, fmt.Errorf("gitfs: %s no %s.csv files under %q", strings.ToUpper(ticker), basename, stockDir)
	}
	sort.Ints(years)
	return years, nil
}
