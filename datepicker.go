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

// DatePickerWidget is the shadcn DatePicker: an outline Button trigger
// (calendar icon + formatted date or "Pick a date" placeholder) that opens a
// Popover hosting a Calendar. Selecting a day sets the value and closes the
// popover (docs/research/03-shadcn-pixel-spec.md §5 Popover/Calendar; the
// DatePicker is the shadcn composition recipe).
//
// Architecture decision: graft-owned composite. It reuses ButtonWidget for
// the trigger and CalendarWidget for the surface content, and manages its own
// overlay (push on open, remove on close, dismiss on outside click/Escape)
// the same way PopoverWidget does, because the popover content here is a
// padding-free (p-0 w-auto) calendar surface rather than the stock
// PopoverContent.
type DatePickerWidget struct {
	widget.WidgetBase

	value       *time.Time
	valueSig    state.Signal[string] // optional controlled value (formatted)
	placeholder string
	format      string
	onChange    func(time.Time)

	open  state.Signal[bool]
	shown bool
	om    widget.OverlayManager

	trigger *ButtonWidget
	content *datePickerContent

	theme *theme.Theme
}

// DatePicker builds a date picker with the default "Pick a date" placeholder.
func DatePicker() *DatePickerWidget {
	d := &DatePickerWidget{
		placeholder: metrics.DatePicker.Placeholder,
		format:      metrics.DatePicker.DefaultFormat,
		open:        state.NewSignal(false),
		theme:       CurrentTheme(),
	}
	d.SetVisible(true)
	d.SetEnabled(true)
	d.rebuildTrigger()
	d.content = newDatePickerContent(d)
	return d
}

// Value sets the selected date.
func (d *DatePickerWidget) Value(t time.Time) *DatePickerWidget {
	tt := t
	d.value = &tt
	d.rebuildTrigger()
	return d
}

// Placeholder sets the empty-state label.
func (d *DatePickerWidget) Placeholder(s string) *DatePickerWidget {
	d.placeholder = s
	d.rebuildTrigger()
	return d
}

// Format sets the Go time layout used to render the selected date.
func (d *DatePickerWidget) Format(layout string) *DatePickerWidget {
	d.format = layout
	d.rebuildTrigger()
	return d
}

// OnChange registers the date-selection observer.
func (d *DatePickerWidget) OnChange(fn func(time.Time)) *DatePickerWidget {
	d.onChange = fn
	return d
}

// Theme pins a specific theme.
func (d *DatePickerWidget) Theme(th *theme.Theme) *DatePickerWidget {
	if th != nil {
		d.theme = th
		d.rebuildTrigger()
		if d.content != nil {
			d.content.cal.Theme(th)
		}
	}
	return d
}

func (d *DatePickerWidget) resolvedTheme() *theme.Theme {
	if d.theme != nil {
		return d.theme
	}
	return CurrentTheme()
}

// triggerLabel returns the formatted date or the placeholder.
func (d *DatePickerWidget) triggerLabel() string {
	if d.value != nil {
		return d.value.Format(d.format)
	}
	return d.placeholder
}

// rebuildTrigger rebuilds the outline button trigger (label changes with the
// value). The placeholder state uses muted-foreground text.
func (d *DatePickerWidget) rebuildTrigger() {
	btn := Button(d.triggerLabel()).
		Outline().
		Icon(icons.Calendar).
		W(metrics.DatePicker.TriggerWidth).
		Theme(d.resolvedTheme()).
		OnClick(func() { d.open.Set(!d.open.Get()) })
	// Empty state uses muted-foreground text (data-[empty=true]).
	if d.value == nil {
		tok := d.resolvedTheme().Active()
		mc := tok.MutedForeground
		btn.Style(func(s *Style) { s.Foreground = &mc })
	}
	d.trigger = btn
	ovlSetParent(d.trigger, d)
}

// Layout sizes the host to the trigger.
func (d *DatePickerWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	size := d.trigger.Layout(ctx, c)
	setWidgetBounds(d.trigger, geometry.NewRect(0, 0, size.Width, size.Height))
	d.SetBounds(geometry.FromPointSize(d.Position(), size))
	return size
}

