package graft_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// clickCalendarDay clicks the in-month day-of-month d on a June 2026 calendar
// (Sunday week start). June 1, 2026 is a Monday, so the grid's first row is
// [May31, Jun1..Jun6] and day n sits at flat index n (0-based: May31=0).
func clickCalendarDay(cal *graft.CalendarWidget, day int) {
	m := metrics.Calendar
	firstWeekY := m.CellSize + m.CaptionGap + m.CellSize + m.HeaderGap
	// June 1 is at flat index 1 (col 1, row 0). Day n is at flat index n.
	idx := day
	row := idx / m.Columns
	col := idx % m.Columns
	cx := float32(col)*m.CellSize + m.CellSize/2
	cy := firstWeekY + float32(row)*(m.CellSize+m.WeekGap) + m.CellSize/2
	uitest.SimulateClick(cal, cx, cy)
}

// TestCalendarRangeFirstClick verifies the first click sets the start and
// clears any prior end; no complete range yet.
func TestCalendarRangeFirstClick(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	cal := graft.Calendar().Range().Month(calJune2026)
	uitest.LayoutWidget(cal, 900, 900)

	clickCalendarDay(cal, 5)
	start, _, complete := cal.SelectedRangeValue()
	if complete {
		t.Fatal("range complete after a single click")
	}
	if start.Day() != 5 || start.Month() != time.June {
		t.Errorf("start = %v, want June 5", start.Format("2006-01-02"))
	}
}

// TestCalendarRangeSecondClick verifies a forward second click closes the
// range and fires OnRangeChange.
func TestCalendarRangeSecondClick(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	var gotS, gotE time.Time
	var fired int
	cal := graft.Calendar().Range().Month(calJune2026).
		OnRangeChange(func(s, e time.Time) { gotS, gotE = s, e; fired++ })
	uitest.LayoutWidget(cal, 900, 900)

	clickCalendarDay(cal, 5)
	clickCalendarDay(cal, 12)

	if fired != 1 {
		t.Fatalf("OnRangeChange fired %d times, want 1", fired)
	}
	s, e, complete := cal.SelectedRangeValue()
	if !complete {
		t.Fatal("range not complete after two clicks")
	}
	if s.Day() != 5 || e.Day() != 12 {
		t.Errorf("range = %v..%v, want 5..12", s.Day(), e.Day())
	}
	if gotS.Day() != 5 || gotE.Day() != 12 {
		t.Errorf("callback range = %v..%v, want 5..12", gotS.Day(), gotE.Day())
	}
}

// TestCalendarRangeSwap verifies a second click BEFORE the start swaps the
// endpoints so start <= end.
func TestCalendarRangeSwap(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	cal := graft.Calendar().Range().Month(calJune2026)
	uitest.LayoutWidget(cal, 900, 900)

	clickCalendarDay(cal, 12)
	clickCalendarDay(cal, 5)

	s, e, complete := cal.SelectedRangeValue()
	if !complete {
		t.Fatal("range not complete after two clicks")
	}
	if s.Day() != 5 || e.Day() != 12 {
		t.Errorf("range = %v..%v, want swapped 5..12", s.Day(), e.Day())
	}
}

// TestCalendarRangeRestart verifies a third click (range already complete)
// restarts the range at the new start.
func TestCalendarRangeRestart(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	cal := graft.Calendar().Range().Month(calJune2026)
	uitest.LayoutWidget(cal, 900, 900)

	clickCalendarDay(cal, 5)
	clickCalendarDay(cal, 12)
	clickCalendarDay(cal, 20) // third click restarts

	start, _, complete := cal.SelectedRangeValue()
	if complete {
		t.Fatal("range should be incomplete after restart")
	}
	if start.Day() != 20 {
		t.Errorf("restart start = %v, want 20", start.Day())
	}
}

