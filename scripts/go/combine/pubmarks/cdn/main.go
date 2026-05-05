package cdn

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

func GetYears(ticker string, years []int, dataType string) (map[int]string, error) {
	if len(years) == 0 {
		return map[int]string{}, nil
	}

	// Many calendar years → many parallel GETs; scale timeout (capped) so full-range fetches can complete.
	deadline := max(45*time.Second, time.Duration(len(years))*2*time.Second)
	deadline = min(deadline, 5*time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()

	results := make([]string, len(years))
	g, ctx := errgroup.WithContext(ctx)

	for i, year := range years {
		g.Go(func() error {
			link := fmt.Sprintf("%s/stocks/%s/%d/%s.csv", pagesOrigin, ticker, year, dataType)
			csv, err := downloadCSV(ctx, link)
			if err != nil {
				return err
			}
			results[i] = csv
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	csvTexts := make(map[int]string, len(years))
	for i, year := range years {
		csvTexts[year] = results[i]
	}
	return csvTexts, nil
}

func downloadCSV(ctx context.Context, link string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}
