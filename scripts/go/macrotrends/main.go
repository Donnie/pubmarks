// Fetches daily OHLCV from MacroTrends (chart iframe + stock_data_download CSV).
// If a year is set (YEAR env or argument), output is filtered to that calendar year;
// otherwise the full series from the download is written as CSV to stdout.
package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	chartIframePath = "https://www.macrotrends.net/production/stocks/desktop/PRODUCTION/stock_price_history.php"
	downloadBase    = "https://www.macrotrends.net/assets/php/stock_data_download.php"
	ohlcvHeader     = "date,open,high,low,close,volume"
	httpTimeout     = 30 * time.Second
)

var stockDataDownloadRE = regexp.MustCompile(`stock_data_download\.php\?s=([^&']+)&t=([^']+)`)

type candle struct {
	date   string
	open   float64
	high   float64
	low    float64
	close  float64
	volume float64
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("macrotrends: ")

	ticker, year, yearSet, err := parseInput()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	client := newHTTPClient(httpTimeout)
	var yb *int
	if yearSet {
		v := yearsBackForYear(year)
		yb = &v
	}

	csvBytes, err := fetchOHLCVCSV(ctx, client, normalizeSymbol(ticker), yb)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := parseMacroTrendsCSV(strings.NewReader(string(csvBytes)))
	if err != nil {
		log.Fatal(err)
	}
	out := rows
	if yearSet {
		from := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
		out = filterByRange(rows, from, to)
	}
	if err := writeCSV(os.Stdout, out); err != nil {
		log.Fatal(err)
	}
}

// parseInput resolves ticker and year from env (TICKER, YEAR) and optional args [ticker] [year].
// Command-line values override environment. If year is never set, yearSet is false (no date filter).
func parseInput() (ticker string, year int, yearSet bool, err error) {
	ticker = strings.TrimSpace(os.Getenv("TICKER"))
	year, yearSet, err = yearFromEnv()
	if err != nil {
		return "", 0, false, err
	}
	args := os.Args[1:]
	switch len(args) {
	case 0:
	case 1:
		if y, ok := parseYearToken(args[0]); ok {
			year = y
			yearSet = true
		} else {
			ticker = args[0]
		}
	case 2:
		ticker = args[0]
		y, ok := parseYearToken(args[1])
		if !ok {
			return "", 0, false, fmt.Errorf("second argument must be a 4-digit year (e.g. %d)", time.Now().Year())
		}
		year = y
		yearSet = true
	default:
		return "", 0, false, fmt.Errorf(`usage: %s [ticker] [year]
  ticker: TICKER env and/or first argument
  year:   YEAR env and/or last argument; a single 4-digit argument sets year only (ticker from env)
  if year is omitted, all rows from the download are printed`, os.Args[0])
	}
	if strings.TrimSpace(ticker) == "" {
		return "", 0, false, fmt.Errorf("set TICKER or pass ticker (e.g. %s AAPL)", os.Args[0])
	}
	return ticker, year, yearSet, nil
}

func yearFromEnv() (year int, set bool, err error) {
	s := strings.TrimSpace(os.Getenv("YEAR"))
	if s == "" {
		return 0, false, nil
	}
	y, ok := parseYearToken(s)
	if !ok {
		return 0, false, fmt.Errorf("YEAR: use a 4-digit year between 1900 and 2100")
	}
	return y, true, nil
}

func parseYearToken(s string) (year int, ok bool) {
	y, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || y < 1900 || y > 2100 {
		return 0, false
	}
	return y, true
}

func yearsBackForYear(targetYear int) int {
	cy := time.Now().Year()
	if targetYear > cy {
		return 1
	}
	n := cy - targetYear + 2
	if n < 1 {
		n = 1
	}
	if n > 40 {
		n = 40
	}
	return n
}

func normalizeSymbol(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}
}

func applyDefaultHeaders(req *http.Request) {
	h := req.Header
	h.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	h.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	h.Set("Accept-Language", "en-US,en;q=0.9")
}

func fetchHTML(ctx context.Context, client *http.Client, urlStr string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return "", err
	}
	applyDefaultHeaders(req)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("GET %s: %s", urlStr, resp.Status)
	}
	return string(body), nil
}

