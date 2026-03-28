package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Parity with scripts/go/statestreet JSON (datasets/etfs/*/latest.json):
//
// State Street keys we set from BlackRock when the data matches:
//   number_of_holdings   ← Number of Holdings (col-numHoldings)
//   price_book_ratio     ← P/B Ratio (col-priceBook)
//   price_earnings       ← P/E Ratio (col-priceEarnings). BlackRock defines this as
//                        trailing: latest price / latest 12m EPS (caps at 60). This is
//                        NOT forward FY1 P/E; see price_earnings_fw below.
//
// State Street keys not available on typical BlackRock EMEA ETF pages (omitted):
//   30_day_median_bid_ask_spread, bid_ask, closing_price, day_high, day_low,
//   est_3_5_year_eps_growth, exchange_volume_shares, premium_discount,
//   price_cash_flow, price_earnings_fw, weighted_average_market_cap
//
// Additional keys emitted only by this scraper (br_*), BlackRock-sourced:
//   br_net_assets, br_inception_date, br_share_class_currency, br_asset_class,
//   br_sfdr_classification, br_total_expense_ratio, br_use_of_income, br_domicile,
//   br_rebalance_frequency, br_ucits, br_fund_manager, br_custodian,
//   br_bloomberg_ticker, br_net_assets_fund_level, br_fund_launch_date, br_base_currency,
//   br_benchmark_index, br_shares_outstanding, br_isin, br_securities_lending_return,
//   br_product_structure, br_methodology, br_issuing_company, br_administrator,
//   br_fiscal_year_end, br_benchmark_ticker, br_3y_beta, br_benchmark_level,
//   br_standard_deviation_3y, br_nav, br_nav_label, br_52_week_range,
//   br_1_day_nav_change, br_nav_total_return_ytd

// colClassToStatestreet maps BlackRock column classes to statestreet top-level keys.
var colClassToStatestreet = map[string]string{
	"col-numHoldings":   "number_of_holdings",
	"col-priceBook":     "price_book_ratio",
	"col-priceEarnings": "price_earnings",
}

// colClassToBR maps BlackRock column classes to br_* output keys (extras).
var colClassToBR = map[string]string{
	"col-totalNetAssets":           "br_net_assets",
	"col-inceptionDate":            "br_inception_date",
	"col-seriesBaseCurrencyCode":   "br_share_class_currency",
	"col-assetClass":               "br_asset_class",
	"col-sfdr":                     "br_sfdr_classification",
	"col-emeaMgt":                  "br_total_expense_ratio",
	"col-useOfProfitsCode":         "br_use_of_income",
	"col-domicile":                 "br_domicile",
	"col-rebalanceFrequency":       "br_rebalance_frequency",
	"col-ucitsCompliantFlag":       "br_ucits",
	"col-fundmanager":              "br_fund_manager",
	"col-fundCustodian":            "br_custodian",
	"col-bbeqtick":                 "br_bloomberg_ticker",
	"col-totalNetAssetsFundLevel":  "br_net_assets_fund_level",
	"col-launchDate":               "br_fund_launch_date",
	"col-baseCurrencyCode":         "br_base_currency",
	"col-indexSeriesName":          "br_benchmark_index",
	"col-sharesOutstanding":        "br_shares_outstanding",
	"col-isin":                     "br_isin",
	"col-oneYearSecLendingReturn":  "br_securities_lending_return",
	"col-productStructure":         "br_product_structure",
	"col-fundMethodologyTypeCode":  "br_methodology",
	"col-issuingCompany":           "br_issuing_company",
	"col-fundAdministrator":        "br_administrator",
	"col-fiscalYearEndDate":        "br_fiscal_year_end",
	"col-benchmarkTicker":          "br_benchmark_ticker",
	"col-threeYrBetaFund":          "br_3y_beta",
	"col-levelAmount":              "br_benchmark_level",
	"col-volatilitySourced3YrAnnualized": "br_standard_deviation_3y",
}

type parsedMeta struct {
	Date        time.Time
	Statestreet map[string]string
	Blackrock   map[string]string
}

func parseProductPage(doc *goquery.Document) (*parsedMeta, error) {
	pm := &parsedMeta{
		Statestreet: make(map[string]string),
		Blackrock:   make(map[string]string),
	}

	doc.Find("#keyFundFacts .product-data-item, #fundamentalsAndRisk .product-data-item").Each(func(_ int, s *goquery.Selection) {
		col := extractColClass(s)
		if col == "" {
			return
		}
		val := strings.TrimSpace(s.Find(".data").First().Text())
		if val == "" {
			return
		}

		if sk, ok := colClassToStatestreet[col]; ok {
			pm.Statestreet[sk] = val
			return
		}
		if bk, ok := colClassToBR[col]; ok {
			pm.Blackrock[bk] = val
			return
		}
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
	label := strings.TrimSpace(navLi.Find(".header-nav-label").First().Text())
	nav := strings.TrimSpace(navLi.Find(".header-nav-data").First().Text())
	if nav != "" {
		pm.Blackrock["br_nav"] = nav
	}
	if label != "" {
		pm.Blackrock["br_nav_label"] = strings.Join(strings.Fields(label), " ")
	}
	if r := strings.TrimSpace(navLi.Find(".fiftyTwoWeekData").First().Text()); r != "" {
		pm.Blackrock["br_52_week_range"] = strings.Join(strings.Fields(r), " ")
	}

	if ch := strings.TrimSpace(doc.Find("li.navAmountChange .header-nav-data").First().Text()); ch != "" {
		pm.Blackrock["br_1_day_nav_change"] = strings.Join(strings.Fields(ch), " ")
	}
	if ytd := strings.TrimSpace(doc.Find("li.yearToDate .header-nav-data").First().Text()); ytd != "" {
		pm.Blackrock["br_nav_total_return_ytd"] = ytd
	}
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
	// Fund Holdings as of,"26-Mar-26"
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
