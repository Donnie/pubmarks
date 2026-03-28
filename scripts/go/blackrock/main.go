package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	pageURL := os.Getenv("PAGE_URL")
	if pageURL == "" {
		fmt.Fprintln(os.Stderr, "PAGE_URL is not set")
		os.Exit(1)
	}

	doc, err := fetchPage(pageURL)
	if err != nil {
		fatal(err)
	}

	jsonURL, err := holdingsJSONURI(doc, pageURL)
	if err != nil {
		fatal(err)
	}
	holdingsRaw, err := fetchBytes(jsonURL)
	if err != nil {
		fatal(err)
	}
	holdingsRaw = bytes.TrimPrefix(holdingsRaw, []byte("\xef\xbb\xbf"))

	holdings, err := parseHoldingsJSON(holdingsRaw)
	if err != nil {
		fatal(err)
	}

	meta, err := parseProductPage(doc)
	if err != nil {
		fatal(err)
	}

	if csvURL, err := holdingsCSVURI(doc, pageURL); err == nil {
		if csvBytes, err := fetchBytes(csvURL); err == nil {
			if d, err := holdingsAsOfFromCSV(csvBytes); err == nil {
				meta.Date = d
			}
		}
	}

	out := make(map[string]any)
	out["date"] = meta.Date.Format("2006-01-02")
	out["holdings"] = holdings

	for k, v := range meta.Meta {
		out[k] = v
	}

	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
