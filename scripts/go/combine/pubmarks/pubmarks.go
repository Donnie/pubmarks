package pubmarks

import (
	"fmt"

	"combine/pubmarks/cdn"
	"combine/pubmarks/parse"
)

func OHLCV(ticker string, years []int) (parse.OHLCV, error) {
	ohlcvCSV, err := cdn.GetYears(ticker, years, "ohlcv")
	if err != nil {
		return nil, fmt.Errorf("failed to get ohlcv: %w", err)
	}

	var ohlcv parse.OHLCV
	if err := ohlcv.Hydrate(ohlcvCSV); err != nil {
		return nil, fmt.Errorf("failed to parse ohlcv csvs: %w", err)
	}
	return ohlcv, nil
}

func Peratio(ticker string, years []int) (parse.EPSTTM, error) {
	peratioCSV, err := cdn.GetYears(ticker, years, "peratio")
	if err != nil {
		return nil, fmt.Errorf("failed to get peratio: %w", err)
	}

	var peratio parse.EPSTTM
	if err := peratio.Hydrate(peratioCSV); err != nil {
		return nil, fmt.Errorf("failed to parse eps csvs: %w", err)
	}
	return peratio, nil
}
