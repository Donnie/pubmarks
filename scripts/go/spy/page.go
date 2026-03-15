package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type section struct {
	Fields map[string]string `json:"fields"`
}

type metadata struct {
	Date                 string  `json:"date"`
	FundCharacteristics  section `json:"fund_characteristics"`
	IndexCharacteristics section `json:"index_characteristics"`
	FundMarketPrice      section `json:"fund_market_price"`
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
	var ok bool

	if m.FundCharacteristics, ok = sections["Fund Characteristics"]; !ok {
		return nil, fmt.Errorf("Fund Characteristics section not found")
	}
	if m.IndexCharacteristics, ok = sections["Index Characteristics"]; !ok {
		return nil, fmt.Errorf("Index Characteristics section not found")
	}
	if m.FundMarketPrice, ok = sections["Fund Market Price"]; !ok {
		return nil, fmt.Errorf("Fund Market Price section not found")
	}

	if rawDate != "" {
		// "as of Mar 12 2026" → "2026-03-12"
		trimmed := strings.TrimPrefix(rawDate, "as of ")
		if t, err := time.Parse("Jan 2 2006", trimmed); err == nil {
			m.Date = t.Format("2006-01-02")
		} else {
			m.Date = rawDate
		}
	}

	return m, nil
}

var dateSections = map[string]bool{
	"Fund Characteristics":  true,
	"Index Characteristics": true,
	"Fund Market Price":     true,
}

func extractSections(html string) (map[string]section, string) {
	out := make(map[string]section)
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

		out[title] = section{Fields: parseTable(tableSub[1])}
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
