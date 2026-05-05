package cdn

import (
	"strings"
	"testing"
)

func TestParseManifestJSONSucceedsOnPublisherShape(t *testing.T) {
	err := parseManifestJSON([]byte(`{
		"datasets": {
			"stocks": {
				"tickers": {
					"zz": {
						"ohlcv":   { "yearRange": { "min": 2004, "max": 2020 } },
						"peratio": { "yearRange": { "min": 2006, "max": 2020 } },
						"combined": { "file": "combined.csv" }
					}
				}
			}
		}
	}`))
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseManifestJSONRejectsEmptyTickers(t *testing.T) {
	err := parseManifestJSON([]byte(`{
		"datasets": {
			"stocks": {
				"tickers": {}
			}
		}
	}`))
	if err == nil {
		t.Fatal("expected error for empty tickers")
	}
}

// TestUpstreamManifestJSONSchema fetches the live CDN manifest and asserts the JSON
// paths still match (ValidateManifestAppDeps + parse). Daily content changes in year
// values are fine; renames or structural moves are not.
func TestUpstreamManifestJSONSchema(t *testing.T) {
	c := NewClient()
	u := manifestURL
	body, err := c.get(u)
	if err != nil {
		t.Fatalf("GET %s: %v", u, err)
	}
	if err := ValidateManifestAppDeps([]byte(body)); err != nil {
		t.Fatalf("app-dependent manifest structure: %v\n  URL: %s", err, u)
	}
	if err := parseManifestJSON([]byte(body)); err != nil {
		t.Fatalf("parseManifestJSON: %v\n  URL: %s", err, u)
	}
}

// TestUpstreamAaplCombinedFetch loads combined.csv from GitHub Pages.
func TestUpstreamAaplCombinedFetch(t *testing.T) {
	const ticker = "aapl"
	c := NewClient()
	body, err := c.FetchCombinedCSV(ticker)
	if err != nil {
		t.Fatal(err)
	}
	if body == "" {
		t.Fatal("empty combined.csv")
	}
	first := body
	if i := strings.IndexByte(body, '\n'); i >= 0 {
		first = body[:i]
	}
	lf := strings.ToLower(first)
	if !strings.Contains(lf, "date") || !strings.Contains(lf, "close") || !strings.Contains(lf, "ttm_net_eps") {
		t.Fatalf("unexpected combined header: %q", first)
	}
}
