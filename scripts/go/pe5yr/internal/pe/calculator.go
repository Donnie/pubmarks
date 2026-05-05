package pe

import (
	"fmt"
	"math"
	"strings"
	"time"

	"donnie.in/sniper360/apps/pubmarks/internal/cdn"
	"donnie.in/sniper360/apps/pubmarks/internal/csvparse"
)

// tradeDay is one calendar trading row from combined.csv inside the 5-year window.
type tradeDay struct {
	date   time.Time
	close  float64
	epsTTM float64
}

// FiveYearAveragePe computes daily TTM P/E stats over an exact 5-calendar-year window.
//
// Price series:  daily OHLCV close from combined.csv.
// EPS series:    same-row ttm_net_eps (required on every in-window row).
// Daily P/E:     close / ttm
// Window:        [endDate - 5 years, endDate] where endDate is the latest trade date in the file.
func FiveYearAveragePe(ticker string, now time.Time) (Result, error) {
	_ = now
	ticker = strings.ToUpper(ticker)
	client := cdn.NewClient()
	text, err := client.FetchCombinedCSV(ticker)
	if err != nil {
		return Result{}, err
	}
	return fiveYearAveragePeFromCombined(ticker, text)
}

func fiveYearAveragePeFromCombined(ticker string, combinedCSV string) (Result, error) {
	startDate, endDate, lastPrice, parsed, err := csvparse.ParseCombinedFiveYearWindow(combinedCSV)
	if err != nil {
		return Result{}, err
	}
	if len(parsed) == 0 {
		return Result{}, fmt.Errorf("%s: no rows in combined 5-year window", ticker)
	}
	days := make([]tradeDay, len(parsed))
	for i := range parsed {
		days[i] = tradeDay{
			date:   parsed[i].Date,
			close:  parsed[i].Close,
			epsTTM: parsed[i].TtmNetEPS,
		}
	}
	return computeFiveYearPeResult(ticker, startDate, endDate, lastPrice, days), nil
}

type pePoint struct {
	pe   float64
	date time.Time
}

func computeFiveYearPeResult(
	ticker string,
	startDate, endDate time.Time,
	lastPrice float64,
	days []tradeDay,
) Result {
	series := make([]pePoint, 0, len(days))
	epsSeries := make([]float64, 0, len(days))
	peSeriesPositive := make([]float64, 0, len(days))
	peSeriesNegative := make([]float64, 0, len(days))
	eySeries := make([]float64, 0, len(days))
	for _, row := range days {
		eps := row.epsTTM
		pe := row.close / eps
		series = append(series, pePoint{pe: pe, date: row.date})
		epsSeries = append(epsSeries, eps)
		if eps > 0 {
			peSeriesPositive = append(peSeriesPositive, pe)
		} else if eps < 0 {
			peSeriesNegative = append(peSeriesNegative, pe)
		}
		eySeries = append(eySeries, eps/row.close)
	}

	mean5yrEps := mean(epsSeries)
	profitable5yrPe := meanIfFinite(peSeriesPositive)
	lossy5yrPe := meanIfFinite(peSeriesNegative)

	mean5yrEy := mean(eySeries)
	ey5yrPe := 1 / mean5yrEy

	pes := make([]float64, len(series))
	for i, p := range series {
		pes[i] = p.pe
	}

	minPoint, maxPoint := minMaxPePoints(series)

	lastEps := days[len(days)-1].epsTTM
	latestPe := lastPrice / lastEps
	shiller5yrPe := lastPrice / mean5yrEps

	return Result{
		Ticker:          ticker,
		StartDate:       startDate,
		EndDate:         endDate,
		MinPe:           minPoint.pe,
		MinPeDate:       minPoint.date,
		MaxPe:           maxPoint.pe,
		MaxPeDate:       maxPoint.date,
		Mean5yrPe:       mean(pes),
		Ey5yrPe:         ey5yrPe,
		Median5yrPe:     median(pes),
		ModePe:          modeIntegerBucket(pes),
		LatestPe:        latestPe,
		LastPrice:       lastPrice,
		LastEps:         lastEps,
		Mean5yrEps:      mean5yrEps,
		Shiller5yrPe:    shiller5yrPe,
		Profitable5yrPe: profitable5yrPe,
		Lossy5yrPe:      lossy5yrPe,
	}
}

func minMaxPePoints(series []pePoint) (minPoint pePoint, maxPoint pePoint) {
	minPoint = series[0]
	maxPoint = series[0]
	for _, p := range series[1:] {
		if p.pe < minPoint.pe {
			minPoint = p
		}
		if p.pe > maxPoint.pe {
			maxPoint = p
		}
	}
	return minPoint, maxPoint
}

func meanIfFinite(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	v := mean(xs)
	if !isFinite(v) {
		return 0
	}
	return v
}

func isFinite(v float64) bool { return !math.IsInf(v, 0) && !math.IsNaN(v) }
