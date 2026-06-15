package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// ResizableDirection selects the axis panels are arranged along.
type ResizableDirection uint8

// Resizable directions (default Horizontal, matching shadcn's
// direction="horizontal").
const (
	// ResizableHorizontal arranges panels left-to-right with vertical
	// dividers (flex row).
	ResizableHorizontal ResizableDirection = iota
	// ResizableVertical arranges panels top-to-bottom with horizontal
	// dividers (flex column).
	ResizableVertical
)

// ResizablePanelGroupWidget is the shadcn Resizable component: a flex row or
// column of ResizablePanels separated by draggable ResizableHandles
// (react-resizable-panels in shadcn).
//
// Architecture decision (DESIGN.md sections 3.1/7): graft-OWNED. gogpu/ui's
// core/splitview models exactly two panels with one divider, while shadcn's
// ResizablePanelGroup takes N panels with a handle between each, tracks per-
// panel ratios, and resizes the two panels adjacent to the dragged handle.
// The N-panel ratio bookkeeping and per-handle drag cannot be expressed by
// wrapping the two-panel core, so the group owns its Layout (ratio split),
// Draw, and drag handling — modeled on slider.go's drag math and the
// dispatchToChildren hit-test convention.
//
// Usage mirrors shadcn:
//
//	graft.ResizablePanelGroup(graft.ResizableHorizontal,
//	    graft.ResizablePanel(leftBody),
//	    graft.ResizableHandle(),
//	    graft.ResizablePanel(rightBody),
//	)
type ResizablePanelGroupWidget struct {
	widget.WidgetBase

	dir   ResizableDirection
	items []Widget // alternating panel, handle, panel, ...
	th    *theme.Theme

	panels  []*ResizablePanelWidget
	handles []*ResizableHandleWidget

	// ratios holds the fractional size of each panel (sums to 1). Lazily
	// initialized to equal shares, then mutated by handle drags.
	ratios []float32

	// dragging is the index into handles of the handle currently being
	// dragged, or -1.
	dragging   int
	dragStartP geometry.Point
	dragRatios []float32 // ratios snapshot at drag start
}

// ResizablePanelGroup assembles a panel group from alternating
// ResizablePanel and ResizableHandle children along the given direction.
func ResizablePanelGroup(dir ResizableDirection, items ...Widget) *ResizablePanelGroupWidget {
	g := &ResizablePanelGroupWidget{
		dir:      dir,
		items:    items,
		th:       CurrentTheme(),
		dragging: -1,
	}
	g.SetVisible(true)
	g.SetEnabled(true)
	for _, it := range items {
		switch w := it.(type) {
		case *ResizablePanelWidget:
			g.panels = append(g.panels, w)
		case *ResizableHandleWidget:
			w.group = g
			g.handles = append(g.handles, w)
		}
		if ps, ok := it.(interface{ SetParent(widget.Widget) }); ok {
			ps.SetParent(g)
		}
	}
	return g
}

// Theme pins a specific theme instead of the snapshotted current theme.
func (g *ResizablePanelGroupWidget) Theme(th *theme.Theme) *ResizablePanelGroupWidget {
	g.th = th
	for _, h := range g.handles {
		h.th = th
	}
	return g
}

// horizontal reports whether the group lays panels out along the x axis.
func (g *ResizablePanelGroupWidget) horizontal() bool {
	return g.dir == ResizableHorizontal
}

// ensureRatios seeds equal panel shares on first use, honoring per-panel
// DefaultSize hints where given.
func (g *ResizablePanelGroupWidget) ensureRatios() {
	if len(g.ratios) == len(g.panels) && len(g.ratios) > 0 {
		return
	}
	n := len(g.panels)
	g.ratios = make([]float32, n)
	if n == 0 {
		return
	}
	var assigned float32
	var unset int
	for i, p := range g.panels {
		if p.defaultSize > 0 {
			g.ratios[i] = p.defaultSize
			assigned += p.defaultSize
		} else {
			g.ratios[i] = -1
			unset++
		}
	}
	rem := 1 - assigned
	if rem < 0 {
		rem = 0
	}
	share := rem
	if unset > 0 {
		share = rem / float32(unset)
	}
	for i := range g.ratios {
		if g.ratios[i] < 0 {
			g.ratios[i] = share
		}
	}
}

// handleThickness is the divider hit thickness reserved between panels.
func (g *ResizablePanelGroupWidget) handleThickness() float32 {
	return metrics.Resizable.HitWidth
}

