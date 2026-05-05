package csvparse

import (
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"pe5yr/internal/dateutil"
)

// CombinedDay is one trading row inside the 5-year window from combined.csv
// (date, close, ttm_net_eps).
type CombinedDay struct {
	Date      time.Time
	Close     float64
	TtmNetEPS float64
}

// ParseCombinedFiveYearWindow reads combined.csv once, uses the file’s last row
// for endDate and last close, and returns rows with date in
// [dateutil.YearsBefore(endDate, 5), endDate], sorted by date ascending.
func ParseCombinedFiveYearWindow(csvText string) (
	startDate, endDate time.Time,
	lastPrice float64,
	days []CombinedDay,
	err error,
) {
	headers, rows, err := readAll(csvText)
	if err != nil {
		return time.Time{}, time.Time{}, 0, nil, err
	}
	idx, err := columnIndex(headers, []string{"date", "close", "ttm_net_eps"}, "combined.csv")
	if err != nil {
		return time.Time{}, time.Time{}, 0, nil, err
	}
	if len(rows) == 0 {
		return time.Time{}, time.Time{}, 0, nil, fmt.Errorf("combined.csv: no data rows")
	}

	last := rows[len(rows)-1]
	endDate, err = parseDateRequired(last[idx["date"]], "combined.csv last row")
	if err != nil {
		return time.Time{}, time.Time{}, 0, nil, err
	}
	lastPrice, err = parseFloatRequired(last[idx["close"]], "combined.csv last row close")
	if err != nil {
		return time.Time{}, time.Time{}, 0, nil, err
	}

	startDate = dateutil.YearsBefore(endDate, 5)
	days = make([]CombinedDay, 0, len(rows))
	for _, row := range rows {
		d, err := parseDateRequired(row[idx["date"]], "combined.csv")
		if err != nil {
			return time.Time{}, time.Time{}, 0, nil, err
		}
		if d.Before(startDate) {
			continue
		}
		if d.After(endDate) {
			continue
		}
		close, err := parseFloatRequired(row[idx["close"]], "combined.csv close")
		if err != nil {
			return time.Time{}, time.Time{}, 0, nil, err
		}
		eps, err := parseFloatRequired(row[idx["ttm_net_eps"]], "combined.csv ttm_net_eps")
		if err != nil {
			return time.Time{}, time.Time{}, 0, nil, err
		}
		days = append(days, CombinedDay{Date: d, Close: close, TtmNetEPS: eps})
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i].Date.Before(days[j].Date)
	})

	return startDate, endDate, lastPrice, days, nil
}

// readAll parses csvText and returns the header row plus all data rows.
func readAll(csvText string) (headers []string, rows [][]string, err error) {
	r := csv.NewReader(strings.NewReader(csvText))
	r.TrimLeadingSpace = true

	all, err := r.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("csv parse error: %w", err)
	}
	if len(all) == 0 {
		return nil, nil, fmt.Errorf("csv has no header row")
	}
	return all[0], all[1:], nil
}

// columnIndex builds a name→index map and validates that all required columns exist.
func columnIndex(headers []string, required []string, filename string) (map[string]int, error) {
	idx := make(map[string]int, len(headers))
	for i, h := range headers {
		idx[strings.TrimSpace(h)] = i
	}
	for _, col := range required {
		if _, ok := idx[col]; !ok {
			return nil, fmt.Errorf("%s missing column %q", filename, col)
		}
	}
	return idx, nil
}

func parseDateRequired(raw, ctx string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("%s: empty date", ctx)
	}
	d, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s: bad date %q", ctx, raw)
	}
	return d, nil
}

func parseFloatRequired(raw, ctx string) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("%s: empty number", ctx)
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", ctx, err)
	}
	return v, nil
}
