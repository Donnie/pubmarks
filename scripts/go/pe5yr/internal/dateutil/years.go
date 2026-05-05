package dateutil

import "time"

// YearsBefore returns d minus the given number of calendar years.
// Feb 29 is clamped to Feb 28 in non-leap target years.
func YearsBefore(d time.Time, years int) time.Time {
	target := d.AddDate(-years, 0, 0)
	// AddDate already normalises Feb 29 → Mar 1 in non-leap years;
	// clamp back to Feb 28 to preserve the "same month" invariant.
	if d.Month() == time.February && d.Day() == 29 && target.Day() != 29 {
		return time.Date(
			target.Year(), time.February, 28,
			d.Hour(), d.Minute(), d.Second(), d.Nanosecond(),
			d.Location(),
		)
	}
	return target
}
