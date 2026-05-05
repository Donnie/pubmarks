// Command pubmarks fetches combined.csv for a ticker and prints the same JSON as the AvgPe HTTP handler.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"donnie.in/sniper360/apps/pubmarks"
	"donnie.in/sniper360/apps/pubmarks/internal/pe"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "" || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Fprintln(os.Stderr, "usage: pubmarks <TICKER>")
		fmt.Fprintln(os.Stderr, "  Prints 5-year P/E metrics JSON (same shape as the Cloud Function success response).")
		os.Exit(2)
	}
	ticker := os.Args[1]

	res, err := pe.FiveYearAveragePe(ticker, time.Now())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(pubmarks.ToPayload(res)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
