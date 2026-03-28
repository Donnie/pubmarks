package main

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type metadata struct {
	Date                 time.Time         `json:"date"`
	FundCharacteristics  map[string]string `json:"fund_characteristics"`
	IndexCharacteristics map[string]string `json:"index_characteristics"`
	FundMarketPrice      map[string]string `json:"fund_market_price"`
}

var dateSections = map[string]bool{
	"Fund Characteristics":  true,
	"Index Characteristics": true,
	"Fund Market Price":     true,
}

func parsePage(doc *goquery.Document) (*metadata, error) {
	sections := make(map[string]map[string]string)
	rawDate := ""

	doc.Find("h2.comp-title").Each(func(_ int, h2 *goquery.Selection) {
		dateText := strings.TrimSpace(h2.Find("span.date").Text())
		title := strings.TrimSpace(h2.Clone().Find("span.date, svg").Remove().End().Text())

		if rawDate == "" && dateSections[title] && dateText != "" {
			rawDate = dateText
		}

		kv := make(map[string]string)
		h2.Parent().Find("table.tb-keyvalue tr").Each(func(_ int, tr *goquery.Selection) {
			label := strings.TrimSpace(tr.Find("td.label").Clone().Find(".info-data, .info").Remove().End().Text())
			label = strings.Join(strings.Fields(label), " ")
			value := strings.TrimSpace(tr.Find("td.data").Text())
			if label != "" {
				kv[label] = value
			}
		})

		if len(kv) > 0 {
			sections[title] = kv
		}
	})

	m := &metadata{}
	m.FundCharacteristics, _ = sections["Fund Characteristics"]
	m.IndexCharacteristics, _ = sections["Index Characteristics"]
	m.FundMarketPrice, _ = sections["Fund Market Price"]

	m.Date = time.Now()
	if rawDate != "" {
		if t, err := time.Parse("Jan 2 2006", strings.TrimPrefix(rawDate, "as of ")); err == nil {
			m.Date = t
		}
	}

	return m, nil
}

// fundTickerFromListing reads the primary listing ticker from the Listing Information table.
func fundTickerFromListing(doc *goquery.Document) string {
	var ticker string
	doc.Find("h2.comp-title").Each(func(_ int, h2 *goquery.Selection) {
		if ticker != "" {
			return
		}
		title := strings.TrimSpace(h2.Clone().Find("span.date, svg").Remove().End().Text())
		if title != "Listing Information" {
			return
		}
		// Columns: Exchange, Listing Date, Trading Currency, Ticker
		t := strings.TrimSpace(h2.Parent().Find("table.data-table tbody tr").First().Find("td").Eq(3).Text())
		if t != "" {
			ticker = strings.ToUpper(t)
		}
	})
	return ticker
}

// fundTickerFromPageURL uses the last path segment after the final hyphen (e.g. ...-spy → SPY).
func fundTickerFromPageURL(pageURL string) string {
	u := strings.TrimSpace(strings.TrimSuffix(pageURL, "/"))
	if i := strings.LastIndex(u, "/"); i >= 0 {
		u = u[i+1:]
	}
	if i := strings.LastIndex(u, "-"); i >= 0 {
		return strings.ToUpper(u[i+1:])
	}
	return ""
}

