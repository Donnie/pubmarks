package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"combine/pubmarks"
	"combine/pubmarks/parse"
)

type peratioOutRow struct {
	TtmNetEps float64
	PeRatio   float64
}

func main() {
	// get the ticker from the command line
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <ticker>", os.Args[0])
	}
	ticker := strings.TrimSpace(strings.ToLower(os.Args[1]))

	// Full published span per series from manifest.json (yearRange.min..max), then fetch those CSVs.
	ohlcvYears, peratioYears, err := pubmarks.YearsFromManifestMaxRange(ticker)
	if err != nil {
		log.Fatal(err)
	}

	ohlcv, err := pubmarks.OHLCV(ticker, ohlcvYears)
	if err != nil {
		log.Fatal(err)
	}

	peratio, err := pubmarks.Peratio(ticker, peratioYears)
	if err != nil {
		log.Fatal(err)
	}

	ohlcvDates := sortedDatesFromOHLCV(ohlcv)
	peratioShifted := shiftPeratioToTradingDays(peratio, ohlcv, ohlcvDates)
	dates := unionSortedDates(ohlcvDates, peratioShifted)

	fmt.Println("date,open,high,low,close,volume,ttm_net_eps,pe_calc")

	epsFilled := fillEPSPiecewiseLinear(dates, peratioShifted)
	printCombinedCSV(dates, ohlcv, peratioShifted, epsFilled)
}

func sortedDatesFromOHLCV(ohlcv parse.OHLCV) []time.Time {
	ohlcvDates := make([]time.Time, 0, len(ohlcv))
	for d := range ohlcv {
		ohlcvDates = append(ohlcvDates, d)
	}
	sort.Slice(ohlcvDates, func(i, j int) bool { return ohlcvDates[i].Before(ohlcvDates[j]) })
	return ohlcvDates
}

// If peratio lands on a non-trading day (no OHLCV row), shift it to the previous trading day.
// If multiple peratio rows collapse onto the same shifted day, keep the one with the latest original date.
func shiftPeratioToTradingDays(peratio parse.EPSTTM, ohlcv parse.OHLCV, ohlcvDates []time.Time) map[time.Time]peratioOutRow {
	peratioShifted := make(map[time.Time]peratioOutRow, len(peratio))
	peratioOrig := make(map[time.Time]time.Time, len(peratio)) // shiftedDate -> originalDate (for tie-breaking)

	for d, row := range peratio {
		shifted := d
		if _, ok := ohlcv[d]; !ok {
			i := sort.Search(len(ohlcvDates), func(i int) bool { return !ohlcvDates[i].Before(d) }) // first >= d
			if i == 0 {
				// no earlier trading day in our OHLCV window; keep as-is
				shifted = d
			} else {
				shifted = ohlcvDates[i-1]
			}
		}

		if prevOrig, ok := peratioOrig[shifted]; ok && prevOrig.After(d) {
			continue
		}
		peratioOrig[shifted] = d
		peratioShifted[shifted] = peratioOutRow{TtmNetEps: row.TtmNetEps, PeRatio: row.PeRatio}
	}

	return peratioShifted
}

func unionSortedDates(ohlcvDates []time.Time, peratioShifted map[time.Time]peratioOutRow) []time.Time {
	dates := make([]time.Time, 0, len(ohlcvDates)+len(peratioShifted))
	seen := make(map[time.Time]struct{}, len(ohlcvDates)+len(peratioShifted))

	for _, d := range ohlcvDates {
		seen[d] = struct{}{}
		dates = append(dates, d)
	}
	for d := range peratioShifted {
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		dates = append(dates, d)
	}

	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	return dates
}

// Build a piecewise-linear EPS series between dates where EPS is actually present.
// EPS stays empty before the first anchor. After the last anchor, EPS is carried forward.
func fillEPSPiecewiseLinear(dates []time.Time, peratioShifted map[time.Time]peratioOutRow) map[time.Time]float64 {
	dateIdx := make(map[time.Time]int, len(dates))
	for i, d := range dates {
		dateIdx[d] = i
	}

	anchors := make([]time.Time, 0, len(peratioShifted))
	for d, row := range peratioShifted {
		if row.TtmNetEps != 0 {
			anchors = append(anchors, d)
		}
	}
	sort.Slice(anchors, func(i, j int) bool { return anchors[i].Before(anchors[j]) })

	epsFilled := make(map[time.Time]float64, len(dates))
	for i := 0; i+1 < len(anchors); i++ {
		a := anchors[i]
		b := anchors[i+1]
		ai, aok := dateIdx[a]
		bi, bok := dateIdx[b]
		if !aok || !bok || bi <= ai {
			continue
		}

		epsA := peratioShifted[a].TtmNetEps
		epsB := peratioShifted[b].TtmNetEps
		span := float64(bi - ai)
		for j := ai; j <= bi; j++ {
			t := float64(j-ai) / span
			epsFilled[dates[j]] = epsA + (epsB-epsA)*t
		}
	}

	if len(anchors) > 0 {
		lastAnchor := anchors[len(anchors)-1]
		lastIdx, ok := dateIdx[lastAnchor]
		if ok {
			lastEps := peratioShifted[lastAnchor].TtmNetEps
			for j := lastIdx; j < len(dates); j++ {
				epsFilled[dates[j]] = lastEps
			}
		}
	}

	return epsFilled
}

func printCombinedCSV(
	dates []time.Time,
	ohlcv parse.OHLCV,
	peratioShifted map[time.Time]peratioOutRow,
	epsFilled map[time.Time]float64,
) {
	for _, d := range dates {
		ohlcvRow, hasOHLCV := ohlcv[d]
		_, hasPERatio := peratioShifted[d]

		open, high, low, close, volume := "", "", "", "", ""
		var closeNum float64
		hasCloseNum := false
		if hasOHLCV {
			open = fmt.Sprintf("%.4f", ohlcvRow.Open)
			high = fmt.Sprintf("%.4f", ohlcvRow.High)
			low = fmt.Sprintf("%.4f", ohlcvRow.Low)
			close = fmt.Sprintf("%.4f", ohlcvRow.Close)
			volume = fmt.Sprintf("%d", ohlcvRow.Volume)
			closeNum = ohlcvRow.Close
			hasCloseNum = true
		}

		ttmNetEps, peCalc := "", ""
		var epsNum float64
		hasEpsNum := false
		if eps, ok := epsFilled[d]; ok {
			epsNum = eps
			hasEpsNum = true
			ttmNetEps = fmt.Sprintf("%.4f", epsNum)
		} else if hasPERatio {
			hasEpsNum = false
		}

		if hasCloseNum && hasEpsNum && epsNum != 0 {
			peCalc = fmt.Sprintf("%.4f", closeNum/epsNum)
		}

		fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s\n",
			d.Format("2006-01-02"),
			open, high, low, close, volume,
			ttmNetEps, peCalc,
		)
	}
}