// Layout splits the group's main-axis extent between panels by ratio,
// reserving the handle hit-width for each divider.
func (g *ResizablePanelGroupWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	g.ensureRatios()

	w := c.MaxWidth
	h := c.MaxHeight
	if w <= 0 || w >= geometry.Infinity {
		w = metrics.Resizable.DefaultWidth
	}
	if h <= 0 || h >= geometry.Infinity {
		h = metrics.Resizable.DefaultHeight
	}
	size := geometry.Sz(w, h)

	n := len(g.panels)
	nh := len(g.handles)
	thick := g.handleThickness()

	var mainExtent, crossExtent float32
	if g.horizontal() {
		mainExtent = w
		crossExtent = h
	} else {
		mainExtent = h
		crossExtent = w
	}
	avail := mainExtent - thick*float32(nh)
	if avail < 0 {
		avail = 0
	}

	// Place panels and handles interleaved along the main axis.
	var cursor float32
	pi, hi := 0, 0
	for _, it := range g.items {
		switch wdg := it.(type) {
		case *ResizablePanelWidget:
			extent := avail * g.ratios[pi]
			var rect geometry.Rect
			if g.horizontal() {
				rect = geometry.NewRect(cursor, 0, extent, crossExtent)
			} else {
				rect = geometry.NewRect(0, cursor, crossExtent, extent)
			}
			wdg.layoutInside(ctx, rect)
			cursor += extent
			pi++
		case *ResizableHandleWidget:
			var rect geometry.Rect
			if g.horizontal() {
				rect = geometry.NewRect(cursor, 0, thick, crossExtent)
			} else {
				rect = geometry.NewRect(0, cursor, crossExtent, thick)
			}
			wdg.SetBounds(rect)
			cursor += thick
			hi++
		}
	}
	_ = n

	g.SetBounds(geometry.FromPointSize(g.Position(), size))
	return size
}

