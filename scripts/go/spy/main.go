package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const ticker = "SPY"

func main() {
	var (
		holdings []holding
		meta     *metadata
		holdErr  error
		metaErr  error
		wg       sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		xlsxPath, cleanup, err := downloadXLSX(ticker)
		if err != nil {
			holdErr = err
			return
		}
		defer cleanup()
		holdings, holdErr = parseHoldings(xlsxPath)
	}()

	go func() {
		defer wg.Done()
		body, err := fetchPage(ticker)
		if err != nil {
			metaErr = err
			return
		}
		meta, metaErr = parsePage(body)
	}()

	wg.Wait()

	if holdErr != nil {
		fatal(holdErr)
	}
	if metaErr != nil {
		fatal(metaErr)
	}

	out := struct {
		*metadata
		Holdings []holding `json:"holdings"`
	}{meta, holdings}

	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
