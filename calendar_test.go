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

// calJune2026 is the fixed month used by the calendar goldens/specs.
var calJune2026 = time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC)

// TestCalendarGridSize pins the 7-column intrinsic width and the full grid
// height (caption row + weekday row + 6 week rows with gaps).
func TestCalendarGridSize(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	cal := graft.Calendar().Month(calJune2026)
	size := uitest.LayoutWidget(cal, 900, 900)

	m := metrics.Calendar
	wantW := m.CellSize * float32(m.Columns)
	if size.Width != wantW {
		t.Errorf("width = %v, want %v (7 cells)", size.Width, wantW)
	}

	wantH := m.CellSize + m.CaptionGap + m.CellSize + m.HeaderGap +
		m.CellSize*float32(m.Rows) + m.WeekGap*float32(m.Rows-1)
	if size.Height != wantH {
		t.Errorf("height = %v, want %v", size.Height, wantH)
	}
}

// TestCalendarSelectedPill verifies the selected day paints a primary pill
// at rounded-md.
func TestCalendarSelectedPill(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	cal := graft.Calendar().
		Month(calJune2026).
		Selected(calJune2026).
		Today(time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC))
	uitest.LayoutWidget(cal, 900, 900)
	mc := uitest.DrawWidget(cal)

	th := graft.CurrentTheme()
	tok := th.Active()

	primaryPills := 0
	for _, rr := range mc.RoundRects {
		if rr.Color == tok.Primary && rr.Radius == th.RadiusMD() &&
			rr.Bounds.Width() == metrics.Calendar.CellSize &&
			rr.Bounds.Height() == metrics.Calendar.CellSize {
			primaryPills++
		}
	}
	if primaryPills != 1 {
		t.Errorf("primary selection pills = %d, want 1", primaryPills)
	}
}

// TestCalendarSelectClick verifies clicking a day fires OnSelect with that
// date.
func TestCalendarSelectClick(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	var got time.Time
	cal := graft.Calendar().
		Month(calJune2026).
		OnSelect(func(d time.Time) { got = d })
	uitest.LayoutWidget(cal, 900, 900)

	// June 1, 2026 is a Monday. With Sunday-start, June 1 is the second cell
	// (col 1) of the first week row. Compute its center.
	m := metrics.Calendar
	firstWeekY := m.CellSize + m.CaptionGap + m.CellSize + m.HeaderGap
	// col 1 (Monday), row 0.
	cx := m.CellSize*1 + m.CellSize/2
	cy := firstWeekY + m.CellSize/2

	uitest.SimulateClick(cal, cx, cy)
	if got.IsZero() {
		t.Fatal("click did not fire OnSelect")
	}
	if got.Day() != 1 || got.Month() != time.June || got.Year() != 2026 {
		t.Errorf("selected = %v, want 2026-06-01", got.Format("2006-01-02"))
	}
}

// TestCalendarOutsideMonthMuted verifies trailing/leading days from adjacent
// months are present (the grid is always 6 rows) and rendered.
func TestCalendarOutsideMonthMuted(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	cal := graft.Calendar().Month(calJune2026)
	uitest.LayoutWidget(cal, 900, 900)
	mc := uitest.DrawWidget(cal)

	// 42 day cells + 7 weekday labels + 1 caption = 50 text draws minimum.
	count := len(mc.StyledTexts) + len(mc.Texts)
	if count < 42+7+1 {
		t.Errorf("text draws = %d, want at least %d", count, 42+7+1)
	}
}

func TestGoldenCalendar(t *testing.T) {
	gtest.GoldenLightDark(t, "calendar-june2026", func() widget.Widget {
		cal := graft.Calendar().
			Month(calJune2026).
			Selected(calJune2026).
			Today(time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC))
		return primitives.Box(cal).Padding(16)
	})
}
