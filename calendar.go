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

// CalendarWidget is the shadcn Calendar: a single-month day grid with a
// month/year caption, prev/next ghost nav buttons, a weekday header row, and
// a 6×7 grid of day cells. The selected day renders as a --primary pill, the
// current day gets an outline, and days outside the displayed month are
// muted (docs/research/03-shadcn-pixel-spec.md §5 "Calendar").
//
// Architecture decision: graft-owned widget. The grid math is pure Go
// time-package arithmetic and the visuals are token-driven ghost-button
// cells, so there is no substantial core machinery to wrap.
type CalendarWidget struct {
	widget.WidgetBase

	month    time.Time // any time within the displayed month
	selected *time.Time
	today    *time.Time // pinned "today" (nil = time.Now at draw)

	onSelect func(time.Time)
	weekday  time.Weekday // first day of week (default Sunday)

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

// Month sets the displayed month (any time within it).
func (c *CalendarWidget) Month(t time.Time) *CalendarWidget {
	c.month = firstOfMonth(t)
	return c
}

// Selected sets the selected day.
func (c *CalendarWidget) Selected(t time.Time) *CalendarWidget {
	tt := t
	c.selected = &tt
	return c
}

// Today pins the "today" highlight to a fixed date (deterministic goldens).
func (c *CalendarWidget) Today(t time.Time) *CalendarWidget {
	tt := t
	c.today = &tt
	return c
}

// OnSelect registers the day-selection observer.
func (c *CalendarWidget) OnSelect(fn func(time.Time)) *CalendarWidget {
	c.onSelect = fn
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
	var selected time.Time
	hasSelected := c.selected != nil
	if hasSelected {
		selected = truncateDay(*c.selected)
	}

	for row := 0; row < m.Rows; row++ {
		rowY := firstWeekY + float32(row)*(m.CellSize+m.WeekGap)
		for col := 0; col < m.Columns; col++ {
			day := start.AddDate(0, 0, row*m.Columns+col)
			idx := row*m.Columns + col + 1
			cellRect := geometry.NewRect(
				bounds.Min.X+float32(col)*m.CellSize, bounds.Min.Y+rowY,
				m.CellSize, m.CellSize)

			outside := day.Month() != c.month.Month()
			isSelected := hasSelected && sameDay(day, selected)
			isToday := sameDay(day, today)
			isHovered := idx == c.hovered

			c.drawDayCell(canvas, th, tok, cellRect, day, outside, isSelected, isToday, isHovered)
		}
	}
}

// drawDayCell paints one day: a primary pill when selected, an accent fill
// on hover, an outline on today, muted text outside the month.
func (c *CalendarWidget) drawDayCell(canvas widget.Canvas, th *theme.Theme, tok *theme.Tokens,
	cell geometry.Rect, day time.Time, outside, selected, today, hovered bool) {
	m := metrics.Calendar
	r := th.RadiusMD()

	fg := tok.Foreground

	if selected {
		canvas.DrawRoundRect(cell, tok.Primary, r)
		fg = tok.PrimaryForeground
	} else if hovered {
		canvas.DrawRoundRect(cell, tok.Accent, r)
		fg = tok.AccentForeground
		if outside {
			fg = tok.MutedForeground
		}
	} else if today {
		canvas.DrawRoundRect(cell, tok.Accent, r)
		fg = tok.AccentForeground
	}

	if outside && !selected {
		fg = tok.MutedForeground
	}

	label := strconv.Itoa(day.Day())
	drawCalendarText(canvas, label, cell, m.DayFontSize, m.DayFontWeight,
		fg, calendarFamily(th, m.DayFontWeight), widget.TextAlignCenter)
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
			c.selectDay(day)
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

func (c *CalendarWidget) selectDay(day time.Time) {
	d := truncateDay(day)
	c.selected = &d
	// Selecting an outside-month day navigates to that month.
	if day.Month() != c.month.Month() {
		c.month = firstOfMonth(day)
	}
	c.SetNeedsRedraw(true)
	if c.onSelect != nil {
		c.onSelect(d)
	}
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
