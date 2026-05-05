package pe

import (
	"math"
	"sort"
)

// mean returns the arithmetic mean of values.
func mean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// median returns the median of values, sorting a copy to preserve the original order.
func median(values []float64) float64 {
	cp := make([]float64, len(values))
	copy(cp, values)
	sort.Float64s(cp)

	n := len(cp)
	if n%2 == 1 {
		return cp[n/2]
	}
	return (cp[n/2-1] + cp[n/2]) / 2
}

// modeIntegerBucket returns the modal P/E after rounding each value to the nearest integer.
// Returns 0 when no integer bucket has count > 1.
// On ties, returns the median of the tied bucket centres.
func modeIntegerBucket(values []float64) float64 {
	counts := make(map[float64]int, len(values))
	for _, v := range values {
		counts[math.Round(v)]++
	}

	best := 0
	for _, n := range counts {
		if n > best {
			best = n
		}
	}
	if best <= 1 {
		return 0
	}

	tops := make([]float64, 0)
	for bucket, n := range counts {
		if n == best {
			tops = append(tops, bucket)
		}
	}
	sort.Float64s(tops)

	return median(tops)
}
