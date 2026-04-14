package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	var ohlcv bool

	cmd := &cobra.Command{
		Use:   "macrotrends [ticker] [year]",
		Short: "Download MacroTrends stock OHLCV as CSV",
		Long: `Download daily OHLCV from MacroTrends (chart iframe + stock_data_download CSV).

Ticker: TICKER environment variable and/or first argument.
Year:   YEAR environment and/or last argument; a single 4-digit argument sets year only (ticker from env).
If year is omitted, all rows from the download are printed.

You must pass --ohlcv to perform a fetch (explicit opt-in).`,
		Example: `  macrotrends --ohlcv AAPL
  macrotrends --ohlcv MSFT 2024
  TICKER=AAPL macrotrends --ohlcv 2024`,
		Args:          cobra.RangeArgs(0, 2),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !ohlcv {
				return errNeedOhlcvFlag(cmd)
			}
			ticker, year, yearSet, err := parseTickerYear(args)
			if err != nil {
				return err
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
				return err
			}
			rows, err := parseMacroTrendsCSV(strings.NewReader(string(csvBytes)))
			if err != nil {
				return err
			}
			out := rows
			if yearSet {
				from := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
				to := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
				out = filterByRange(rows, from, to)
			}
			return writeCSV(cmd.OutOrStdout(), out)
		},
	}

	cmd.Flags().BoolVar(&ohlcv, "ohlcv", false, "fetch OHLCV CSV from MacroTrends (required)")
	return cmd
}

func errNeedOhlcvFlag(cmd *cobra.Command) error {
	path := cmd.CommandPath()
	var b strings.Builder
	fmt.Fprintf(&b, "OHLCV output requires --ohlcv\n\n")
	fmt.Fprintf(&b, "Examples:\n")
	fmt.Fprintf(&b, "  %s --ohlcv AAPL\n", path)
	fmt.Fprintf(&b, "  %s --ohlcv MSFT 2024\n", path)
	fmt.Fprintf(&b, "  TICKER=AAPL %s --ohlcv 2024\n\n", path)
	fmt.Fprintf(&b, "Run \"%s --help\" for full usage.\n", path)
	return errors.New(strings.TrimRight(b.String(), "\n"))
}
