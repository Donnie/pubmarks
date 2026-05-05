package cdn

import (
	"encoding/json"
	"fmt"
)

// ValidateManifestAppDeps reports whether b has the JSON structure the repo cares about for
// upstream compatibility: datasets.stocks.tickers[*] with ohlcv.yearRange.{min,max} and
// peratio.yearRange.{min,max} (integer years). Other top-level or sibling keys are ignored.
func ValidateManifestAppDeps(b []byte) error {
	var root map[string]any
	if err := json.Unmarshal(b, &root); err != nil {
		return fmt.Errorf("cdn manifest deps: %w", err)
	}
	ds, err := requireObjMap(root, "datasets")
	if err != nil {
		return fmt.Errorf("cdn manifest deps: %w", err)
	}
	stocks, err := requireObjMap(ds, "stocks")
	if err != nil {
		return fmt.Errorf("cdn manifest deps: %w", err)
	}
	tickers, err := requireObjMap(stocks, "tickers")
	if err != nil {
		return fmt.Errorf("cdn manifest deps: %w", err)
	}
	for sym, raw := range tickers {
		t, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("cdn manifest deps: datasets.stocks.tickers[%q] is not a JSON object", sym)
		}
		if err := requireYearRange(t, "ohlcv", sym); err != nil {
			return err
		}
		if err := requireYearRange(t, "peratio", sym); err != nil {
			return err
		}
	}
	return nil
}

func requireObjMap(m map[string]any, key string) (map[string]any, error) {
	v, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("missing %q", key)
	}
	out, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%q is not a JSON object", key)
	}
	return out, nil
}

func requireYearRange(t map[string]any, field, sym string) error {
	series, err := requireObjMap(t, field)
	if err != nil {
		return fmt.Errorf("cdn manifest deps: ticker %q: %s: %w", sym, field, err)
	}
	yr, err := requireObjMap(series, "yearRange")
	if err != nil {
		return fmt.Errorf("cdn manifest deps: ticker %q: %s: %w", sym, field, err)
	}
	if _, err := intYear(yr, "min"); err != nil {
		return fmt.Errorf("cdn manifest deps: ticker %q: %s.yearRange: %w", sym, field, err)
	}
	if _, err := intYear(yr, "max"); err != nil {
		return fmt.Errorf("cdn manifest deps: ticker %q: %s.yearRange: %w", sym, field, err)
	}
	return nil
}

// intYear reads a calendar year as encoded by encoding/json into map (float64) or a Go int.
func intYear(m map[string]any, key string) (int, error) {
	v, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("missing %q", key)
	}
	switch x := v.(type) {
	case float64:
		if x != float64(int(x)) {
			return 0, fmt.Errorf("%q is not an integer year: %v", key, v)
		}
		return int(x), nil
	case int:
		return x, nil
	default:
		return 0, fmt.Errorf("%q: invalid type for year", key)
	}
}

func parseManifestJSON(b []byte) error {
	var raw struct {
		Datasets struct {
			Stocks struct {
				Tickers map[string]any `json:"tickers"`
			} `json:"stocks"`
		} `json:"datasets"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("cdn manifest: %w", err)
	}
	if len(raw.Datasets.Stocks.Tickers) == 0 {
		return fmt.Errorf("cdn manifest: no stock tickers")
	}
	return nil
}
