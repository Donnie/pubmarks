package parse

import (
	"github.com/gocarina/gocsv"
)

func Price(csvText string) (OHLCV, error) {
	var rows []*ohlcvCSVRow
	if err := gocsv.UnmarshalString(csvText, &rows); err != nil {
		return nil, err
	}

	ohlcv := make(OHLCV, len(rows))
	for _, row := range rows {
		date := row.Date.Time
		ohlcv[date] = OHLCVRow{
			Date:   date,
			Open:   row.Open,
			High:   row.High,
			Low:    row.Low,
			Close:  row.Close,
			Volume: row.Volume,
		}
	}
	return ohlcv, nil
}

func Peratio(csvText string) (EPSTTM, error) {
	var rows []*peratioCSVRow
	if err := gocsv.UnmarshalString(csvText, &rows); err != nil {
		return nil, err
	}

	peratio := make(EPSTTM, len(rows))
	for _, row := range rows {
		date := row.Date.Time
		peratio[date] = EPSTTMRow{
			Date:       date,
			StockPrice: row.StockPrice,
			TtmNetEps:  row.TtmNetEps,
			PeRatio:    row.PeRatio,
		}
	}
	return peratio, nil
}
