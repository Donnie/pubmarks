package pubmarks

import (
	"fmt"

	"combine/pubmarks/cdn"
)

// YearsFromManifestMaxRange returns OHLCV and peratio calendar-year lists for the ticker using each series'
// full yearRange.min..max from manifest.json (maximum span the CDN index advertises).
func YearsFromManifestMaxRange(ticker string) (ohlcvYears, peratioYears []int, err error) {
	m, err := cdn.FetchManifest()
	if err != nil {
		return nil, nil, fmt.Errorf("fetch manifest: %w", err)
	}
	ohlcvYears, err = m.OhlcvAllPublishedYears(ticker)
	if err != nil {
		return nil, nil, fmt.Errorf("ohlcv years: %w", err)
	}
	peratioYears, err = m.PeratioAllPublishedYears(ticker)
	if err != nil {
		return nil, nil, fmt.Errorf("peratio years: %w", err)
	}
	return ohlcvYears, peratioYears, nil
}
