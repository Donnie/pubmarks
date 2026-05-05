package pe

import (
	"math"
	"testing"
	"time"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parse date %q: %v", s, err)
	}
	return d
}

func nearlyEqual(a, b, tol float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	return math.Abs(a-b) <= tol
}

func TestComputeFiveYearPeResult_ConstantPositiveEps(t *testing.T) {
	ticker := "ZZ"
	startDate := mustDate(t, "2020-01-01")
	endDate := mustDate(t, "2020-01-03")

	days := []tradeDay{
		{date: mustDate(t, "2020-01-01"), close: 10, epsTTM: 2},
		{date: mustDate(t, "2020-01-02"), close: 12, epsTTM: 2},
		{date: mustDate(t, "2020-01-03"), close: 14, epsTTM: 2},
	}

	lastPrice := 14.0
	got := computeFiveYearPeResult(ticker, startDate, endDate, lastPrice, days)

	if got.Ticker != ticker {
		t.Fatalf("Ticker: got %q want %q", got.Ticker, ticker)
	}
	if !got.StartDate.Equal(startDate) || !got.EndDate.Equal(endDate) {
		t.Fatalf("dates: got %s..%s want %s..%s", got.StartDate, got.EndDate, startDate, endDate)
	}
	if got.MinPe != 5 || !got.MinPeDate.Equal(mustDate(t, "2020-01-01")) {
		t.Fatalf("MinPe: got %v @ %s want 5 @ 2020-01-01", got.MinPe, got.MinPeDate)
	}
	if got.MaxPe != 7 || !got.MaxPeDate.Equal(mustDate(t, "2020-01-03")) {
		t.Fatalf("MaxPe: got %v @ %s want 7 @ 2020-01-03", got.MaxPe, got.MaxPeDate)
	}
	if got.Mean5yrPe != 6 {
		t.Fatalf("Mean5yrPe: got %v want 6", got.Mean5yrPe)
	}
	if got.Median5yrPe != 6 {
		t.Fatalf("Median5yrPe: got %v want 6", got.Median5yrPe)
	}
	if got.ModePe != 0 {
		t.Fatalf("ModePe: got %v want 0", got.ModePe)
	}
	if got.LatestPe != 7 {
		t.Fatalf("LatestPe: got %v want 7", got.LatestPe)
	}
	if got.LastPrice != 14 {
		t.Fatalf("LastPrice: got %v want 14", got.LastPrice)
	}
	if got.LastEps != 2 {
		t.Fatalf("LastEps: got %v want 2", got.LastEps)
	}
	if got.Mean5yrEps != 2 {
		t.Fatalf("Mean5yrEps: got %v want 2", got.Mean5yrEps)
	}
	if got.Shiller5yrPe != 7 {
		t.Fatalf("Shiller5yrPe: got %v want 7", got.Shiller5yrPe)
	}
	if got.Profitable5yrPe != 6 {
		t.Fatalf("Profitable5yrPe: got %v want 6", got.Profitable5yrPe)
	}
	if got.Lossy5yrPe != 0 {
		t.Fatalf("Lossy5yrPe: got %v want 0", got.Lossy5yrPe)
	}

	meanEY := (0.2 + (2.0 / 12.0) + (2.0 / 14.0)) / 3.0
	wantEYPe := 1.0 / meanEY
	if !nearlyEqual(got.Ey5yrPe, wantEYPe, 1e-12) {
		t.Fatalf("Ey5yrPe: got %.15f want %.15f", got.Ey5yrPe, wantEYPe)
	}
}

func TestComputeFiveYearPeResult_ProfitableAndLossySplit(t *testing.T) {
	ticker := "ZZ"
	startDate := mustDate(t, "2020-01-01")
	endDate := mustDate(t, "2020-01-03")

	days := []tradeDay{
		{date: mustDate(t, "2020-01-01"), close: 10, epsTTM: 2},
		{date: mustDate(t, "2020-01-02"), close: 10, epsTTM: 2},
		{date: mustDate(t, "2020-01-03"), close: 10, epsTTM: -1},
	}

	lastPrice := 10.0
	got := computeFiveYearPeResult(ticker, startDate, endDate, lastPrice, days)

	// Daily P/E: 5, 5, -10.
	if got.Profitable5yrPe != 5 {
		t.Fatalf("Profitable5yrPe: got %v want 5", got.Profitable5yrPe)
	}
	if got.Lossy5yrPe != -10 {
		t.Fatalf("Lossy5yrPe: got %v want -10", got.Lossy5yrPe)
	}
	if got.LatestPe != -10 {
		t.Fatalf("LatestPe: got %v want -10", got.LatestPe)
	}
}
