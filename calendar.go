package graft

import (
	"strconv"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// CalendarMode is the selection behavior of a Calendar (shadcn:
// mode="single" | "range" | "multiple").
type CalendarMode int

const (
	// CalendarSingle selects exactly one day (the default).
	CalendarSingle CalendarMode = iota
	// CalendarRange selects a contiguous start..end span.
	CalendarRange
	// CalendarMultiple selects an arbitrary set of days via toggle.
	CalendarMultiple
)

// CalendarWidget is the shadcn Calendar: a single-month day grid with a
// month/year caption, prev/next ghost nav buttons, a weekday header row, and
// a 6×7 grid of day cells. The selected day renders as a --primary pill, the
// current day gets an outline, and days outside the displayed month are
// muted (docs/research/03-shadcn-pixel-spec.md §5 "Calendar").
//
// Three selection modes mirror react-day-picker: single (one --primary pill),
// range (start/end --primary pills with a --accent band on the in-between
// days), and multiple (a set of toggled --primary pills).
//
// Architecture decision: graft-owned widget. The grid math is pure Go
// time-package arithmetic and the visuals are token-driven ghost-button
// cells, so there is no substantial core machinery to wrap.
type CalendarWidget struct {
	widget.WidgetBase

	mode  CalendarMode
	month time.Time  // any time within the displayed month
	today *time.Time // pinned "today" (nil = time.Now at draw)

	// single
	selected *time.Time

	// range
	rangeStart *time.Time
	rangeEnd   *time.Time

	// multiple
	dates []time.Time

	onSelect      func(time.Time)
	onRangeChange func(start, end time.Time)
	onDatesChange func([]time.Time)
	weekday       time.Weekday // first day of week (default Sunday)

	hovered int // 1-based day index hovered, or 0

	prevBtn *ButtonWidget
	nextBtn *ButtonWidget

	theme *theme.Theme
}

// Calendar builds a calendar showing the current month.
func Calendar() *CalendarWidget {
	c := &CalendarWidget{
		month:   firstOfMonth(time.Now()),
		weekday: time.Sunday,
		theme:   CurrentTheme(),
	}
	c.SetVisible(true)
	c.SetEnabled(true)
	c.prevBtn = Button("").Ghost().IconOnly(icons.ChevronLeft)
	c.nextBtn = Button("").Ghost().IconOnly(icons.ChevronRight)
	c.prevBtn.OnClick(func() { c.shiftMonth(-1) })
	c.nextBtn.OnClick(func() { c.shiftMonth(1) })
	ovlSetParent(c.prevBtn, c)
	ovlSetParent(c.nextBtn, c)
	return c
}

// Mode sets the selection mode (single, range, or multiple).
func (c *CalendarWidget) Mode(m CalendarMode) *CalendarWidget {
	c.mode = m
	return c
}

// Range switches the calendar to range-selection mode.
func (c *CalendarWidget) Range() *CalendarWidget {
	c.mode = CalendarRange
	return c
}

// Multiple switches the calendar to multiple-selection mode.
func (c *CalendarWidget) Multiple() *CalendarWidget {
	c.mode = CalendarMultiple
	return c
}

// Month sets the displayed month (any time within it).
func (c *CalendarWidget) Month(t time.Time) *CalendarWidget {
	c.month = firstOfMonth(t)
	return c
}

// Selected sets the selected day (single mode).
func (c *CalendarWidget) Selected(t time.Time) *CalendarWidget {
	tt := t
	c.selected = &tt
	return c
}

// SelectedRange presets the start/end of a range (range mode). The pair is
// normalized so start <= end.
func (c *CalendarWidget) SelectedRange(start, end time.Time) *CalendarWidget {
	c.mode = CalendarRange
	s := truncateDay(start)
	e := truncateDay(end)
	if e.Before(s) {
		s, e = e, s
	}
	c.rangeStart = &s
	c.rangeEnd = &e
	return c
}

// SelectedDates presets the selected set (multiple mode). Dates are truncated
// to day granularity and de-duplicated.
func (c *CalendarWidget) SelectedDates(dates ...time.Time) *CalendarWidget {
	c.mode = CalendarMultiple
	c.dates = c.dates[:0]
	for _, d := range dates {
		c.addDate(truncateDay(d))
	}
	return c
}

// Today pins the "today" highlight to a fixed date (deterministic goldens).
func (c *CalendarWidget) Today(t time.Time) *CalendarWidget {
	tt := t
	c.today = &tt
	return c
}

// OnSelect registers the day-selection observer (single mode).
func (c *CalendarWidget) OnSelect(fn func(time.Time)) *CalendarWidget {
	c.onSelect = fn
	return c
}

// OnRangeChange registers the range-selection observer (range mode). It fires
// when a complete start..end range is formed.
func (c *CalendarWidget) OnRangeChange(fn func(start, end time.Time)) *CalendarWidget {
	c.onRangeChange = fn
	return c
}

// OnDatesChange registers the set-selection observer (multiple mode). It fires
// with the full selected set on every toggle.
func (c *CalendarWidget) OnDatesChange(fn func([]time.Time)) *CalendarWidget {
	c.onDatesChange = fn
	return c
}

// WeekStartsOn sets the first day of the week (default Sunday).
func (c *CalendarWidget) WeekStartsOn(d time.Weekday) *CalendarWidget {
	c.weekday = d
	return c
}

// Theme pins a specific theme.
func (c *CalendarWidget) Theme(th *theme.Theme) *CalendarWidget {
	if th != nil {
		c.theme = th
		c.prevBtn.Theme(th)
		c.nextBtn.Theme(th)
	}
	return c
}

func (c *CalendarWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

func (c *CalendarWidget) resolvedToday() time.Time {
	if c.today != nil {
		return *c.today
	}
	return time.Now()
}

func (c *CalendarWidget) shiftMonth(delta int) {
	c.month = c.month.AddDate(0, delta, 0)
	c.SetNeedsRedraw(true)
}

// firstOfMonth returns midnight on the first day of t's month (local).
func firstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// gridStart returns the date in the top-left cell of the grid: the first
// day of the displayed month rolled back to the configured week start.
func (c *CalendarWidget) gridStart() time.Time {
	first := firstOfMonth(c.month)
	offset := (int(first.Weekday()) - int(c.weekday) + 7) % 7
	return first.AddDate(0, 0, -offset)
}

// gridSize returns the calendar's intrinsic size.
func (c *CalendarWidget) calendarSize() geometry.Size {
	m := metrics.Calendar
	w := m.CellSize * float32(m.Columns)
	h := m.CellSize // header row (caption + nav)
	h += m.CaptionGap
	h += m.CellSize // weekday header row
	h += m.HeaderGap
	for i := 0; i < m.Rows; i++ {
		h += m.CellSize
		if i > 0 {
			h += m.WeekGap
		}
	}
	return geometry.Sz(w, h)
}

// Layout sizes the calendar to its intrinsic size and positions nav buttons.
func (c *CalendarWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	m := metrics.Calendar
	size := cons.Constrain(c.calendarSize())
	c.SetBounds(geometry.FromPointSize(c.Position(), size))

	// Nav buttons are 32px ghost icon buttons at the header row corners.
	c.prevBtn.Layout(ctx, geometry.Loose(geometry.Sz(m.CellSize, m.CellSize)))
	c.nextBtn.Layout(ctx, geometry.Loose(geometry.Sz(m.CellSize, m.CellSize)))
	setWidgetBounds(c.prevBtn, geometry.NewRect(0, 0, m.CellSize, m.CellSize))
	setWidgetBounds(c.nextBtn, geometry.NewRect(size.Width-m.CellSize, 0, m.CellSize, m.CellSize))
	return size
}

// rowYs returns the y offsets (widget-local) of: caption row, weekday row,
// and the first week row.
func (c *CalendarWidget) rowYs() (caption, weekday, firstWeek float32) {
	m := metrics.Calendar
	caption = 0
	weekday = m.CellSize + m.CaptionGap
	firstWeek = weekday + m.CellSize + m.HeaderGap
	return
}

// Draw paints the caption, nav buttons, weekday header, and day grid.
func (c *CalendarWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.resolvedTheme()
	tok := th.Active()
	m := metrics.Calendar
	bounds := c.Bounds()
	captionY, weekdayY, firstWeekY := c.rowYs()

	// Caption: month + year, centered in the header row.
	caption := c.month.Format("January 2006")
	captionRect := geometry.NewRect(
		bounds.Min.X, bounds.Min.Y+captionY, bounds.Width(), m.CellSize)
	drawCalendarText(canvas, caption, captionRect, m.CaptionFontSize, m.CaptionFontWeight,
		tok.Foreground, calendarFamily(th, m.CaptionFontWeight), widget.TextAlignCenter)

	// Nav buttons.
	canvas.PushTransform(bounds.Min)
	widget.StampScreenOrigin(c.prevBtn, canvas)
	widget.DrawChild(c.prevBtn, ctx, canvas)
	widget.StampScreenOrigin(c.nextBtn, canvas)
	widget.DrawChild(c.nextBtn, ctx, canvas)
	canvas.PopTransform()

	// Weekday header row.
	start := c.gridStart()
	for col := 0; col < m.Columns; col++ {
		label := weekdayLabel(start.AddDate(0, 0, col).Weekday())
		cellRect := geometry.NewRect(
			bounds.Min.X+float32(col)*m.CellSize, bounds.Min.Y+weekdayY,
			m.CellSize, m.CellSize)
		drawCalendarText(canvas, label, cellRect, m.WeekdayFontSize, m.WeekdayFontWeight,
			tok.MutedForeground, calendarFamily(th, m.WeekdayFontWeight), widget.TextAlignCenter)
	}

	// Day grid.
	today := truncateDay(c.resolvedToday())

	for row := 0; row < m.Rows; row++ {
		rowY := firstWeekY + float32(row)*(m.CellSize+m.WeekGap)
		for col := 0; col < m.Columns; col++ {
			day := start.AddDate(0, 0, row*m.Columns+col)
			idx := row*m.Columns + col + 1
			cellRect := geometry.NewRect(
				bounds.Min.X+float32(col)*m.CellSize, bounds.Min.Y+rowY,
				m.CellSize, m.CellSize)

			outside := day.Month() != c.month.Month()
			isToday := sameDay(day, today)
			isHovered := idx == c.hovered

			c.drawDayCell(canvas, th, tok, cellRect, day, col, outside, isToday, isHovered)
		}
	}
}

// dayState describes how a single day participates in the current selection.
type dayState struct {
	selected   bool // a filled --primary pill (single, multiple, or range end)
	rangeStart bool
	rangeEnd   bool
	rangeMid   bool // in-between day: --accent band, no pill
}

// stateFor classifies day against the active selection mode.
func (c *CalendarWidget) stateFor(day time.Time) dayState {
	switch c.mode {
	case CalendarRange:
		s, e := c.rangeStart, c.rangeEnd
		switch {
		case s != nil && e != nil:
			st := dayState{}
			if sameDay(day, *s) {
				st.selected, st.rangeStart = true, true
			}
			if sameDay(day, *e) {
				st.selected, st.rangeEnd = true, true
			}
			if !st.selected && day.After(*s) && day.Before(*e) {
				st.rangeMid = true
			}
			return st
		case s != nil:
			if sameDay(day, *s) {
				return dayState{selected: true, rangeStart: true, rangeEnd: true}
			}
		}
		return dayState{}
	case CalendarMultiple:
		if c.hasDate(day) {
			return dayState{selected: true}
		}
		return dayState{}
	default: // CalendarSingle
		if c.selected != nil && sameDay(day, truncateDay(*c.selected)) {
			return dayState{selected: true}
		}
		return dayState{}
	}
}

// drawDayCell paints one day: a primary pill when selected, an accent band for
// the in-between days of a range, an accent fill on hover, an accent fill on
// today, and muted text outside the month.
func (c *CalendarWidget) drawDayCell(canvas widget.Canvas, th *theme.Theme, tok *theme.Tokens,
	cell geometry.Rect, day time.Time, col int, outside, today, hovered bool) {
	m := metrics.Calendar
	r := th.RadiusMD()

	st := c.stateFor(day)
	fg := tok.Foreground

	switch {
	case st.rangeMid:
		// Continuous square accent band behind the in-between days.
		canvas.DrawRect(cell, tok.Accent)
		fg = tok.AccentForeground
	case st.selected:
		// Range ends bleed a half accent band toward the middle so the band
		// is visually continuous under the rounded pill.
		if st.rangeStart && !st.rangeEnd && col < m.Columns-1 {
			canvas.DrawRect(rightHalf(cell), tok.Accent)
		}
		if st.rangeEnd && !st.rangeStart && col > 0 {
			canvas.DrawRect(leftHalf(cell), tok.Accent)
		}
		canvas.DrawRoundRect(cell, tok.Primary, r)
		fg = tok.PrimaryForeground
	case hovered:
		canvas.DrawRoundRect(cell, tok.Accent, r)
		fg = tok.AccentForeground
		if outside {
			fg = tok.MutedForeground
		}
	case today:
		canvas.DrawRoundRect(cell, tok.Accent, r)
		fg = tok.AccentForeground
	}

	if outside && !st.selected && !st.rangeMid {
		fg = tok.MutedForeground
	}

	label := strconv.Itoa(day.Day())
	drawCalendarText(canvas, label, cell, m.DayFontSize, m.DayFontWeight,
		fg, calendarFamily(th, m.DayFontWeight), widget.TextAlignCenter)
}

// leftHalf / rightHalf return the inner half of a cell, used to bridge the
// accent band from a range end into the adjacent in-between day.
func leftHalf(cell geometry.Rect) geometry.Rect {
	return geometry.NewRect(cell.Min.X, cell.Min.Y, cell.Width()/2, cell.Height())
}

func rightHalf(cell geometry.Rect) geometry.Rect {
	return geometry.NewRect(cell.Min.X+cell.Width()/2, cell.Min.Y, cell.Width()/2, cell.Height())
}

// Event handles nav clicks and day selection/hover.
func (c *CalendarWidget) Event(ctx widget.Context, e event.Event) bool {
	offset := c.Bounds().Min
	local := ovlTranslate(e, offset)

	// Route mouse events to a nav button only when the pointer is inside its
	// (local) bounds, because ButtonWidget consumes any left press regardless
	// of position.
	if me, ok := local.(*event.MouseEvent); ok {
		if c.prevBtn.Bounds().Contains(me.Position) {
			return c.prevBtn.Event(ctx, local)
		}
		if c.nextBtn.Bounds().Contains(me.Position) {
			return c.nextBtn.Event(ctx, local)
		}
	} else {
		// Non-mouse events (keys) still reach both buttons.
		if c.prevBtn.Event(ctx, local) || c.nextBtn.Event(ctx, local) {
			return true
		}
	}

	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	idx, day, inGrid := c.hitDay(me.Position)
	switch me.MouseType {
	case event.MouseMove, event.MouseEnter:
		newHover := 0
		if inGrid {
			newHover = idx
		}
		if newHover != c.hovered {
			c.hovered = newHover
			c.SetNeedsRedraw(true)
		}
		if inGrid {
			ctx.SetCursor(widget.CursorPointer)
		}
		return inGrid
	case event.MouseLeave:
		if c.hovered != 0 {
			c.hovered = 0
			c.SetNeedsRedraw(true)
		}
		return false
	case event.MouseRelease:
		if me.Button == event.ButtonLeft && inGrid {
			c.selectDay(day, ctx)
			return true
		}
	}
	return false
}

// hitDay returns the 1-based cell index, its date, and whether the point is
// inside the day grid.
func (c *CalendarWidget) hitDay(pos geometry.Point) (int, time.Time, bool) {
	m := metrics.Calendar
	_, _, firstWeekY := c.rowYs()
	local := pos.Sub(c.Bounds().Min)
	if local.X < 0 || local.X >= m.CellSize*float32(m.Columns) {
		return 0, time.Time{}, false
	}
	relY := local.Y - firstWeekY
	if relY < 0 {
		return 0, time.Time{}, false
	}
	rowStride := m.CellSize + m.WeekGap
	row := int(relY / rowStride)
	if row >= m.Rows {
		return 0, time.Time{}, false
	}
	if within := relY - float32(row)*rowStride; within > m.CellSize {
		return 0, time.Time{}, false // in the inter-row gap
	}
	col := int(local.X / m.CellSize)
	idx := row*m.Columns + col + 1
	day := c.gridStart().AddDate(0, 0, row*m.Columns+col)
	return idx, day, true
}

// selectDay applies a click to the active selection mode and repaints.
func (c *CalendarWidget) selectDay(day time.Time, ctx widget.Context) {
	d := truncateDay(day)
	// Selecting an outside-month day navigates to that month.
	if day.Month() != c.month.Month() {
		c.month = firstOfMonth(day)
	}
	switch c.mode {
	case CalendarRange:
		c.applyRangeClick(d)
	case CalendarMultiple:
		c.toggleDate(d)
	default:
		c.applySingleClick(d)
	}
	c.SetNeedsRedraw(true)
	if ctx != nil {
		ctx.Invalidate()
	}
}

func (c *CalendarWidget) applySingleClick(d time.Time) {
	c.selected = &d
	if c.onSelect != nil {
		c.onSelect(d)
	}
}

// applyRangeClick runs the three-click range state machine: first click sets
// the start (clears the end); second click sets the end (swapping if it falls
// before the start); a click while a complete range exists restarts at the new
// start.
func (c *CalendarWidget) applyRangeClick(d time.Time) {
	switch {
	case c.rangeStart == nil || c.rangeEnd != nil:
		// Start a fresh range.
		c.rangeStart = &d
		c.rangeEnd = nil
	default:
		// Close the range, swapping if the second click precedes the start.
		start := *c.rangeStart
		end := d
		if end.Before(start) {
			start, end = end, start
		}
		c.rangeStart = &start
		c.rangeEnd = &end
		if c.onRangeChange != nil {
			c.onRangeChange(start, end)
		}
	}
}

func (c *CalendarWidget) toggleDate(d time.Time) {
	if c.hasDate(d) {
		c.removeDate(d)
	} else {
		c.addDate(d)
	}
	if c.onDatesChange != nil {
		out := make([]time.Time, len(c.dates))
		copy(out, c.dates)
		c.onDatesChange(out)
	}
}

func (c *CalendarWidget) hasDate(d time.Time) bool {
	dd := truncateDay(d)
	for _, x := range c.dates {
		if sameDay(x, dd) {
			return true
		}
	}
	return false
}

func (c *CalendarWidget) addDate(d time.Time) {
	if !c.hasDate(d) {
		c.dates = append(c.dates, truncateDay(d))
	}
}

func (c *CalendarWidget) removeDate(d time.Time) {
	dd := truncateDay(d)
	out := c.dates[:0]
	for _, x := range c.dates {
		if !sameDay(x, dd) {
			out = append(out, x)
		}
	}
	c.dates = out
}

// SelectedDay returns the single-mode selection, if any.
func (c *CalendarWidget) SelectedDay() (time.Time, bool) {
	if c.selected == nil {
		return time.Time{}, false
	}
	return *c.selected, true
}

// SelectedRangeValue returns the current range and whether it is complete.
func (c *CalendarWidget) SelectedRangeValue() (start, end time.Time, complete bool) {
	if c.rangeStart == nil {
		return time.Time{}, time.Time{}, false
	}
	if c.rangeEnd == nil {
		return *c.rangeStart, time.Time{}, false
	}
	return *c.rangeStart, *c.rangeEnd, true
}

// SelectedDatesValue returns a copy of the multiple-mode selection set.
func (c *CalendarWidget) SelectedDatesValue() []time.Time {
	out := make([]time.Time, len(c.dates))
	copy(out, c.dates)
	return out
}

// Children returns the nav buttons.
func (c *CalendarWidget) Children() []widget.Widget {
	return []widget.Widget{c.prevBtn, c.nextBtn}
}

// --- helpers ------------------------------------------------------------

// truncateDay zeroes the time-of-day, keeping the calendar date.
func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// sameDay reports whether a and b fall on the same calendar date.
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// weekdayLabel returns the two-letter weekday header (Su Mo Tu We Th Fr Sa).
func weekdayLabel(d time.Weekday) string {
	return [...]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}[d]
}

// calendarFamily resolves the Geist family for a weight, honoring a custom
// theme sans font.
func calendarFamily(th *theme.Theme, weight int) string {
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	return fonts.Family(weight)
}

// drawCalendarText paints a single line of calendar text.
func drawCalendarText(canvas widget.Canvas, s string, bounds geometry.Rect,
	size float32, weight int, col widget.Color, family string, align widget.TextAlign) {
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(s, bounds, widget.TextStyle{
			FontFamily: family,
			FontSize:   size,
			Color:      col,
			Align:      align,
		})
		return
	}
	canvas.DrawText(s, bounds, size, col, weight >= 600, align)
}

// Compile-time interface checks.
var _ widget.Widget = (*CalendarWidget)(nil)
