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

func parsePage(body []byte) (*metadata, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

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

	if rawDate != "" {
		t, _ := time.Parse("Jan 2 2006", strings.TrimPrefix(rawDate, "as of "))
		if t.IsZero() {
			t = time.Now()
		}
		m.Date = t
	}

	return m, nil
}
