package csvparse

import (
	"testing"
	"time"

	"donnie.in/sniper360/apps/pubmarks/internal/dateutil"
)

func TestParseCombinedFiveYearWindow_PrefixOutsideWindow(t *testing.T) {
	// Last row sets endDate; only rows with date in [endDate-5y, endDate] are returned.
	// The 1980 line is before the window start and is skipped after parsing its date.
	const csv = `date,open,high,low,close,volume,ttm_net_eps,pe_calc
1980-01-02,1,1,1,0.5,1,1,
2020-01-03,1,1,1,200,1,6,
`
	start, end, last, days, err := ParseCombinedFiveYearWindow(csv)
	if err != nil {
		t.Fatal(err)
	}
	dEnd := mustDate(t, "2020-01-03")
	dStart := dateutil.YearsBefore(dEnd, 5)
	if !start.Equal(dStart) || !end.Equal(dEnd) {
		t.Fatalf("bounds got %s..%s want %s..%s", start, end, dStart, dEnd)
	}
	if last != 200 {
		t.Fatalf("lastPrice %v", last)
	}
	if len(days) != 1 {
		t.Fatalf("want 1 row in window (prefix dropped), got %d: %+v", len(days), days)
	}
	if days[0].Date.Format("2006-01-02") != "2020-01-03" || days[0].Close != 200 || days[0].TtmNetEPS != 6 {
		t.Fatalf("got %+v", days[0])
	}
}

func TestParseCombinedFiveYearWindow_InWindowRowMissingEps_ReturnsError(t *testing.T) {
	const csv = `date,open,high,low,close,volume,ttm_net_eps,pe_calc
2020-01-02,1,1,1,10,100,,
2020-01-03,1,1,1,12,100,2,
`
	_, _, _, _, err := ParseCombinedFiveYearWindow(csv)
	if err == nil {
		t.Fatal("expected error for empty ttm_net_eps in window")
	}
}

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
