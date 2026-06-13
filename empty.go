package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// EmptyWidget is shadcn's Empty state: a centered column with a dashed rounded
// border holding an optional media tile, a title, a description, and content
// actions, all muted and centered.
//
// shadcn root: "flex min-w-0 flex-1 flex-col items-center justify-center gap-6
// rounded-lg border-dashed p-6 text-center". graft renders the base (non-md)
// padding p-6 = 24px, the dashed 1px border in the Border token, and a rounded-lg
// surface (transparent fill).
//
// Architecture: graft-owned container. It lays out its section children in a
// centered vertical stack and paints the dashed border itself so the stroke
// color resolves from the active token set at draw time.
type EmptyWidget struct {
	widget.WidgetBase

	children []widget.Widget
	theme    *theme.Theme
}

// Empty assembles an empty state from sections (EmptyHeader, EmptyContent) or
// any widgets.
func Empty(children ...widget.Widget) *EmptyWidget {
	e := &EmptyWidget{children: children, theme: CurrentTheme()}
	e.SetVisible(true)
	e.SetEnabled(true)
	for _, ch := range children {
		ovlSetParent(ch, e)
	}
	return e
}

// Theme pins a specific theme instead of the process-wide current theme.
func (e *EmptyWidget) Theme(th *theme.Theme) *EmptyWidget {
	e.theme = th
	return e
}

func (e *EmptyWidget) resolvedTheme() *theme.Theme {
	if e.theme != nil {
		return e.theme
	}
	return CurrentTheme()
}

// Layout stacks the sections vertically, centered, with gap-6 inside p-6,
// measuring each child at the inner content width.
func (e *EmptyWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	pad := metrics.EmptyPadding
	availW := cons.MaxWidth
	if availW <= 0 || availW >= geometry.Infinity {
		availW = metrics.EmptyMaxWidth + 2*pad
	}
	innerW := availW - 2*pad
	if innerW < 0 {
		innerW = 0
	}
	childCons := geometry.Loose(geometry.Sz(innerW, 100000))

	sizes := make([]geometry.Size, len(e.children))
	var stackH, maxChildW float32
	for i, ch := range e.children {
		if i > 0 {
			stackH += metrics.EmptyGap
		}
		sizes[i] = ch.Layout(ctx, childCons)
		stackH += sizes[i].Height
		if sizes[i].Width > maxChildW {
			maxChildW = sizes[i].Width
		}
	}

	outerW := availW
	outerH := stackH + 2*pad

	// Center each child horizontally within the inner content box.
	x0 := e.Position().X + pad
	cursorY := e.Position().Y + pad
	for i, ch := range e.children {
		if i > 0 {
			cursorY += metrics.EmptyGap
		}
		cx := x0 + (innerW-sizes[i].Width)/2
		ovlSetBounds(ch, geometry.FromPointSize(geometry.Pt(cx, cursorY), sizes[i]))
		cursorY += sizes[i].Height
	}

	size := cons.Constrain(geometry.Sz(outerW, outerH))
	e.SetBounds(geometry.FromPointSize(e.Position(), size))
	return size
}

// Draw paints the dashed border and the section children.
func (e *EmptyWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !e.IsVisible() {
		return
	}
	th := e.resolvedTheme()
	tok := th.Active()
	b := e.Bounds()

	drawDashedBorder(canvas, b, tok.Border, metrics.CardBorderWidth)

	for _, ch := range e.children {
		ch.Draw(ctx, canvas)
	}
}

// Event forwards input to the children (action buttons inside EmptyContent).
func (e *EmptyWidget) Event(ctx widget.Context, ev event.Event) bool {
	for _, ch := range e.children {
		if ch.Event(ctx, ev) {
			return true
		}
	}
	return false
}

// Children returns the section widgets.
func (e *EmptyWidget) Children() []widget.Widget { return e.children }

// EmptyHeader stacks the media tile, title, and description in a centered
// column with gap-2, capped at max-w-sm.
func EmptyHeader(children ...widget.Widget) *EmptySectionWidget {
	return newEmptySection(children, metrics.EmptyHeaderGap)
}

// EmptyContent stacks action widgets in a centered column with gap-4, capped at
// max-w-sm.
func EmptyContent(children ...widget.Widget) *EmptySectionWidget {
	return newEmptySection(children, metrics.EmptyContentGap)
}

// EmptySectionWidget is a centered vertical stack (EmptyHeader / EmptyContent)
// capped at max-w-sm.
type EmptySectionWidget struct {
	widget.WidgetBase
	children []widget.Widget
	gap      float32
}

func newEmptySection(children []widget.Widget, gap float32) *EmptySectionWidget {
	s := &EmptySectionWidget{children: children, gap: gap}
	s.SetVisible(true)
	s.SetEnabled(true)
	for _, ch := range children {
		ovlSetParent(ch, s)
	}
	return s
}