// Draw renders the trigger and reconciles the overlay with the open signal.
func (d *DatePickerWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
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

// Event forwards to the trigger button, whose OnClick toggles the popover.
func (d *DatePickerWidget) Event(ctx widget.Context, e event.Event) bool {
	offset := d.Bounds().Min
	consumed := d.trigger.Event(ctx, ovlTranslate(e, offset))
	d.syncOverlay(ctx)
	return consumed
}

// Children returns the inline trigger.
func (d *DatePickerWidget) Children() []widget.Widget {
	return []widget.Widget{d.trigger}
}

// Mount binds the open signal.
func (d *DatePickerWidget) Mount(ctx widget.Context) {
	if sched := ctx.Scheduler(); sched != nil {
		d.AddBinding(state.BindToScheduler(d.open, d, sched))
	}
}

// Unmount implements widget.Lifecycle.
func (d *DatePickerWidget) Unmount() {}

// selectDate sets the value, closes the popover, and fires OnChange.
func (d *DatePickerWidget) selectDate(t time.Time) {
	tt := t
	d.value = &tt
	d.rebuildTrigger()
	d.open.Set(false)
	d.SetNeedsRedraw(true)
	if d.onChange != nil {
		d.onChange(t)
	}
}

// syncOverlay pushes or removes the calendar overlay to match the open
// signal (mirrors PopoverWidget.syncOverlay).
func (d *DatePickerWidget) syncOverlay(ctx widget.Context) {
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
		// Show the selected month, or the current month when empty.
		if d.value != nil {
			d.content.cal.Month(*d.value).Selected(*d.value)
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

func (d *DatePickerWidget) handleDismiss() {
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

func (d *DatePickerWidget) anchorBounds() geometry.Rect {
	if r := d.ScreenBounds(); !r.IsEmpty() {
		return r
	}
	return d.Bounds()
}

// IsOpen reports whether the popover is open.
func (d *DatePickerWidget) IsOpen() bool { return d.open.Get() }

// datePickerContent is the padding-free popover surface (p-0, w-auto)
// hosting the calendar: bg-popover, 1px border, rounded-md, shadow-md, with
// a small inset around the calendar.
type datePickerContent struct {
	widget.WidgetBase

	owner *DatePickerWidget
	cal   *CalendarWidget
}

func newDatePickerContent(owner *DatePickerWidget) *datePickerContent {
	c := &datePickerContent{owner: owner}
	c.SetVisible(true)
	c.SetEnabled(true)
	c.cal = Calendar().
		Theme(owner.resolvedTheme()).
		OnSelect(owner.selectDate)
	ovlSetParent(c.cal, c)
	return c
}

// Layout sizes the surface to the calendar plus the content inset.
func (c *datePickerContent) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	inset := metrics.DatePicker.ContentInset
	calSize := c.cal.Layout(ctx, geometry.Loose(geometry.Sz(geometry.Infinity, geometry.Infinity)))
	setWidgetBounds(c.cal, geometry.NewRect(inset, inset, calSize.Width, calSize.Height))
	size := cons.Constrain(geometry.Sz(calSize.Width+2*inset, calSize.Height+2*inset))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw paints the popover surface and the calendar.
func (c *datePickerContent) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.owner.resolvedTheme()
	tok := th.Active()
	r := th.RadiusMD()
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
func (c *datePickerContent) Event(ctx widget.Context, e event.Event) bool {
	return c.cal.Event(ctx, ovlTranslate(e, c.Bounds().Min))
}

// Children returns the calendar.
func (c *datePickerContent) Children() []widget.Widget {
	return []widget.Widget{c.cal}
}

// DatePickerContentPreview renders the open popover surface (border + calendar)
// as direct content for goldens and docs, with the given month/selection.
func DatePickerContentPreview(month, selected time.Time, th *theme.Theme) Widget {
	if th == nil {
		th = CurrentTheme()
	}
	owner := DatePicker().Theme(th).Value(selected)
	c := newDatePickerContent(owner)
	c.cal.Month(month).Selected(selected).
		Today(time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location()))
	return c
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*DatePickerWidget)(nil)
	_ widget.Lifecycle = (*DatePickerWidget)(nil)
	_ widget.Widget    = (*datePickerContent)(nil)
)
