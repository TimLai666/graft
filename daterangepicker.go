package graft

import (
	"time"

	corepopover "github.com/gogpu/ui/core/popover"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// DateRangePickerWidget is the range variant of the shadcn DatePicker: an
// outline Button trigger that opens a Popover hosting a range-mode Calendar.
// The trigger shows "start – end" once a complete range is picked, the start
// date while mid-selection, or the placeholder when empty. Selecting the end
// of a range closes the popover (mirrors DatePickerWidget plumbing).
//
// Architecture decision: graft-owned composite, reusing ButtonWidget for the
// trigger and CalendarWidget (range mode) for the surface, with the same
// overlay lifecycle as DatePickerWidget/PopoverWidget.
type DateRangePickerWidget struct {
	widget.WidgetBase

	start       *time.Time
	end         *time.Time
	placeholder string
	format      string
	separator   string
	onChange    func(start, end time.Time)

	open  state.Signal[bool]
	shown bool
	om    widget.OverlayManager

	trigger *ButtonWidget
	content *dateRangePickerContent

	theme *theme.Theme
}

// DateRangePicker builds a date-range picker with the default placeholder.
func DateRangePicker() *DateRangePickerWidget {
	d := &DateRangePickerWidget{
		placeholder: "Pick a date range",
		format:      metrics.DatePicker.DefaultFormat,
		separator:   " – ",
		open:        state.NewSignal(false),
		theme:       CurrentTheme(),
	}
	d.SetVisible(true)
	d.SetEnabled(true)
	d.rebuildTrigger()
	d.content = newDateRangePickerContent(d)
	return d
}

// DatePickerRange is a fluent alias matching the spec's DatePicker().Range()
// suggestion: it returns a fresh range picker.
func (d *DatePickerWidget) Range() *DateRangePickerWidget {
	rp := DateRangePicker().Theme(d.resolvedTheme())
	if d.value != nil {
		rp.placeholder = d.placeholder
	}
	rp.format = d.format
	rp.rebuildTrigger()
	return rp
}

// Value presets the selected range (normalized so start <= end).
func (d *DateRangePickerWidget) Value(start, end time.Time) *DateRangePickerWidget {
	s := truncateDay(start)
	e := truncateDay(end)
	if e.Before(s) {
		s, e = e, s
	}
	d.start = &s
	d.end = &e
	d.rebuildTrigger()
	if d.content != nil {
		d.content.cal.SelectedRange(s, e).Month(s)
	}
	return d
}

// Placeholder sets the empty-state label.
func (d *DateRangePickerWidget) Placeholder(s string) *DateRangePickerWidget {
	d.placeholder = s
	d.rebuildTrigger()
	return d
}

// Format sets the Go time layout used to render each endpoint.
func (d *DateRangePickerWidget) Format(layout string) *DateRangePickerWidget {
	d.format = layout
	d.rebuildTrigger()
	return d
}

// Separator sets the string drawn between the start and end dates.
func (d *DateRangePickerWidget) Separator(s string) *DateRangePickerWidget {
	d.separator = s
	d.rebuildTrigger()
	return d
}

// OnChange registers the range-selection observer (fires on a complete range).
func (d *DateRangePickerWidget) OnChange(fn func(start, end time.Time)) *DateRangePickerWidget {
	d.onChange = fn
	return d
}

// Theme pins a specific theme.
func (d *DateRangePickerWidget) Theme(th *theme.Theme) *DateRangePickerWidget {
	if th != nil {
		d.theme = th
		d.rebuildTrigger()
		if d.content != nil {
			d.content.cal.Theme(th)
		}
	}
	return d
}

func (d *DateRangePickerWidget) resolvedTheme() *theme.Theme {
	if d.theme != nil {
		return d.theme
	}
	return CurrentTheme()
}

// triggerLabel returns "start – end", the start alone, or the placeholder.
func (d *DateRangePickerWidget) triggerLabel() string {
	switch {
	case d.start != nil && d.end != nil:
		return d.start.Format(d.format) + d.separator + d.end.Format(d.format)
	case d.start != nil:
		return d.start.Format(d.format)
	default:
		return d.placeholder
	}
}

// rebuildTrigger rebuilds the outline button trigger.
func (d *DateRangePickerWidget) rebuildTrigger() {
	btn := Button(d.triggerLabel()).
		Outline().
		Icon(icons.Calendar).
		W(metrics.DatePicker.TriggerWidth).
		Theme(d.resolvedTheme()).
		OnClick(func() { d.open.Set(!d.open.Get()) })
	if d.start == nil {
		tok := d.resolvedTheme().Active()
		mc := tok.MutedForeground
		btn.Style(func(s *Style) { s.Foreground = &mc })
	}
	d.trigger = btn
	ovlSetParent(d.trigger, d)
}

// Layout sizes the host to the trigger.
func (d *DateRangePickerWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	size := d.trigger.Layout(ctx, c)
	setWidgetBounds(d.trigger, geometry.NewRect(0, 0, size.Width, size.Height))
	d.SetBounds(geometry.FromPointSize(d.Position(), size))
	return size
}

// Draw renders the trigger and reconciles the overlay with the open signal.
func (d *DateRangePickerWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !d.IsVisible() {
		return
	}
	bounds := d.Bounds()
	canvas.PushTransform(bounds.Min)
	widget.StampScreenOrigin(d.trigger, canvas)
	widget.DrawChild(d.trigger, ctx, canvas)
	canvas.PopTransform()
	d.syncOverlay(ctx)
}

// Event forwards to the trigger button.
func (d *DateRangePickerWidget) Event(ctx widget.Context, e event.Event) bool {
	offset := d.Bounds().Min
	consumed := d.trigger.Event(ctx, ovlTranslate(e, offset))
	d.syncOverlay(ctx)
	return consumed
}

// Children returns the inline trigger.
func (d *DateRangePickerWidget) Children() []widget.Widget {
	return []widget.Widget{d.trigger}
}

// Mount binds the open signal.
func (d *DateRangePickerWidget) Mount(ctx widget.Context) {
	if sched := ctx.Scheduler(); sched != nil {
		d.AddBinding(state.BindToScheduler(d.open, d, sched))
	}
}

// Unmount implements widget.Lifecycle.
func (d *DateRangePickerWidget) Unmount() {}

// applyRange records a calendar selection. A complete range updates the
// trigger, fires OnChange, and closes the popover; a partial selection (just
// the start) keeps the popover open for the second click.
func (d *DateRangePickerWidget) applyRange(start, end time.Time) {
	s := truncateDay(start)
	e := truncateDay(end)
	d.start = &s
	d.end = &e
	d.rebuildTrigger()
	d.open.Set(false)
	d.SetNeedsRedraw(true)
	if d.onChange != nil {
		d.onChange(s, e)
	}
}

// syncOverlay pushes or removes the calendar overlay to match the open signal.
func (d *DateRangePickerWidget) syncOverlay(ctx widget.Context) {
	if ctx == nil || d.content == nil {
		return
	}
	om := ctx.OverlayManager()
	if om == nil {
		return
	}
	want := d.open.Get()
	if want == d.shown {
		return
	}
	if want {
		if d.start != nil {
			d.content.cal.Month(*d.start)
			if d.end != nil {
				d.content.cal.SelectedRange(*d.start, *d.end)
			}
		}
		size := d.content.Layout(ctx, geometry.Loose(ctx.WindowSize()))
		pos := corepopover.CalculatePosition(
			corepopover.Bottom, d.anchorBounds(), size, ctx.WindowSize(),
			metrics.Popover.SideOffset)
		d.content.SetBounds(geometry.FromPointSize(pos, size))
		d.om = om
		om.PushOverlay(d.content, d.handleDismiss)
		d.shown = true
	} else {
		om.RemoveOverlay(d.content)
		d.shown = false
	}
	d.SetNeedsRedraw(true)
}

func (d *DateRangePickerWidget) handleDismiss() {
	if !d.shown {
		return
	}
	d.shown = false
	if d.om != nil {
		d.om.RemoveOverlay(d.content)
	}
	if d.open.Get() {
		d.open.Set(false)
	}
	d.SetNeedsRedraw(true)
}

func (d *DateRangePickerWidget) anchorBounds() geometry.Rect {
	if r := d.ScreenBounds(); !r.IsEmpty() {
		return r
	}
	return d.Bounds()
}

// IsOpen reports whether the popover is open.
func (d *DateRangePickerWidget) IsOpen() bool { return d.open.Get() }

// dateRangePickerContent is the padding-free popover surface hosting the
// range-mode calendar (identical chrome to datePickerContent).
type dateRangePickerContent struct {
	widget.WidgetBase

	owner *DateRangePickerWidget
	cal   *CalendarWidget
}

func newDateRangePickerContent(owner *DateRangePickerWidget) *dateRangePickerContent {
	c := &dateRangePickerContent{owner: owner}
	c.SetVisible(true)
	c.SetEnabled(true)
	c.cal = Calendar().
		Range().
		Theme(owner.resolvedTheme()).
		OnRangeChange(owner.applyRange)
	ovlSetParent(c.cal, c)
	return c
}

// Layout sizes the surface to the calendar plus the content inset.
func (c *dateRangePickerContent) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	inset := metrics.DatePicker.ContentInset
	calSize := c.cal.Layout(ctx, geometry.Loose(geometry.Sz(geometry.Infinity, geometry.Infinity)))
	setWidgetBounds(c.cal, geometry.NewRect(inset, inset, calSize.Width, calSize.Height))
	size := cons.Constrain(geometry.Sz(calSize.Width+2*inset, calSize.Height+2*inset))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw paints the popover surface and the calendar.
func (c *dateRangePickerContent) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.owner.resolvedTheme()
	tok := th.Active()
	r := th.RadiusLG()
	bounds := c.Bounds()

	draw.Shadow(canvas, bounds, r, metrics.ShadowMD)
	canvas.DrawRoundRect(bounds, tok.Popover, r)
	draw.InsideBorder(canvas, bounds, r, tok.Border, metrics.Popover.BorderWidth)

	canvas.PushTransform(bounds.Min)
	widget.StampScreenOrigin(c.cal, canvas)
	widget.DrawChild(c.cal, ctx, canvas)
	canvas.PopTransform()
}

// Event dispatches to the calendar.
func (c *dateRangePickerContent) Event(ctx widget.Context, e event.Event) bool {
	return c.cal.Event(ctx, ovlTranslate(e, c.Bounds().Min))
}

// Children returns the calendar.
func (c *dateRangePickerContent) Children() []widget.Widget {
	return []widget.Widget{c.cal}
}

// DateRangePickerContentPreview renders the open range popover surface (border
// + range calendar) as direct content for goldens and docs.
func DateRangePickerContentPreview(month, start, end time.Time, th *theme.Theme) Widget {
	if th == nil {
		th = CurrentTheme()
	}
	owner := DateRangePicker().Theme(th).Value(start, end)
	c := newDateRangePickerContent(owner)
	c.cal.Month(month).SelectedRange(start, end).
		Today(time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location()))
	return c
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*DateRangePickerWidget)(nil)
	_ widget.Lifecycle = (*DateRangePickerWidget)(nil)
	_ widget.Widget    = (*dateRangePickerContent)(nil)
)
