package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const userAgent = "Mozilla/5.0 (compatible; pubmarks/1.0)"

var httpClient = &http.Client{Timeout: 30 * time.Second}

// fetchPage fetches the page and returns the parsed document.
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

// holdingsXLSXURL finds the holdings-daily .xlsx link in the document and returns its absolute URL.
func holdingsXLSXURL(doc *goquery.Document, pageURLStr string) (string, error) {
	base, err := url.Parse(pageURLStr)
	if err != nil {
		return "", fmt.Errorf("parse page URL: %w", err)
	}

	var xlsxURL string
	doc.Find("a[href]").EachWithBreak(func(_ int, a *goquery.Selection) bool {
		href, _ := a.Attr("href")
		href = strings.TrimSpace(href)
		lower := strings.ToLower(href)
		if strings.HasSuffix(lower, ".xlsx") && strings.Contains(lower, "holdings-daily") {
			if ref, err := url.Parse(href); err == nil {
				xlsxURL = base.ResolveReference(ref).String()
			}
			return false
		}
		return true
	})

	if xlsxURL == "" {
		return "", fmt.Errorf("no holdings-daily .xlsx link found in page")
	}
	return xlsxURL, nil
}

func downloadXLSX(xlsxURL string) (path string, cleanup func(), err error) {
	req, err := http.NewRequest("GET", xlsxURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	tmp, err := os.CreateTemp("", "ssga-*.xlsx")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp file: %w", err)
	}

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", nil, fmt.Errorf("writing temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", nil, fmt.Errorf("closing temp file: %w", err)
	}

	remove := func() { os.Remove(tmp.Name()) }
	return tmp.Name(), remove, nil
}
