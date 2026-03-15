package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

// holdingsXLSXURL finds the holdings-daily .xlsx link in the page body and returns its absolute URL.
func holdingsXLSXURL(body []byte, pageURLStr string) (string, error) {
	for _, href := range findXLSXLinks(body) {
		if strings.Contains(strings.ToLower(href), "holdings-daily") {
			base, err := url.Parse(pageURLStr)
			if err != nil {
				return "", fmt.Errorf("parse page URL: %w", err)
			}
			ref, err := url.Parse(strings.TrimSpace(href))
			if err != nil {
				return "", fmt.Errorf("parse xlsx href: %w", err)
			}
			return base.ResolveReference(ref).String(), nil
		}
	}
	return "", fmt.Errorf("no holdings-daily .xlsx link found in page")
}

var reHrefXLSX = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']*\.xlsx[^"']*)["']`)

func findXLSXLinks(html []byte) []string {
	seen := make(map[string]bool)
	var out []string
	for _, m := range reHrefXLSX.FindAllSubmatch(html, -1) {
		if len(m) < 2 {
			continue
		}
		s := strings.TrimSpace(string(m[1]))
		s = strings.ReplaceAll(s, "&amp;", "&")
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func downloadXLSX(xlsxURL string) (path string, cleanup func(), err error) {
	req, err := http.NewRequest("GET", xlsxURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; pubmarks/1.0)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("reading body: %w", err)
	}

	tmp, err := os.CreateTemp("", "ssga-*.xlsx")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp file: %w", err)
	}

	if _, err := tmp.Write(data); err != nil {
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

// fetchPage returns the page body and the URL it was fetched from (for resolving relative links).
func fetchPage(pageURL string) (body []byte, err error) {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; pubmarks/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	return body, nil
}