// chartIframeURL builds the chart iframe URL. If yearsBack is nil, the yb query param is omitted (MacroTrends default range).
func chartIframeURL(symbol string, yearsBack *int) string {
	sym := normalizeSymbol(symbol)
	q := url.Values{}
	q.Set("t", sym)
	if yearsBack != nil {
		q.Set("yb", fmt.Sprintf("%d", *yearsBack))
	}
	return chartIframePath + "?" + q.Encode()
}

func downloadURLFromChartHTML(chartHTML string) (string, error) {
	m := stockDataDownloadRE.FindStringSubmatch(chartHTML)
	if len(m) != 3 {
		return "", fmt.Errorf("stock_data_download URL not found in chart HTML")
	}
	u, err := url.Parse(downloadBase)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("s", m[1])
	q.Set("t", m[2])
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func fetchOHLCVCSV(ctx context.Context, client *http.Client, symbol string, yearsBack *int) ([]byte, error) {
	chartURL := chartIframeURL(symbol, yearsBack)
	html, err := fetchHTML(ctx, client, chartURL)
	if err != nil {
		return nil, err
	}
	dl, err := downloadURLFromChartHTML(html)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dl, nil)
	if err != nil {
		return nil, err
	}
	applyDefaultHeaders(req)
	req.Header.Set("Accept", "text/csv,*/*;q=0.9")
	req.Header.Set("Referer", chartURL)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET %s: %s", dl, resp.Status)
	}
	return body, nil
}

func parseMacroTrendsCSV(r io.Reader) ([]candle, error) {
	cr := csv.NewReader(r)
	cr.ReuseRecord = true
	cr.FieldsPerRecord = -1
	var headerFound bool
	var out []candle
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(rec) == 0 {
			continue
		}
		if !headerFound {
			if strings.EqualFold(joinCSVRecord(rec), ohlcvHeader) {
				headerFound = true
			}
			continue
		}
		if len(rec) != 6 {
			continue
		}
		c, err := rowToCandle(rec)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if !headerFound {
		return nil, fmt.Errorf("missing CSV header %q", ohlcvHeader)
	}
	return out, nil
}

func joinCSVRecord(rec []string) string {
	parts := make([]string, len(rec))
	for i, s := range rec {
		parts[i] = strings.TrimSpace(s)
	}
	return strings.Join(parts, ",")
}

func rowToCandle(rec []string) (candle, error) {
	date := strings.TrimSpace(rec[0])
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return candle{}, fmt.Errorf("bad date %q: %w", date, err)
	}
	open, err := strconv.ParseFloat(strings.TrimSpace(rec[1]), 64)
	if err != nil {
		return candle{}, err
	}
	high, err := strconv.ParseFloat(strings.TrimSpace(rec[2]), 64)
	if err != nil {
		return candle{}, err
	}
	low, err := strconv.ParseFloat(strings.TrimSpace(rec[3]), 64)
	if err != nil {
		return candle{}, err
	}
	close, err := strconv.ParseFloat(strings.TrimSpace(rec[4]), 64)
	if err != nil {
		return candle{}, err
	}
	vol, err := strconv.ParseFloat(strings.TrimSpace(rec[5]), 64)
	if err != nil {
		return candle{}, err
	}
	return candle{date: date, open: open, high: high, low: low, close: close, volume: vol}, nil
}

func filterByRange(rows []candle, from, to time.Time) []candle {
	out := rows[:0]
	for _, c := range rows {
		d, err := time.Parse("2006-01-02", c.date)
		if err != nil {
			continue
		}
		if d.Before(from) || d.After(to) {
			continue
		}
		out = append(out, c)
	}
	return out
}

func writeCSV(w io.Writer, rows []candle) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(strings.Split(ohlcvHeader, ",")); err != nil {
		return err
	}
	for _, c := range rows {
		rec := []string{
			c.date,
			fmtFloat(c.open),
			fmtFloat(c.high),
			fmtFloat(c.low),
			fmtFloat(c.close),
			fmtFloat(c.volume),
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
