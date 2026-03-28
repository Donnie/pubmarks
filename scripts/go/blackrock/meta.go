package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Metadata keys use the same normalizeKey + keyRenames as statestreet (see keys.go).
//
// BlackRock captions that differ from SSGA wording but match the same datapoint use
// explicit keys (Index Characteristics on SSGA uses "Price/Earnings" for trailing P/E;
// "Price/Book Ratio" for P/B). BlackRock shows "P/E Ratio" / "P/B Ratio" which would
// normalize to different strings, so those columns are remapped.
//
// Typical statestreet keys absent on BlackRock EMEA ETF pages (omitted):
//   30_day_median_bid_ask_spread, bid_ask, closing_price, day_high, day_low,
//   est_3_5_year_eps_growth, exchange_volume_shares, premium_discount,
//   price_cash_flow, price_earnings_fw, weighted_average_market_cap
// base_currency and price (float64) are derived from the header NAV string in main.
//
// Other BlackRock fields use normalizeKey(clean caption text), same as statestreet would
// if that label appeared in its key/value tables.

// colClassStatestreetKey: SSGA-equivalent JSON keys (see file comment).
var colClassStatestreetKey = map[string]string{
	"col-priceBook":     "price_book_ratio",
	"col-priceEarnings": "price_earnings",
}

type parsedMeta struct {
	Date time.Time
	Meta map[string]string
}

func parseProductPage(doc *goquery.Document) (*parsedMeta, error) {
	pm := &parsedMeta{Meta: make(map[string]string)}

	doc.Find("#keyFundFacts .product-data-item, #fundamentalsAndRisk .product-data-item").Each(func(_ int, s *goquery.Selection) {
		col := extractColClass(s)
		if col == "" {
			return
		}
		val := strings.TrimSpace(s.Find(".data").First().Text())
		if val == "" {
			return
		}

		label := cleanCaption(s.Find(".caption").First())
		key := normalizeKey(label)
		if sk, ok := colClassStatestreetKey[col]; ok {
			key = sk
		}
		pm.Meta[key] = val
	})

	parseFundHeader(doc, pm)

	if t := holdingsTabDate(doc); !t.IsZero() {
		pm.Date = t
	}
	if pm.Date.IsZero() {
		pm.Date = time.Now()
	}

	return pm, nil
}

func parseFundHeader(doc *goquery.Document, pm *parsedMeta) {
	navLi := doc.Find("li.navAmount").First()
	if navLi.Length() == 0 {
		return
	}
	if nav := strings.TrimSpace(navLi.Find(".header-nav-data").First().Text()); nav != "" {
		pm.Meta[normalizeKey("NAV")] = nav
	}
	if r := strings.TrimSpace(navLi.Find(".fiftyTwoWeekData").First().Text()); r != "" {
		pm.Meta[normalizeKey("52 Week Range")] = strings.Join(strings.Fields(r), " ")
	}
	if ch := strings.TrimSpace(doc.Find("li.navAmountChange .header-nav-data").First().Text()); ch != "" {
		pm.Meta[normalizeKey("1 Day NAV Change")] = strings.Join(strings.Fields(ch), " ")
	}
	if ytd := strings.TrimSpace(doc.Find("li.yearToDate .header-nav-data").First().Text()); ytd != "" {
		pm.Meta[normalizeKey("NAV Total Return YTD")] = ytd
	}
}

func cleanCaption(sel *goquery.Selection) string {
	c := sel.Clone()
	c.Find(".as-of-date, .product-info-bubble, button").Remove()
	return strings.Join(strings.Fields(strings.TrimSpace(c.Text())), " ")
}

func holdingsTabDate(doc *goquery.Document) time.Time {
	sel := doc.Find("#allHoldingsTab select.date-dropdown option[selected]")
	if sel.Length() == 0 {
		sel = doc.Find("#allHoldingsTab select.date-dropdown option").First()
	}
	text := strings.TrimSpace(sel.Text())
	if text == "" {
		return time.Time{}
	}
	t, err := time.Parse("02-Jan-06", text)
	if err != nil {
		return time.Time{}
	}
	return t
}

func holdingsAsOfFromCSV(csv []byte) (time.Time, error) {
	s := string(csv)
	s = strings.TrimPrefix(s, "\ufeff")
	line, _, _ := strings.Cut(s, "\n")
	line = strings.TrimSpace(line)
	const prefix = `Fund Holdings as of,`
	if !strings.HasPrefix(line, prefix) {
		return time.Time{}, fmt.Errorf("unexpected csv first line: %q", line)
	}
	rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	rest = strings.Trim(rest, `"`)
	rest = strings.TrimSpace(rest)
	return time.Parse("02-Jan-06", rest)
}

func extractColClass(s *goquery.Selection) string {
	class, _ := s.Attr("class")
	for _, p := range strings.Fields(class) {
		if strings.HasPrefix(p, "col-") {
			return p
		}
	}
	return ""
}