// Layout stacks children vertically, centered, capping width at max-w-sm.
func (s *EmptySectionWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	maxW := metrics.EmptyMaxWidth
	if cons.MaxWidth > 0 && cons.MaxWidth < maxW {
		maxW = cons.MaxWidth
	}
	childCons := geometry.Loose(geometry.Sz(maxW, 100000))

	sizes := make([]geometry.Size, len(s.children))
	var stackH, contentW float32
	for i, ch := range s.children {
		if i > 0 {
			stackH += s.gap
		}
		sizes[i] = ch.Layout(ctx, childCons)
		stackH += sizes[i].Height
		if sizes[i].Width > contentW {
			contentW = sizes[i].Width
		}
	}

	x0 := s.Position().X
	cursorY := s.Position().Y
	for i, ch := range s.children {
		if i > 0 {
			cursorY += s.gap
		}
		cx := x0 + (contentW-sizes[i].Width)/2
		ovlSetBounds(ch, geometry.FromPointSize(geometry.Pt(cx, cursorY), sizes[i]))
		cursorY += sizes[i].Height
	}

	size := cons.Constrain(geometry.Sz(contentW, stackH))
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

func (s *EmptySectionWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	for _, ch := range s.children {
		ch.Draw(ctx, canvas)
	}
}

func (s *EmptySectionWidget) Event(ctx widget.Context, ev event.Event) bool {
	for _, ch := range s.children {
		if ch.Event(ctx, ev) {
			return true
		}
	}
	return false
}

func (s *EmptySectionWidget) Children() []widget.Widget { return s.children }

// EmptyMediaWidget is the icon tile: a size-10 rounded-lg muted square with a
// centered size-6 icon in the foreground token (the "icon" variant).
type EmptyMediaWidget struct {
	widget.WidgetBase
	ico   icon.IconData
	theme *theme.Theme
}

// EmptyMedia builds the icon media tile (the shadcn "icon" variant).
func EmptyMedia(ico icon.IconData) *EmptyMediaWidget {
	m := &EmptyMediaWidget{ico: ico, theme: CurrentTheme()}
	m.SetVisible(true)
	m.SetEnabled(true)
	return m
}

// Theme pins a specific theme instead of the process-wide current theme.
func (m *EmptyMediaWidget) Theme(th *theme.Theme) *EmptyMediaWidget {
	m.theme = th
	return m
}

func (m *EmptyMediaWidget) resolvedTheme() *theme.Theme {
	if m.theme != nil {
		return m.theme
	}
	return CurrentTheme()
}

// Layout reserves the size-10 tile plus mb-2 margin below it.
func (m *EmptyMediaWidget) Layout(_ widget.Context, cons geometry.Constraints) geometry.Size {
	size := cons.Constrain(geometry.Sz(
		metrics.EmptyMediaSize,
		metrics.EmptyMediaSize+metrics.EmptyMediaMarginBottom,
	))
	m.SetBounds(geometry.FromPointSize(m.Position(), size))
	return size
}

// Draw paints the muted rounded tile and the centered foreground icon.
func (m *EmptyMediaWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !m.IsVisible() {
		return
	}
	th := m.resolvedTheme()
	tok := th.Active()
	b := m.Bounds()

	tile := geometry.NewRect(b.Min.X, b.Min.Y, metrics.EmptyMediaSize, metrics.EmptyMediaSize)
	canvas.DrawRoundRect(tile, tok.Muted, th.RadiusLG())

	iconRect := geometry.NewRect(
		tile.Min.X+(metrics.EmptyMediaSize-metrics.EmptyMediaIconSize)/2,
		tile.Min.Y+(metrics.EmptyMediaSize-metrics.EmptyMediaIconSize)/2,
		metrics.EmptyMediaIconSize, metrics.EmptyMediaIconSize,
	)
	icon.Draw(canvas, m.ico, iconRect, tok.Foreground)
}

func (m *EmptyMediaWidget) Event(widget.Context, event.Event) bool { return false }
func (m *EmptyMediaWidget) Children() []widget.Widget              { return nil }

// EmptyTitle renders the empty-state title: 18px / 500 / tracking-tight,
// centered, in the foreground token.
func EmptyTitle(text string) *TypographyWidget {
	return styled(text, metrics.EmptyTitleFontSize, metrics.EmptyTitleFontWeight, metrics.EmptyTitleLineHeight).
		Align(widget.TextAlignCenter)
}

// EmptyDescription renders the empty-state description: 14px muted, centered.
func EmptyDescription(text string) *TypographyWidget {
	return styled(text, metrics.EmptyDescriptionFontSize, 400, metrics.EmptyDescriptionLineHeight).
		Muted().
		Align(widget.TextAlignCenter)
}

// drawDashedBorder strokes a dashed 1px rectangle along the four edges of
// bounds (inset by w/2 like InsideBorder). The rounded corners are approximated
// by leaving small gaps at the corners; the visible character of border-dashed
// comes from the edge dashes.
func drawDashedBorder(canvas widget.Canvas, bounds geometry.Rect, col widget.Color, w float32) {
	if w <= 0 {
		return
	}
	const dash, gap = 6, 4
	inset := w / 2
	left := bounds.Min.X + inset
	right := bounds.Max.X - inset
	top := bounds.Min.Y + inset
	bottom := bounds.Max.Y - inset

	// Horizontal edges (top, bottom).
	for x := left; x < right; x += dash + gap {
		x2 := x + dash
		if x2 > right {
			x2 = right
		}
		canvas.DrawLine(geometry.Pt(x, top), geometry.Pt(x2, top), col, w)
		canvas.DrawLine(geometry.Pt(x, bottom), geometry.Pt(x2, bottom), col, w)
	}
	// Vertical edges (left, right).
	for y := top; y < bottom; y += dash + gap {
		y2 := y + dash
		if y2 > bottom {
			y2 = bottom
		}
		canvas.DrawLine(geometry.Pt(left, y), geometry.Pt(left, y2), col, w)
		canvas.DrawLine(geometry.Pt(right, y), geometry.Pt(right, y2), col, w)
	}
}

// Compile-time interface checks.
var (
	_ widget.Widget = (*EmptyWidget)(nil)
	_ widget.Widget = (*EmptySectionWidget)(nil)
	_ widget.Widget = (*EmptyMediaWidget)(nil)
)
