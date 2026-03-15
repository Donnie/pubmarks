package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func downloadXLSX(ticker string) (path string, cleanup func(), err error) {
	urlFmt := os.Getenv("SOURCE_URL")
	if urlFmt == "" {
		return "", nil, fmt.Errorf("SOURCE_URL is not set")
	}
	url := fmt.Sprintf(urlFmt, strings.ToLower(ticker))

	resp, err := http.Get(url)
	if err != nil {
		return "", nil, fmt.Errorf("http.Get: %w", err)
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