// Draw paints panels and dividers in local coordinates.
func (g *ResizablePanelGroupWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !g.IsVisible() {
		return
	}
	canvas.PushTransform(g.Bounds().Min)
	for _, it := range g.items {
		widget.StampScreenOrigin(it, canvas)
		widget.DrawChild(it, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event drives handle drags (which resize the adjacent panels) and forwards
// other input to the children via hit-testing.
func (g *ResizablePanelGroupWidget) Event(ctx widget.Context, e event.Event) bool {
	if !g.IsVisible() || !g.IsEnabled() {
		return false
	}
	if me, ok := e.(*event.MouseEvent); ok {
		local := *me
		local.Position = me.Position.Sub(g.Bounds().Min)
		if g.handleMouse(ctx, &local) {
			return true
		}
		// Fall through to children with the local position.
		return dispatchToChildren(ctx, &local, geometry.NewRect(0, 0, g.Bounds().Width(), g.Bounds().Height()), g.items)
	}
	return dispatchToChildren(ctx, e, g.Bounds(), g.items)
}

// handleMouse implements drag-to-resize on the dividers.
func (g *ResizablePanelGroupWidget) handleMouse(ctx widget.Context, me *event.MouseEvent) bool {
	switch me.MouseType {
	case event.MousePress:
		if me.Button != event.ButtonLeft {
			return false
		}
		for i, h := range g.handles {
			if h.Bounds().Contains(me.Position) {
				g.dragging = i
				g.dragStartP = me.Position
				g.dragRatios = append([]float32(nil), g.ratios...)
				h.dragging = true
				h.SetNeedsRedraw(true)
				ctx.Invalidate()
				return true
			}
		}
		return false
	case event.MouseMove, event.MouseDrag:
		if g.dragging < 0 {
			return false
		}
		if !me.Buttons.IsLeftPressed() {
			g.endDrag(ctx)
			return false
		}
		g.updateDrag(ctx, me.Position)
		return true
	case event.MouseRelease:
		if g.dragging < 0 {
			return false
		}
		g.endDrag(ctx)
		return true
	}
	return false
}

// updateDrag recomputes the two ratios adjacent to the dragged handle from
// the pointer delta along the main axis.
func (g *ResizablePanelGroupWidget) updateDrag(ctx widget.Context, pos geometry.Point) {
	idx := g.dragging // handle index == left-panel index
	if idx < 0 || idx+1 >= len(g.panels) {
		return
	}
	nh := len(g.handles)
	var mainExtent, delta float32
	if g.horizontal() {
		mainExtent = g.Bounds().Width()
		delta = pos.X - g.dragStartP.X
	} else {
		mainExtent = g.Bounds().Height()
		delta = pos.Y - g.dragStartP.Y
	}
	avail := mainExtent - g.handleThickness()*float32(nh)
	if avail <= 0 {
		return
	}
	dRatio := delta / avail

	left := g.dragRatios[idx] + dRatio
	right := g.dragRatios[idx+1] - dRatio
	// Clamp so neither adjacent panel goes negative; the pair's total is
	// preserved so every other panel keeps its ratio.
	pairTotal := g.dragRatios[idx] + g.dragRatios[idx+1]
	if left < 0 {
		left = 0
		right = pairTotal
	}
	if right < 0 {
		right = 0
		left = pairTotal
	}
	if g.ratios[idx] == left && g.ratios[idx+1] == right {
		return
	}
	g.ratios[idx] = left
	g.ratios[idx+1] = right
	g.SetNeedsRedraw(true)
	ctx.Invalidate()
}

// endDrag clears the active drag state.
func (g *ResizablePanelGroupWidget) endDrag(ctx widget.Context) {
	if g.dragging >= 0 && g.dragging < len(g.handles) {
		g.handles[g.dragging].dragging = false
		g.handles[g.dragging].SetNeedsRedraw(true)
	}
	g.dragging = -1
	g.dragRatios = nil
	if ctx != nil {
		ctx.Invalidate()
	}
}

// Ratios returns a copy of the current panel ratios (for tests/inspection).
func (g *ResizablePanelGroupWidget) Ratios() []float32 {
	g.ensureRatios()
	return append([]float32(nil), g.ratios...)
}

// Children returns the interleaved panels and handles.
func (g *ResizablePanelGroupWidget) Children() []widget.Widget { return g.items }

// ── ResizablePanel ────────────────────────────────────────────────────────

// ResizablePanelWidget hosts one panel's content, sized by its group ratio.
type ResizablePanelWidget struct {
	widget.WidgetBase

	content     Widget
	defaultSize float32 // initial ratio hint (0 = equal share)
}

// ResizablePanel wraps content as a resizable panel.
func ResizablePanel(content Widget) *ResizablePanelWidget {
	p := &ResizablePanelWidget{content: content}
	p.SetVisible(true)
	p.SetEnabled(true)
	if content != nil {
		if ps, ok := content.(interface{ SetParent(widget.Widget) }); ok {
			ps.SetParent(p)
		}
	}
	return p
}

// DefaultSize sets the panel's initial ratio (0..1). Unset panels share the
// remaining space equally (react-resizable-panels defaultSize is a percent;
// graft uses a 0..1 fraction).
func (p *ResizablePanelWidget) DefaultSize(ratio float32) *ResizablePanelWidget {
	p.defaultSize = ratio
	return p
}

// layoutInside positions the panel and its content within rect.
func (p *ResizablePanelWidget) layoutInside(ctx widget.Context, rect geometry.Rect) {
	p.SetBounds(rect)
	if p.content == nil {
		return
	}
	p.content.Layout(ctx, geometry.Tight(rect.Size()))
	setWidgetBounds(p.content, geometry.FromPointSize(rect.Min, rect.Size()))
}

// Layout supports standalone use; the group drives sizing via layoutInside.
func (p *ResizablePanelWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	var sz geometry.Size
	if p.content != nil {
		sz = p.content.Layout(ctx, c)
		setWidgetBounds(p.content, geometry.FromPointSize(p.Position(), sz))
	} else {
		sz = c.Constrain(geometry.Sz(0, 0))
	}
	p.SetBounds(geometry.FromPointSize(p.Position(), sz))
	return sz
}

// Draw renders the panel content.
func (p *ResizablePanelWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !p.IsVisible() || p.content == nil {
		return
	}
	p.content.Draw(ctx, canvas)
}

// Event forwards to the content.
func (p *ResizablePanelWidget) Event(ctx widget.Context, e event.Event) bool {
	if p.content == nil {
		return false
	}
	return p.content.Event(ctx, e)
}

// Children returns the content.
func (p *ResizablePanelWidget) Children() []widget.Widget {
	if p.content == nil {
		return nil
	}
	return []widget.Widget{p.content}
}

// ── ResizableHandle ───────────────────────────────────────────────────────

// ResizableHandleWidget is the draggable divider between two panels: a 1px
// --border line, optionally carrying a grip chip (GripVertical) when
// WithHandle is set.
type ResizableHandleWidget struct {
	widget.WidgetBase

	withHandle bool
	group      *ResizablePanelGroupWidget
	th         *theme.Theme

	dragging bool
	hovered  bool
}

// ResizableHandle creates a divider between two panels.
func ResizableHandle() *ResizableHandleWidget {
	h := &ResizableHandleWidget{th: CurrentTheme()}
	h.SetVisible(true)
	h.SetEnabled(true)
	return h
}

// WithHandle adds the centered grip chip with a GripVertical glyph.
func (h *ResizableHandleWidget) WithHandle() *ResizableHandleWidget {
	h.withHandle = true
	return h
}

// resolvedTheme prefers the handle's own theme, then the group's.
func (h *ResizableHandleWidget) resolvedTheme() *theme.Theme {
	if h.th != nil {
		return h.th
	}
	if h.group != nil && h.group.th != nil {
		return h.group.th
	}
	return CurrentTheme()
}

// horizontal reports whether the parent group lays out along x (so this
// divider is a vertical line).
func (h *ResizableHandleWidget) horizontal() bool {
	return h.group == nil || h.group.horizontal()
}

// Layout reports the hit-width thickness; the group positions it.
func (h *ResizableHandleWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	thick := metrics.Resizable.HitWidth
	var size geometry.Size
	if h.horizontal() {
		cross := c.MaxHeight
		if cross <= 0 || cross >= geometry.Infinity {
			cross = metrics.Resizable.GripLongSide
		}
		size = geometry.Sz(thick, cross)
	} else {
		cross := c.MaxWidth
		if cross <= 0 || cross >= geometry.Infinity {
			cross = metrics.Resizable.GripLongSide
		}
		size = geometry.Sz(cross, thick)
	}
	h.SetBounds(geometry.FromPointSize(h.Position(), size))
	return size
}

// Draw paints the 1px border line (centered in the hit band) and, when
// requested, the centered grip chip.
func (h *ResizableHandleWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !h.IsVisible() {
		return
	}
	th := h.resolvedTheme()
	tok := th.Active()
	b := h.Bounds()
	line := metrics.Resizable.LineWidth

	var lineRect geometry.Rect
	if h.horizontal() {
		// Vertical line centered in the hit band.
		cx := b.Min.X + (b.Width()-line)/2
		lineRect = geometry.NewRect(cx, b.Min.Y, line, b.Height())
	} else {
		cy := b.Min.Y + (b.Height()-line)/2
		lineRect = geometry.NewRect(b.Min.X, cy, b.Width(), line)
	}
	canvas.DrawRect(lineRect, tok.Border)

	if h.withHandle {
		h.drawGrip(canvas, b, th, tok)
	}
}

// drawGrip paints the centered 16×12 (h-4 w-3) rounded-xs grip chip with a
// 1px border in --border and a 10px GripVertical glyph; rotated 90° for a
// horizontal divider.
func (h *ResizableHandleWidget) drawGrip(canvas widget.Canvas, b geometry.Rect, th *theme.Theme, tok *theme.Tokens) {
	long := metrics.Resizable.GripLongSide
	short := metrics.Resizable.GripShortSide
	var gw, gh float32
	var glyph icon.IconData = icons.GripVertical
	if h.horizontal() {
		// Vertical divider: tall chip, vertical grip dots.
		gw, gh = short, long
	} else {
		// Horizontal divider: wide chip; shadcn rotates the icon 90°. The
		// vendored set has no grip-horizontal, so reuse GripVertical inside a
		// wide chip (dots read horizontally enough at 10px).
		gw, gh = long, short
		glyph = icons.GripVertical
	}
	cx := b.Min.X + b.Width()/2
	cy := b.Min.Y + b.Height()/2
	chip := geometry.NewRect(cx-gw/2, cy-gh/2, gw, gh)

	draw.BorderFill(canvas, chip, tok.Border, tok.Border, th.RadiusXS(), metrics.Resizable.GripBorderWidth)

	ic := metrics.Resizable.GripIconSize
	iconRect := geometry.NewRect(cx-ic/2, cy-ic/2, ic, ic)
	icon.Draw(canvas, glyph, iconRect, tok.MutedForeground)
}

// Event tracks hover (for the resize cursor); the group owns the drag.
func (h *ResizableHandleWidget) Event(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	switch me.MouseType {
	case event.MouseEnter, event.MouseMove:
		if !h.hovered {
			h.hovered = true
		}
		if ctx != nil {
			if h.horizontal() {
				ctx.SetCursor(widget.CursorResizeEW)
			} else {
				ctx.SetCursor(widget.CursorResizeNS)
			}
		}
		return true
	case event.MouseLeave:
		if h.hovered {
			h.hovered = false
			if ctx != nil {
				ctx.SetCursor(widget.CursorDefault)
			}
		}
		return false
	}
	return false
}

// Children returns nil; the handle is a leaf.
func (h *ResizableHandleWidget) Children() []widget.Widget { return nil }

// Compile-time interface checks.
var (
	_ widget.Widget = (*ResizablePanelGroupWidget)(nil)
	_ widget.Widget = (*ResizablePanelWidget)(nil)
	_ widget.Widget = (*ResizableHandleWidget)(nil)
)
