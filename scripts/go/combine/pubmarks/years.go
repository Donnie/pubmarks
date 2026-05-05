package pubmarks

import (
	"fmt"

	"combine/pubmarks/gitfs"
)

// YearsFromDatasetsTickerDir returns OHLCV and peratio calendar-year lists discovered from
// datasets/stocks/<ticker>/*/ohlcv.csv and .../peratio.csv (years that exist on disk).
func YearsFromDatasetsTickerDir(ticker string) (ohlcvYears, peratioYears []int, err error) {
	ohlcvYears, err = gitfs.YearsWithCSV(ticker, "ohlcv")
	if err != nil {
		return nil, nil, fmt.Errorf("ohlcv years: %w", err)
	}
	peratioYears, err = gitfs.YearsWithCSV(ticker, "peratio")
	if err != nil {
		return nil, nil, fmt.Errorf("peratio years: %w", err)
	}
	return ohlcvYears, peratioYears, nil
}