// TestCalendarRangeBand verifies the in-between days draw an --accent square
// band and the two ends draw --primary pills.
func TestCalendarRangeBand(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	start := time.Date(2026, time.June, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.June, 11, 0, 0, 0, 0, time.UTC)
	cal := graft.Calendar().
		SelectedRange(start, end).
		Month(calJune2026).
		Today(time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC))
	uitest.LayoutWidget(cal, 900, 900)
	mc := uitest.DrawWidget(cal)

	th := graft.CurrentTheme()
	tok := th.Active()

	// Two --primary pills (start + end).
	pills := 0
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Primary && rr.Radius == th.RadiusMD() &&
			rr.Bounds.Width() == metrics.Calendar.CellSize {
			pills++
		}
	}
	if pills != 2 {
		t.Errorf("primary range-end pills = %d, want 2", pills)
	}

	// Full-cell accent squares for the two in-between days (June 9, 10).
	bands := 0
	for _, r := range mc.Rects {
		if r.Color == tok.Accent &&
			r.Bounds.Width() == metrics.Calendar.CellSize &&
			r.Bounds.Height() == metrics.Calendar.CellSize {
			bands++
		}
	}
	if bands != 2 {
		t.Errorf("full-cell accent bands = %d, want 2 (days 9,10)", bands)
	}
}

// TestCalendarMultipleToggle verifies clicking toggles each day in and out of
// the selected set and fires OnDatesChange with the full set.
func TestCalendarMultipleToggle(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	var last []time.Time
	cal := graft.Calendar().Multiple().Month(calJune2026).
		OnDatesChange(func(d []time.Time) { last = d })
	uitest.LayoutWidget(cal, 900, 900)

	clickCalendarDay(cal, 3)
	clickCalendarDay(cal, 7)
	if got := len(cal.SelectedDatesValue()); got != 2 {
		t.Fatalf("after two selects, set size = %d, want 2", got)
	}
	if len(last) != 2 {
		t.Errorf("callback set size = %d, want 2", len(last))
	}

	clickCalendarDay(cal, 3) // toggle off
	set := cal.SelectedDatesValue()
	if len(set) != 1 {
		t.Fatalf("after toggle-off, set size = %d, want 1", len(set))
	}
	if set[0].Day() != 7 {
		t.Errorf("remaining day = %v, want 7", set[0].Day())
	}
}

// TestCalendarMultiplePreset verifies SelectedDates de-duplicates and renders
// one --primary pill per distinct day.
func TestCalendarMultiplePreset(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	d := func(day int) time.Time {
		return time.Date(2026, time.June, day, 0, 0, 0, 0, time.UTC)
	}
	cal := graft.Calendar().
		SelectedDates(d(3), d(7), d(7), d(15)).
		Month(calJune2026).
		Today(d(1))
	uitest.LayoutWidget(cal, 900, 900)
	mc := uitest.DrawWidget(cal)

	if got := len(cal.SelectedDatesValue()); got != 3 {
		t.Errorf("deduped set size = %d, want 3", got)
	}

	th := graft.CurrentTheme()
	tok := th.Active()
	pills := 0
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Primary && rr.Radius == th.RadiusMD() &&
			rr.Bounds.Width() == metrics.Calendar.CellSize {
			pills++
		}
	}
	if pills != 3 {
		t.Errorf("primary pills = %d, want 3", pills)
	}
}

func TestGoldenCalendarRange(t *testing.T) {
	start := time.Date(2026, time.June, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.June, 16, 0, 0, 0, 0, time.UTC)
	gtest.GoldenLightDark(t, "calendar-range-june2026", func() widget.Widget {
		cal := graft.Calendar().
			SelectedRange(start, end).
			Month(calJune2026).
			Today(time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC))
		return primitives.Box(cal).Padding(16)
	})
}

func TestGoldenCalendarMultiple(t *testing.T) {
	d := func(day int) time.Time {
		return time.Date(2026, time.June, day, 0, 0, 0, 0, time.UTC)
	}
	gtest.GoldenLightDark(t, "calendar-multiple-june2026", func() widget.Widget {
		cal := graft.Calendar().
			SelectedDates(d(3), d(11), d(19), d(27)).
			Month(calJune2026).
			Today(d(1))
		return primitives.Box(cal).Padding(16)
	})
}

func TestGoldenDateRangePicker(t *testing.T) {
	start := time.Date(2026, time.June, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.June, 16, 0, 0, 0, 0, time.UTC)
	gtest.GoldenLightDark(t, "daterangepicker-open", func() widget.Widget {
		content := graft.DateRangePickerContentPreview(calJune2026, start, end, graft.CurrentTheme())
		return primitives.Box(content).Padding(24)
	})
}
