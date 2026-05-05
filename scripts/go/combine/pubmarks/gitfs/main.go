package gitfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sync/errgroup"
)

func GetYears(ticker string, years []int, dataType string) (map[int]string, error) {
	if len(years) == 0 {
		return map[int]string{}, nil
	}

	ds, err := ResolveDatasetsDir()
	if err != nil {
		return nil, err
	}
	t := strings.ToLower(strings.TrimSpace(ticker))
	base := filepath.Join(ds, "stocks", t)

	results := make([]string, len(years))
	var g errgroup.Group

	for i, year := range years {
		i, year := i, year
		g.Go(func() error {
			p := filepath.Join(base, strconv.Itoa(year), dataType+".csv")
			b, err := os.ReadFile(p)
			if err != nil {
				return fmt.Errorf("read %q: %w", p, err)
			}
			results[i] = string(b)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	out := make(map[int]string, len(years))
	for i, year := range years {
		out[year] = results[i]
	}
	return out, nil
}
