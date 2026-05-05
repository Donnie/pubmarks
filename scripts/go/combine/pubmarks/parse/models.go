package parse

import (
	"strings"
	"time"
)

type CSVDate struct {
	time.Time
}

func (d *CSVDate) UnmarshalCSV(value string) error {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

type ohlcvCSVRow struct {
	Date   CSVDate `csv:"date"`
	Open   float64 `csv:"open"`
	High   float64 `csv:"high"`
	Low    float64 `csv:"low"`
	Close  float64 `csv:"close"`
	Volume int64   `csv:"volume"`
}

type peratioCSVRow struct {
	Date       CSVDate `csv:"date"`
	StockPrice float64 `csv:"stock_price"`
	TtmNetEps  float64 `csv:"ttm_net_eps"`
	PeRatio    float64 `csv:"pe_ratio"`
}

type OHLCVRow struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

type OHLCV map[time.Time]OHLCVRow

type EPSTTMRow struct {
	Date       time.Time
	StockPrice float64
	TtmNetEps  float64
	PeRatio    float64
}

type EPSTTM map[time.Time]EPSTTMRow
