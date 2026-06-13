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

var dpDate = time.Date(2026, time.June, 15, 0, 0, 0, 0, time.UTC)

// TestDatePickerTriggerWidth pins the 240px outline-button trigger width.
func TestDatePickerTriggerWidth(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	dp := graft.DatePicker()
	size := uitest.LayoutWidget(dp, 900, 900)
	if size.Width != metrics.DatePicker.TriggerWidth {
		t.Errorf("trigger width = %v, want %v", size.Width, metrics.DatePicker.TriggerWidth)
	}
}

// TestDatePickerPlaceholderMuted verifies the empty-state label is drawn in
// muted-foreground and the value label is drawn in foreground.
func TestDatePickerPlaceholderMuted(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)
	tok := graft.CurrentTheme().Active()

	empty := graft.DatePicker()
	uitest.LayoutWidget(empty, 900, 900)
	mcEmpty := uitest.DrawWidget(empty)
	if !drewTextColor(mcEmpty, "Pick a date", tok.MutedForeground) {
		t.Errorf("placeholder not drawn in muted-foreground")
	}

	filled := graft.DatePicker().Value(dpDate)
	uitest.LayoutWidget(filled, 900, 900)
	mcFilled := uitest.DrawWidget(filled)
	label := dpDate.Format(metrics.DatePicker.DefaultFormat)
	if !drewTextColor(mcFilled, label, tok.Foreground) {
		t.Errorf("value %q not drawn in foreground", label)
	}
}

// drewTextColor reports whether text s was drawn in color col.
func drewTextColor(mc *uitest.MockCanvas, s string, col widget.Color) bool {
	for _, st := range mc.StyledTexts {
		if st.Text == s && st.Style.Color == col {
			return true
		}
	}
	for _, dt := range mc.Texts {
		if dt.Text == s && dt.Color == col {
			return true
		}
	}
	return false
}

// TestDatePickerOpensAndSelects drives the trigger -> overlay -> day-select
// cycle against a fake overlay manager.
func TestDatePickerOpensAndSelects(t *testing.T) {
	if err := graft.LoadAssets(); err != nil {
		t.Fatal(err)
	}
	forceLightMode(t)

	var got time.Time
	dp := graft.DatePicker().Value(dpDate).OnChange(func(d time.Time) { got = d })

	ctx := uitest.NewMockContext()
	om := newFakeOverlayManager()
	ctx.OverlayVal = om

	uitest.LayoutWidget(dp, 800, 600)
	uitest.DrawWidgetWithContext(dp, ctx) // closed

	if len(om.pushed) != 0 {
		t.Fatal("date picker pushed overlay while closed")
	}

	// Click the trigger to open.
	uitest.SimulateClickWithContext(dp, ctx, 10, 10)
	if !dp.IsOpen() {
		t.Fatal("trigger click did not open the popover")
	}
	if len(om.pushed) != 1 {
		t.Fatalf("open pushed %d overlays, want 1", len(om.pushed))
	}

	// Dismiss (click outside / Escape).
	content := om.pushed[0]
	om.onDismiss[content]()
	if dp.IsOpen() {
		t.Fatal("dismiss did not close the popover")
	}
	_ = got
}

func TestGoldenDatePicker(t *testing.T) {
	gtest.GoldenLightDark(t, "datepicker-trigger-empty", func() widget.Widget {
		return primitives.Box(graft.DatePicker()).Padding(24)
	})

	gtest.GoldenLightDark(t, "datepicker-trigger-selected", func() widget.Widget {
		return primitives.Box(graft.DatePicker().Value(dpDate)).Padding(24)
	})

	gtest.GoldenLightDark(t, "datepicker-open", func() widget.Widget {
		content := graft.DatePickerContentPreview(dpDate, dpDate, graft.CurrentTheme())
		return primitives.Box(content).Padding(24)
	})
}
