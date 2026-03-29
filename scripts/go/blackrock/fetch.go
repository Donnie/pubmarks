package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const userAgent = "Mozilla/5.0 (compatible; pubmarks/1.0)"

var httpClient = &http.Client{Timeout: 60 * time.Second}

func fetchPage(pageURL string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parsing page HTML: %w", err)
	}
	return doc, nil
}

func fetchBytes(target string) ([]byte, error) {
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	return b, nil
}

// holdingsJSONURI returns absolute URL for all-holdings JSON from #allHoldingsTab[data-ajaxuri].
func holdingsJSONURI(doc *goquery.Document, pageURL string) (string, error) {
	base, err := url.Parse(pageURL)
	if err != nil {
		return "", fmt.Errorf("parse page URL: %w", err)
	}

	raw, exists := doc.Find("#allHoldingsTab").Attr("data-ajaxuri")
	if !exists || strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("no #allHoldingsTab data-ajaxuri")
	}
	raw = strings.TrimSpace(raw)
	ref, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse ajax uri: %w", err)
	}
	return base.ResolveReference(ref).String(), nil
}

// holdingsCSVURI finds the Detailed Holdings CSV export link near holdings.
func holdingsCSVURI(doc *goquery.Document, pageURL string) (string, error) {
	base, err := url.Parse(pageURL)
	if err != nil {
		return "", fmt.Errorf("parse page URL: %w", err)
	}

	var href string
	doc.Find("#holdings a.icon-xls-export[href]").EachWithBreak(func(_ int, a *goquery.Selection) bool {
		h, _ := a.Attr("href")
		h = strings.TrimSpace(h)
		if strings.Contains(h, "fileType=csv") && strings.Contains(h, "dataType=fund") {
			href = h
			return false
		}
		return true
	})

	if href == "" {
		return "", fmt.Errorf("no holdings CSV export link")
	}
	ref, err := url.Parse(href)
	if err != nil {
		return "", fmt.Errorf("parse csv href: %w", err)
	}
	return base.ResolveReference(ref).String(), nil
}
