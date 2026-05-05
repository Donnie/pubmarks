package pe

import (
	"math"
	"testing"
)

func TestMean(t *testing.T) {
	got := mean([]float64{1, 2, 3, 4})
	if got != 2.5 {
		t.Fatalf("got %v want 2.5", got)
	}
}

func TestMedian_OddEvenAndPreservesInputOrder(t *testing.T) {
	odd := []float64{9, 1, 5}
	if got := median(odd); got != 5 {
		t.Fatalf("odd: got %v want 5", got)
	}
	// Ensure median sorts a copy, not the input.
	if odd[0] != 9 || odd[1] != 1 || odd[2] != 5 {
		t.Fatalf("odd input mutated: %v", odd)
	}

	even := []float64{10, 2, 8, 4} // sorted => [2,4,8,10], median => (4+8)/2 = 6
	if got := median(even); got != 6 {
		t.Fatalf("even: got %v want 6", got)
	}
	if even[0] != 10 || even[1] != 2 || even[2] != 8 || even[3] != 4 {
		t.Fatalf("even input mutated: %v", even)
	}
}

func TestModeIntegerBucket_NoModeWhenAllBucketsUnique(t *testing.T) {
	// Each rounds to a different integer bucket: 1,2,3.
	got := modeIntegerBucket([]float64{1.1, 2.1, 3.1})
	if got != 0 {
		t.Fatalf("got %v want 0", got)
	}
}

func TestModeIntegerBucket_ReturnsBucketForClearWinner(t *testing.T) {
	// Rounds to: 10, 10, 11, 12 -> mode is 10.
	got := modeIntegerBucket([]float64{9.6, 10.4, 10.6, 12.4})
	if got != 10 {
		t.Fatalf("got %v want 10", got)
	}
}

func TestModeIntegerBucket_TiesReturnMedianOfBucketCentres(t *testing.T) {
	t.Run("two-way-tie", func(t *testing.T) {
		// Rounds to: 10,10, 12,12 -> tie between 10 and 12 => median is (10+12)/2 = 11.
		got := modeIntegerBucket([]float64{9.6, 10.4, 11.6, 12.4})
		if got != 11 {
			t.Fatalf("got %v want 11", got)
		}
	})

	t.Run("three-way-tie", func(t *testing.T) {
		// Rounds to: 10,10, 11,11, 12,12 -> tie among 10,11,12 => median is 11.
		got := modeIntegerBucket([]float64{9.6, 10.4, 10.6, 11.4, 11.6, 12.4})
		if got != 11 {
			t.Fatalf("got %v want 11", got)
		}
	})
}

func TestModeIntegerBucket_RoundHalfAwayFromZero(t *testing.T) {
	// Guard the intended behavior against a change in rounding method.
	if math.Round(2.5) != 3 || math.Round(-2.5) != -3 {
		t.Fatalf("unexpected math.Round behavior: Round(2.5)=%v Round(-2.5)=%v", math.Round(2.5), math.Round(-2.5))
	}

	// Buckets: 3,3,-3,-3 => tie between -3 and 3 => median is 0.
	got := modeIntegerBucket([]float64{2.5, 2.51, -2.5, -2.51})
	if got != 0 {
		t.Fatalf("got %v want 0", got)
	}
}
