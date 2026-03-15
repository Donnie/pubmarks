package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const ticker = "SPY"

func main() {
	xlsxPath, cleanup, err := downloadXLSX(ticker)
	if err != nil {
		fatal(err)
	}
	defer cleanup()

	holdings, err := parseHoldings(xlsxPath)
	if err != nil {
		fatal(err)
	}

	if err := json.NewEncoder(os.Stdout).Encode(holdings); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
