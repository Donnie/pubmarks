package parse

import (
	"maps"
)

func hydrateMap[M ~map[K]V, K comparable, V any](dst *M, yearCSV map[int]string, parse func(string) (M, error)) error {
	if *dst == nil {
		*dst = make(M)
	}
	for _, csvText := range yearCSV {
		parsed, err := parse(csvText)
		if err != nil {
			return err
		}
		maps.Copy(*dst, parsed)
	}
	return nil
}

func (o *OHLCV) Hydrate(yearCSV map[int]string) error {
	return hydrateMap(o, yearCSV, Price)
}

func (e *EPSTTM) Hydrate(yearCSV map[int]string) error {
	return hydrateMap(e, yearCSV, Peratio)
}
