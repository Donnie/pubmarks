package dateutil

import (
	"testing"
	"time"
)

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	tt, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return tt
}

func TestYearsBefore_Basic(t *testing.T) {
	d := mustTime(t, "2023-07-15T12:34:56Z")
	got := YearsBefore(d, 3)
	want := mustTime(t, "2020-07-15T12:34:56Z")
	if !got.Equal(want) {
		t.Fatalf("got %s want %s", got.Format(time.RFC3339Nano), want.Format(time.RFC3339Nano))
	}
}

func TestYearsBefore_LeapDayClampsToFeb28AndPreservesClock(t *testing.T) {
	d := mustTime(t, "2024-02-29T08:09:10.123456789Z")
	got := YearsBefore(d, 1)

	want := mustTime(t, "2023-02-28T08:09:10.123456789Z")
	if !got.Equal(want) {
		t.Fatalf("got %s want %s", got.Format(time.RFC3339Nano), want.Format(time.RFC3339Nano))
	}
}

func TestYearsBefore_LeapDayToLeapDay(t *testing.T) {
	d := mustTime(t, "2024-02-29T01:02:03Z")
	got := YearsBefore(d, 4)
	want := mustTime(t, "2020-02-29T01:02:03Z")
	if !got.Equal(want) {
		t.Fatalf("got %s want %s", got.Format(time.RFC3339Nano), want.Format(time.RFC3339Nano))
	}
}

func TestYearsBefore_PreservesLocation(t *testing.T) {
	loc := time.FixedZone("X", -7*60*60)
	d := time.Date(2024, time.February, 29, 23, 0, 0, 0, loc)
	got := YearsBefore(d, 1)

	if got.Location() != loc {
		t.Fatalf("location changed: got %v want %v", got.Location(), loc)
	}
	if got.Year() != 2023 || got.Month() != time.February || got.Day() != 28 {
		t.Fatalf("unexpected date: got %s", got.Format(time.RFC3339Nano))
	}
	if got.Hour() != 23 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
		t.Fatalf("clock changed: got %s", got.Format(time.RFC3339Nano))
	}
}
