package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TickerNotInManifestError is returned when a symbol is not listed in manifest.json under datasets.stocks.tickers.
type TickerNotInManifestError struct {
	Ticker string
}

func (e TickerNotInManifestError) Error() string {
	return fmt.Sprintf("cdn: ticker %q is not in manifest", e.Ticker)
}

// Manifest is the pubmarks year-range index for stock CSVs on pagesOrigin (authoritative for published years).
type Manifest struct {
	stocks map[string]stockInfo
}

type stockInfo struct {
	ohlcv   yearSpan
	peratio yearSpan
}

type yearSpan struct{ min, max int }

// FetchManifest loads manifest.json from the GitHub Pages site used for CSVs.
func FetchManifest() (*Manifest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET manifest: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET manifest: HTTP %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	return parseManifest(body)
}

func parseManifest(b []byte) (*Manifest, error) {
	var raw struct {
		Datasets struct {
			Stocks struct {
				Tickers map[string]struct {
					Ohlcv struct {
						YearRange struct {
							Min int `json:"min"`
							Max int `json:"max"`
						} `json:"yearRange"`
					} `json:"ohlcv"`
					Peratio struct {
						YearRange struct {
							Min int `json:"min"`
							Max int `json:"max"`
						} `json:"yearRange"`
					} `json:"peratio"`
				} `json:"tickers"`
			} `json:"stocks"`
		} `json:"datasets"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("cdn manifest: %w", err)
	}
	m := &Manifest{stocks: make(map[string]stockInfo, len(raw.Datasets.Stocks.Tickers))}
	for key, t := range raw.Datasets.Stocks.Tickers {
		lt := strings.ToLower(key)
		m.stocks[lt] = stockInfo{
			ohlcv:   yearSpan{min: t.Ohlcv.YearRange.Min, max: t.Ohlcv.YearRange.Max},
			peratio: yearSpan{min: t.Peratio.YearRange.Min, max: t.Peratio.YearRange.Max},
		}
	}
	if len(m.stocks) == 0 {
		return nil, fmt.Errorf("cdn manifest: no stock tickers")
	}
	return m, nil
}

// OhlcvYearSlice returns calendar years in [needMin, needMax] that exist in the manifest for OHLCV.
func (m *Manifest) OhlcvYearSlice(ticker string, needMin, needMax int) ([]int, error) {
	span, err := m.lookupOhlcv(ticker)
	if err != nil {
		return nil, err
	}
	ys := intersectYearSpan(needMin, needMax, span)
	if len(ys) == 0 {
		return nil, fmt.Errorf(
			"cdn: %s OHLCV need %d..%d does not overlap manifest range %d..%d",
			strings.ToUpper(ticker), needMin, needMax, span.min, span.max,
		)
	}
	return ys, nil
}

// PeratioYearSlice returns calendar years in [needMin, needMax] that exist in the manifest for peratio.csv.
func (m *Manifest) PeratioYearSlice(ticker string, needMin, needMax int) ([]int, error) {
	span, err := m.lookupPeratio(ticker)
	if err != nil {
		return nil, err
	}
	ys := intersectYearSpan(needMin, needMax, span)
	if len(ys) == 0 {
		return nil, fmt.Errorf(
			"cdn: %s peratio need %d..%d does not overlap manifest range %d..%d",
			strings.ToUpper(ticker), needMin, needMax, span.min, span.max,
		)
	}
	return ys, nil
}

// OhlcvAllPublishedYears returns every calendar year the manifest lists for OHLCV for ticker (yearRange.min through max, inclusive).
func (m *Manifest) OhlcvAllPublishedYears(ticker string) ([]int, error) {
	span, err := m.lookupOhlcv(ticker)
	if err != nil {
		return nil, err
	}
	if span.min > span.max {
		return nil, fmt.Errorf(
			"cdn: %s OHLCV manifest yearRange invalid %d..%d",
			strings.ToUpper(ticker), span.min, span.max,
		)
	}
	return yearsInclusive(span.min, span.max), nil
}

// PeratioAllPublishedYears returns every calendar year the manifest lists for peratio.csv for ticker (yearRange.min through max, inclusive).
func (m *Manifest) PeratioAllPublishedYears(ticker string) ([]int, error) {
	span, err := m.lookupPeratio(ticker)
	if err != nil {
		return nil, err
	}
	if span.min > span.max {
		return nil, fmt.Errorf(
			"cdn: %s peratio manifest yearRange invalid %d..%d",
			strings.ToUpper(ticker), span.min, span.max,
		)
	}
	return yearsInclusive(span.min, span.max), nil
}

func (m *Manifest) lookupOhlcv(ticker string) (yearSpan, error) {
	t := strings.ToLower(ticker)
	s, ok := m.stocks[t]
	if !ok {
		return yearSpan{}, TickerNotInManifestError{Ticker: strings.ToUpper(ticker)}
	}
	return s.ohlcv, nil
}

func (m *Manifest) lookupPeratio(ticker string) (yearSpan, error) {
	t := strings.ToLower(ticker)
	s, ok := m.stocks[t]
	if !ok {
		return yearSpan{}, TickerNotInManifestError{Ticker: strings.ToUpper(ticker)}
	}
	return s.peratio, nil
}

// intersectYearSpan returns the inclusive year list for [needMin,needMax] ∩ [d.min,d.max] (binding manifest).
func intersectYearSpan(needMin, needMax int, d yearSpan) []int {
	if needMax < needMin {
		return nil
	}
	lo, hi := needMin, needMax
	if d.min > lo {
		lo = d.min
	}
	if d.max < hi {
		hi = d.max
	}
	if lo > hi {
		return nil
	}
	return yearsInclusive(lo, hi)
}

func yearsInclusive(lo, hi int) []int {
	if hi < lo {
		return nil
	}
	out := make([]int, 0, hi-lo+1)
	for y := lo; y <= hi; y++ {
		out = append(out, y)
	}
	return out
}
