package cdn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

var ErrNotFound = errors.New("cdn: HTTP 404 (resource not found)")

// pagesOrigin is the pubmarks data site (GitHub Pages: Donnie/pubmarks, branch gh-pages).
const pagesOrigin = "https://donnie.github.io/pubmarks"

// manifestPath is the path on pagesOrigin for manifest.json (leading slash).
const manifestPath = "/manifest.json"

// manifestURL is pagesOrigin + manifestPath (used by manifest schema tests).
const manifestURL = pagesOrigin + manifestPath

type Client struct{ *http.Client }

func NewClient() *Client {
	return &Client{&http.Client{Timeout: 30 * time.Second}}
}

// FetchCombinedCSV loads stocks/{ticker}/combined.csv (full-history OHLCV + TTM EPS).
func (c *Client) FetchCombinedCSV(ticker string) (string, error) {
	u := fmt.Sprintf("%s/stocks/%s/combined.csv", pagesOrigin, strings.ToLower(ticker))
	return c.get(u)
}

func (c *Client) get(url string) (string, error) {
	const n = 3
	var err error
	for a := 1; a <= n; a++ {
		req, e := http.NewRequest(http.MethodGet, url, nil)
		if e != nil {
			return "", e
		}
		resp, e := c.Do(req)
		if e != nil {
			err = e
			if a < n && retryable(e) {
				time.Sleep(time.Duration(a) * 500 * time.Millisecond)
				continue
			}
			return "", fmt.Errorf("GET %s: %w", url, e)
		}
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return "", fmt.Errorf("GET %s: %w", url, ErrNotFound)
		}
		if resp.StatusCode >= 500 && a < n {
			resp.Body.Close()
			err = fmt.Errorf("GET %s -> HTTP %d", url, resp.StatusCode)
			time.Sleep(time.Duration(a) * 500 * time.Millisecond)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return "", fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
		}
		body, e := io.ReadAll(resp.Body)
		resp.Body.Close()
		if e != nil {
			err = fmt.Errorf("reading %s: %w", url, e)
			if a < n && retryable(e) {
				time.Sleep(time.Duration(a) * 500 * time.Millisecond)
				continue
			}
			return "", err
		}
		return string(body), nil
	}
	return "", err
}

func retryable(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}
