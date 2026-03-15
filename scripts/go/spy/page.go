package main

import (
	"regexp"
	"strings"
	"time"
)

type metadata struct {
	Date                 time.Time         `json:"date"`
	FundCharacteristics  map[string]string `json:"fund_characteristics"`
	IndexCharacteristics map[string]string `json:"index_characteristics"`
	FundMarketPrice      map[string]string `json:"fund_market_price"`
}

var (
	reSectionHeader = regexp.MustCompile(`(?s)<h2 class="comp-title">(.*?)</h2>`)
	reDateSpan      = regexp.MustCompile(`<span class="date">(.*?)</span>`)
	reTable         = regexp.MustCompile(`(?s)<table class="tb-keyvalue">(.*?)</table>`)
	reRow           = regexp.MustCompile(`(?s)<td class="label">(.*?)</td>\s*<td class="data">(.*?)</td>`)
	reTags          = regexp.MustCompile(`<[^>]+>`)
	reInfoDiv       = regexp.MustCompile(`(?s)<div class="info-data">.*?</div>`)
	reEntities      = strings.NewReplacer(
		"&#xfeff;", "",
		"&#34;", `"`,
		"&#39;", "'",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
	)
)

func parsePage(body []byte) (*metadata, error) {
	sections, rawDate := extractSections(string(body))

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

var dateSections = map[string]bool{
	"Fund Characteristics":  true,
	"Index Characteristics": true,
	"Fund Market Price":     true,
}

func extractSections(html string) (map[string]map[string]string, string) {
	out := make(map[string]map[string]string)
	rawDate := ""
	headers := reSectionHeader.FindAllStringIndex(html, -1)
	for i, loc := range headers {
		headerHTML := html[loc[0]:loc[1]]

		titleHTML := reDateSpan.ReplaceAllString(headerHTML, "")
		title := stripTags(titleHTML)

		if rawDate == "" && dateSections[title] {
			if ds := reDateSpan.FindStringSubmatch(headerHTML); ds != nil {
				rawDate = strings.TrimSpace(ds[1])
			}
		}

		searchFrom := loc[1]
		searchTo := len(html)
		if i+1 < len(headers) {
			searchTo = headers[i+1][0]
		}

		tableSub := reTable.FindStringSubmatch(html[searchFrom:searchTo])
		if tableSub == nil {
			continue
		}

		out[title] = parseTable(tableSub[1])
	}
	return out, rawDate
}

func parseTable(tableBody string) map[string]string {
	fields := make(map[string]string)
	for _, row := range reRow.FindAllStringSubmatch(tableBody, -1) {
		label := cleanLabel(row[1])
		value := reEntities.Replace(stripTags(row[2]))
		if label != "" {
			fields[label] = value
		}
	}
	return fields
}

func cleanLabel(s string) string {
	s = reInfoDiv.ReplaceAllString(s, "")
	return reEntities.Replace(stripTags(s))
}

func stripTags(s string) string {
	s = reTags.ReplaceAllString(s, "")
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}
