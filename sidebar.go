package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// SidebarLayoutWidget arranges a sidebar panel and main content side by side.
// The sidebar sits on the left edge; the main content fills the remaining
// width. This is the top-level container for app navigation layouts.
//
// Architecture decision (DESIGN.md 3.1/3.2): graft-OWNED widget. shadcn's
// SidebarProvider + Sidebar + SidebarInset anatomy has no gogpu/ui core
// equivalent (it is a CSS-grid layout with CSS-variable-driven widths). This
// widget implements the horizontal split directly, supporting expanded and
// collapsed (icon-only) modes.
//
// Usage mirrors shadcn:
//
//	graft.SidebarLayout(
//	    graft.Sidebar(
//	        graft.SidebarHeader(graft.H4("App")),
//	        graft.SidebarContent(
//	            graft.SidebarGroup("Navigation",
//	                graft.SidebarMenuItem("Dashboard").Icon(icons.Home).Active(true),
//	                graft.SidebarMenuItem("Settings").Icon(icons.Settings),
//	            ),
//	        ),
//	        graft.SidebarFooter(graft.MutedText("v1.0")),
//	    ),
//	    mainContent,
//	)
type SidebarLayoutWidget struct {
	widget.WidgetBase

	sidebar *SidebarWidget
	main    Widget
}

// SidebarLayout creates a horizontal split: sidebar on the left, main content
// filling the rest.
func SidebarLayout(sidebar *SidebarWidget, main Widget) *SidebarLayoutWidget {
	l := &SidebarLayoutWidget{sidebar: sidebar, main: main}
	l.SetVisible(true)
	l.SetEnabled(true)
	sidebar.SetParent(l)
	if ps, ok := main.(interface{ SetParent(widget.Widget) }); ok {
		ps.SetParent(l)
	}
	return l
}

// Layout positions the sidebar at x=0 and the main content to its right.
func (l *SidebarLayoutWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	m := metrics.Sidebar

	sidebarW := m.Width
	if l.sidebar.st.isCollapsed() {
		sidebarW = m.CollapsedWidth
	}

	h := cons.MaxHeight
	if h >= geometry.Infinity {
		h = 600 // fallback for unconstrained height
	}

	// Sidebar gets its full width and available height.
	sCons := geometry.Constraints{
		MinWidth: sidebarW, MaxWidth: sidebarW,
		MinHeight: h, MaxHeight: h,
	}
	l.sidebar.Layout(ctx, sCons)
	setChildBounds(l.sidebar, geometry.NewRect(0, 0, sidebarW, h))

	// Main content fills the remaining width.
	mainW := cons.MaxWidth - sidebarW
	if mainW < 0 {
		mainW = 0
	}
	mCons := geometry.Constraints{
		MinWidth: mainW, MaxWidth: mainW,
		MinHeight: h, MaxHeight: h,
	}
	l.main.Layout(ctx, mCons)
	setChildBounds(l.main, geometry.NewRect(sidebarW, 0, mainW, h))

	size := cons.Constrain(geometry.Sz(cons.MaxWidth, h))
	l.SetBounds(geometry.FromPointSize(l.Position(), size))
	return size
}

// Draw paints the sidebar then the main content.
func (l *SidebarLayoutWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !l.IsVisible() {
		return
	}
	canvas.PushTransform(l.Bounds().Min)
	widget.StampScreenOrigin(l.sidebar, canvas)
	widget.DrawChild(l.sidebar, ctx, canvas)
	widget.StampScreenOrigin(l.main, canvas)
	widget.DrawChild(l.main, ctx, canvas)
	canvas.PopTransform()
}

// Event dispatches to the sidebar and main content.
func (l *SidebarLayoutWidget) Event(ctx widget.Context, e event.Event) bool {
	if !l.IsVisible() || !l.IsEnabled() {
		return false
	}
	return dispatchToChildren(ctx, e, l.Bounds(), []widget.Widget{l.sidebar, l.main})
}

// Children returns the sidebar and main content.
func (l *SidebarLayoutWidget) Children() []widget.Widget {
	return []widget.Widget{l.sidebar, l.main}
}

// ── SidebarWidget ───────────────────────────────────────────────────────

// SidebarWidget is the vertical navigation panel: a fixed-width column
// with header, scrollable content, and footer, backed by the sidebar-*
// theme tokens (SidebarBackground, SidebarForeground, etc.).
type SidebarWidget struct {
	widget.WidgetBase

	st       *sidebarState
	sections []Widget // header, content, footer sections
}

// sidebarState is the collapsed/expanded state shared by the sidebar tree.
type sidebarState struct {
	collapsed  bool
	sig        state.Signal[bool]
	onCollapse func(bool)
	th         *theme.Theme
	invalidate func()
}

// isCollapsed reports the current collapsed state (bound signal wins).
func (st *sidebarState) isCollapsed() bool {
	if st.sig != nil {
		return st.sig.Get()
	}
	return st.collapsed
}

// setCollapsed updates the collapsed state and fires the observer.
func (st *sidebarState) setCollapsed(v bool) {
	if st.sig != nil {
		st.sig.Set(v)
	} else {
		st.collapsed = v
	}
	if st.onCollapse != nil {
		st.onCollapse(v)
	}
	if st.invalidate != nil {
		st.invalidate()
	}
}

// Toggle flips the collapsed state (used by sidebar collapse buttons).
func (st *sidebarState) Toggle() {
	st.setCollapsed(!st.isCollapsed())
}

// Sidebar creates a sidebar from composable sections (SidebarHeader,
// SidebarContent, SidebarFooter).
func Sidebar(sections ...Widget) *SidebarWidget {
	st := &sidebarState{th: CurrentTheme()}
	s := &SidebarWidget{st: st, sections: sections}
	s.SetVisible(true)
	s.SetEnabled(true)

	// Propagate sidebar state to content sections that need it.
	for _, sec := range sections {
		if sc, ok := sec.(*SidebarContentWidget); ok {
			sc.st = st
		}
		if ps, ok := sec.(interface{ SetParent(widget.Widget) }); ok {
			ps.SetParent(s)
		}
	}
	return s
}

// Collapsed sets the initial collapsed state (uncontrolled).
func (s *SidebarWidget) Collapsed(v bool) *SidebarWidget {
	s.st.collapsed = v
	return s
}

// Bind makes the collapsed state controlled by a boolean signal.
func (s *SidebarWidget) Bind(sig state.Signal[bool]) *SidebarWidget {
	s.st.sig = sig
	return s
}

// OnCollapse registers an observer fired whenever the collapsed state changes.
func (s *SidebarWidget) OnCollapse(fn func(bool)) *SidebarWidget {
	s.st.onCollapse = fn
	return s
}

// Theme pins a specific theme instead of the snapshotted current theme.
func (s *SidebarWidget) Theme(th *theme.Theme) *SidebarWidget {
	if th != nil {
		s.st.th = th
	}
	return s
}

// Layout stacks the sections vertically within the sidebar width. The content
// section expands to fill the space between header and footer.
func (s *SidebarWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	s.st.invalidate = func() {
		s.SetNeedsRedraw(true)
		if ctx != nil {
			ctx.Invalidate()
		}
	}

	m := metrics.Sidebar
	w := m.Width
	if s.st.isCollapsed() {
		w = m.CollapsedWidth
	}
	h := cons.MaxHeight
	if h >= geometry.Infinity {
		h = 600
	}

	pad := m.SectionPad
	innerW := w - 2*pad
	if innerW < 0 {
		innerW = 0
	}

	// First pass: measure header and footer heights.
	var headerH, footerH float32
	var headerSec, footerSec, contentSec Widget
	for _, sec := range s.sections {
		switch sec.(type) {
		case *SidebarHeaderWidget:
			headerSec = sec
		case *SidebarFooterWidget:
			footerSec = sec
		case *SidebarContentWidget:
			contentSec = sec
		}
	}

	childCons := geometry.Constraints{
		MinWidth: innerW, MaxWidth: innerW,
		MinHeight: 0, MaxHeight: geometry.Infinity,
	}

	if headerSec != nil {
		sz := headerSec.Layout(ctx, childCons)
		headerH = sz.Height + pad
	}
	if footerSec != nil {
		sz := footerSec.Layout(ctx, childCons)
		footerH = sz.Height + pad
	}

	// Content fills the remaining space.
	contentH := h - headerH - footerH
	if contentH < 0 {
		contentH = 0
	}

	// Position sections.
	y := float32(0)
	if headerSec != nil {
		setChildBounds(headerSec, geometry.NewRect(pad, y+pad, innerW, headerH-pad))
		y += headerH
	}
	if contentSec != nil {
		contentCons := geometry.Constraints{
			MinWidth: innerW, MaxWidth: innerW,
			MinHeight: contentH, MaxHeight: contentH,
		}
		contentSec.Layout(ctx, contentCons)
		setChildBounds(contentSec, geometry.NewRect(pad, y, innerW, contentH))
		y += contentH
	}
	if footerSec != nil {
		setChildBounds(footerSec, geometry.NewRect(pad, y, innerW, footerH-pad))
	}

	size := cons.Constrain(geometry.Sz(w, h))
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

// Draw paints the sidebar background, sections, and the right-edge border.
func (s *SidebarWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !s.IsVisible() {
		return
	}
	tok := s.st.th.Active()
	bounds := s.Bounds()

	// Background fill.
	canvas.DrawRect(bounds, tok.Sidebar)

	// Sections.
	canvas.PushTransform(bounds.Min)
	for _, sec := range s.sections {
		widget.StampScreenOrigin(sec, canvas)
		widget.DrawChild(sec, ctx, canvas)
	}
	canvas.PopTransform()

	// Right-edge border (border-r).
	m := metrics.Sidebar
	bx := bounds.Max.X - m.BorderWidth
	canvas.DrawRect(geometry.NewRect(bx, bounds.Min.Y, m.BorderWidth, bounds.Height()), tok.SidebarBorder)
}

// Event dispatches to sections.
func (s *SidebarWidget) Event(ctx widget.Context, e event.Event) bool {
	if !s.IsVisible() || !s.IsEnabled() {
		return false
	}
	children := make([]widget.Widget, len(s.sections))
	copy(children, s.sections)
	return dispatchToChildren(ctx, e, s.Bounds(), children)
}

// Children returns the sections.
func (s *SidebarWidget) Children() []widget.Widget {
	out := make([]widget.Widget, len(s.sections))
	copy(out, s.sections)
	return out
}

// Mount binds the controlled-collapsed signal for push invalidation.
func (s *SidebarWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil || s.st.sig == nil {
		return
	}
	s.AddBinding(state.BindToScheduler(s.st.sig, s, sched))
}

// Unmount implements widget.Lifecycle; bindings clean up automatically.
func (s *SidebarWidget) Unmount() {}

// ── SidebarHeader ───────────────────────────────────────────────────────

// SidebarHeaderWidget is the top section of the sidebar, typically holding
// a logo or app title. It stacks children vertically with gap-2.
type SidebarHeaderWidget struct {
	widget.WidgetBase
	children []widget.Widget
}

// SidebarHeader creates the sidebar header section.
func SidebarHeader(children ...Widget) *SidebarHeaderWidget {
	h := &SidebarHeaderWidget{children: children}
	h.SetVisible(true)
	h.SetEnabled(true)
	for _, ch := range children {
		if ps, ok := ch.(interface{ SetParent(widget.Widget) }); ok {
			ps.SetParent(h)
		}
	}
	return h
}

// Layout stacks children vertically with gap.
func (h *SidebarHeaderWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	m := metrics.Sidebar
	w := cons.MaxWidth
	childCons := geometry.Constraints{
		MinWidth: 0, MaxWidth: w,
		MinHeight: 0, MaxHeight: geometry.Infinity,
	}
	var y float32
	for i, ch := range h.children {
		if i > 0 {
			y += m.SectionGap
		}
		sz := ch.Layout(ctx, childCons)
		setChildBounds(ch, geometry.NewRect(0, y, w, sz.Height))
		y += sz.Height
	}
	// Enforce minimum header height.
	if y < m.HeaderHeight-2*m.SectionPad {
		y = m.HeaderHeight - 2*m.SectionPad
	}
	size := cons.Constrain(geometry.Sz(w, y))
	h.SetBounds(geometry.FromPointSize(h.Position(), size))
	return size
}

// Draw renders header children.
func (h *SidebarHeaderWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !h.IsVisible() {
		return
	}
	canvas.PushTransform(h.Bounds().Min)
	for _, ch := range h.children {
		widget.StampScreenOrigin(ch, canvas)
		widget.DrawChild(ch, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event dispatches to children.
func (h *SidebarHeaderWidget) Event(ctx widget.Context, e event.Event) bool {
	return dispatchToChildren(ctx, e, h.Bounds(), h.children)
}

// Children returns the header's child widgets.
func (h *SidebarHeaderWidget) Children() []widget.Widget { return h.children }

// ── SidebarContent ──────────────────────────────────────────────────────

// SidebarContentWidget is the scrollable middle section of the sidebar,
// holding groups and menu items.
type SidebarContentWidget struct {
	widget.WidgetBase
	st       *sidebarState
	children []widget.Widget
}

// SidebarContent creates the sidebar content section from groups or items.
func SidebarContent(children ...Widget) *SidebarContentWidget {
	c := &SidebarContentWidget{children: children}
	c.SetVisible(true)
	c.SetEnabled(true)
	for _, ch := range children {
		if ps, ok := ch.(interface{ SetParent(widget.Widget) }); ok {
			ps.SetParent(c)
		}
	}
	return c
}

// Layout stacks children vertically with gap.
func (c *SidebarContentWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	m := metrics.Sidebar
	w := cons.MaxWidth
	childCons := geometry.Constraints{
		MinWidth: 0, MaxWidth: w,
		MinHeight: 0, MaxHeight: geometry.Infinity,
	}
	var y float32
	for i, ch := range c.children {
		if i > 0 {
			y += m.SectionGap
		}
		// Propagate collapsed state to groups.
		if g, ok := ch.(*SidebarGroupWidget); ok && c.st != nil {
			g.collapsed = c.st.isCollapsed()
		}
		sz := ch.Layout(ctx, childCons)
		setChildBounds(ch, geometry.NewRect(0, y, w, sz.Height))
		y += sz.Height
	}
	size := cons.Constrain(geometry.Sz(w, y))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw renders content children.
func (c *SidebarContentWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	canvas.PushTransform(c.Bounds().Min)
	for _, ch := range c.children {
		widget.StampScreenOrigin(ch, canvas)
		widget.DrawChild(ch, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event dispatches to children.
func (c *SidebarContentWidget) Event(ctx widget.Context, e event.Event) bool {
	return dispatchToChildren(ctx, e, c.Bounds(), c.children)
}

// Children returns the content's child widgets.
func (c *SidebarContentWidget) Children() []widget.Widget { return c.children }

// ── SidebarFooter ───────────────────────────────────────────────────────

// SidebarFooterWidget is the bottom section of the sidebar, typically
// holding version info or user controls.
type SidebarFooterWidget struct {
	widget.WidgetBase
	children []widget.Widget
}

// SidebarFooter creates the sidebar footer section.
func SidebarFooter(children ...Widget) *SidebarFooterWidget {
	f := &SidebarFooterWidget{children: children}
	f.SetVisible(true)
	f.SetEnabled(true)
	for _, ch := range children {
		if ps, ok := ch.(interface{ SetParent(widget.Widget) }); ok {
			ps.SetParent(f)
		}
	}
	return f
}

// Layout stacks children vertically with gap.
func (f *SidebarFooterWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	m := metrics.Sidebar
	w := cons.MaxWidth
	childCons := geometry.Constraints{
		MinWidth: 0, MaxWidth: w,
		MinHeight: 0, MaxHeight: geometry.Infinity,
	}
	var y float32
	for i, ch := range f.children {
		if i > 0 {
			y += m.SectionGap
		}
		sz := ch.Layout(ctx, childCons)
		setChildBounds(ch, geometry.NewRect(0, y, w, sz.Height))
		y += sz.Height
	}
	size := cons.Constrain(geometry.Sz(w, y))
	f.SetBounds(geometry.FromPointSize(f.Position(), size))
	return size
}

// Draw renders footer children.
func (f *SidebarFooterWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !f.IsVisible() {
		return
	}
	canvas.PushTransform(f.Bounds().Min)
	for _, ch := range f.children {
		widget.StampScreenOrigin(ch, canvas)
		widget.DrawChild(ch, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event dispatches to children.
func (f *SidebarFooterWidget) Event(ctx widget.Context, e event.Event) bool {
	return dispatchToChildren(ctx, e, f.Bounds(), f.children)
}

// Children returns the footer's child widgets.
func (f *SidebarFooterWidget) Children() []widget.Widget { return f.children }

// ── SidebarGroup ────────────────────────────────────────────────────────

// SidebarGroupWidget groups menu items under an optional label inside the
// sidebar content area.
type SidebarGroupWidget struct {
	widget.WidgetBase

	label     string
	items     []*SidebarMenuItemWidget
	collapsed bool // inherited from the sidebar collapsed state
	theme     *theme.Theme
}

// SidebarGroup creates a group of menu items with an optional label. Pass ""
// for label to create an unlabeled group.
func SidebarGroup(label string, items ...*SidebarMenuItemWidget) *SidebarGroupWidget {
	g := &SidebarGroupWidget{
		label: label,
		items: items,
		theme: CurrentTheme(),
	}
	g.SetVisible(true)
	g.SetEnabled(true)
	for _, it := range items {
		it.theme = g.theme
		it.SetParent(g)
	}
	return g
}

// fontFamily resolves the group label font family.
func (g *SidebarGroupWidget) fontFamily() string {
	th := g.theme
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	return fonts.Family(metrics.Sidebar.GroupLabelWeight)
}

// Layout stacks the label (when expanded) and items vertically with GroupGap.
func (g *SidebarGroupWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	m := metrics.Sidebar
	w := cons.MaxWidth

	// Propagate collapsed state to items.
	for _, it := range g.items {
		it.collapsed = g.collapsed
	}

	var y float32

	// Group label (hidden when collapsed).
	if g.label != "" && !g.collapsed {
		labelH := m.GroupLabelSize + 2*m.GroupLabelPadY
		y += labelH
	}

	// Items.
	itemCons := geometry.Constraints{
		MinWidth: w, MaxWidth: w,
		MinHeight: 0, MaxHeight: geometry.Infinity,
	}
	for i, it := range g.items {
		if i > 0 || (g.label != "" && !g.collapsed) {
			y += m.GroupGap
		}
		sz := it.Layout(ctx, itemCons)
		setChildBounds(it, geometry.NewRect(0, y, sz.Width, sz.Height))
		y += sz.Height
	}

	size := cons.Constrain(geometry.Sz(w, y))
	g.SetBounds(geometry.FromPointSize(g.Position(), size))
	return size
}

// Draw paints the group label and items.
func (g *SidebarGroupWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !g.IsVisible() {
		return
	}
	m := metrics.Sidebar
	tok := g.theme.Active()
	bounds := g.Bounds()

	canvas.PushTransform(bounds.Min)

	// Group label (hidden when collapsed).
	if g.label != "" && !g.collapsed {
		family := g.fontFamily()
		labelRect := geometry.NewRect(
			m.ItemPadX,
			m.GroupLabelPadY,
			bounds.Width()-2*m.ItemPadX,
			m.GroupLabelSize,
		)
		labelCol := draw.MulAlpha(tok.SidebarForeground, 0.5) // text-sidebar-foreground/50
		if std, ok := canvas.(widget.StyledTextDrawer); ok {
			std.DrawStyledText(g.label, labelRect, widget.TextStyle{
				FontFamily: family,
				FontSize:   m.GroupLabelSize,
				Color:      labelCol,
				Align:      widget.TextAlignLeft,
			})
		} else {
			canvas.DrawText(g.label, labelRect, m.GroupLabelSize, labelCol,
				m.GroupLabelWeight >= 600, widget.TextAlignLeft)
		}
	}

	// Items.
	for _, it := range g.items {
		widget.StampScreenOrigin(it, canvas)
		widget.DrawChild(it, ctx, canvas)
	}

	canvas.PopTransform()
}

// Event dispatches to items.
func (g *SidebarGroupWidget) Event(ctx widget.Context, e event.Event) bool {
	children := make([]widget.Widget, len(g.items))
	for i, it := range g.items {
		children[i] = it
	}
	return dispatchToChildren(ctx, e, g.Bounds(), children)
}

// Children returns the items.
func (g *SidebarGroupWidget) Children() []widget.Widget {
	out := make([]widget.Widget, len(g.items))
	for i, it := range g.items {
		out[i] = it
	}
	return out
}

// ── SidebarMenuItem ─────────────────────────────────────────────────────

// SidebarMenuItemWidget is a single navigation entry in the sidebar: a
// rounded row with an optional leading icon and label text. The active
// item gets the sidebar-accent background and sidebar-accent-foreground
// text color.
type SidebarMenuItemWidget struct {
	widget.WidgetBase

	label     string
	iconData  *icon.IconData
	active    bool
	onClick   func()
	collapsed bool // inherited from the sidebar collapsed state
	theme     *theme.Theme

	hovered      bool
	pressed      bool
	pointerFocus bool
	focusVisible bool
}

// SidebarMenuItem creates a sidebar menu entry with the given label.
func SidebarMenuItem(label string) *SidebarMenuItemWidget {
	it := &SidebarMenuItemWidget{
		label: label,
		theme: CurrentTheme(),
	}
	it.SetVisible(true)
	it.SetEnabled(true)
	return it
}

// Icon adds a leading icon to the menu item.
func (it *SidebarMenuItemWidget) Icon(ic icon.IconData) *SidebarMenuItemWidget {
	it.iconData = &ic
	return it
}

// Active marks the item as the currently selected/active entry.
func (it *SidebarMenuItemWidget) Active(v bool) *SidebarMenuItemWidget {
	it.active = v
	return it
}

// OnClick sets the click handler.
func (it *SidebarMenuItemWidget) OnClick(fn func()) *SidebarMenuItemWidget {
	it.onClick = fn
	return it
}

// fontFamily resolves the item font family.
func (it *SidebarMenuItemWidget) fontFamily() string {
	th := it.theme
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	weight := metrics.Sidebar.FontWeight
	if it.active {
		weight = metrics.Sidebar.ActiveFontWeight
	}
	return fonts.Family(weight)
}

// Layout sizes the item to full width and fixed item height.
func (it *SidebarMenuItemWidget) Layout(_ widget.Context, cons geometry.Constraints) geometry.Size {
	m := metrics.Sidebar
	w := cons.MaxWidth
	if w >= geometry.Infinity {
		// Estimate natural width from label.
		lw := textmetrics.Width(it.fontFamily(), m.FontSize, it.label)
		w = 2*m.ItemPadX + lw
		if it.iconData != nil {
			w += m.IconSize + m.IconGap
		}
	}
	size := cons.Constrain(geometry.Sz(w, m.ItemHeight))
	it.SetBounds(geometry.FromPointSize(it.Position(), size))
	return size
}

// Draw paints the item: hover/active background, icon, and label.
func (it *SidebarMenuItemWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !it.IsVisible() {
		return
	}
	m := metrics.Sidebar
	tok := it.theme.Active()
	bounds := it.Bounds()

	// Background: active or hovered gets sidebar-accent fill.
	if it.active || it.hovered {
		canvas.DrawRoundRect(bounds, tok.SidebarAccent, m.ItemRadius)
	}

	// Resolve text/icon color.
	textCol := tok.SidebarForeground
	if it.active || it.hovered {
		textCol = tok.SidebarAccentForeground
	}

	x := bounds.Min.X + m.ItemPadX
	centerY := bounds.Min.Y + bounds.Height()/2

	// Icon.
	if it.iconData != nil {
		iconTop := centerY - m.IconSize/2
		iconRect := geometry.NewRect(x, iconTop, m.IconSize, m.IconSize)
		icon.Draw(canvas, *it.iconData, iconRect, textCol)
		x += m.IconSize + m.IconGap
	}

	// Label (hidden when collapsed, unless no icon).
	if !it.collapsed || it.iconData == nil {
		family := it.fontFamily()
		labelRect := geometry.NewRect(
			x,
			centerY-m.FontSize/2,
			bounds.Max.X-x-m.ItemPadX,
			m.FontSize,
		)
		if std, ok := canvas.(widget.StyledTextDrawer); ok {
			std.DrawStyledText(it.label, labelRect, widget.TextStyle{
				FontFamily: family,
				FontSize:   m.FontSize,
				Color:      textCol,
				Align:      widget.TextAlignLeft,
			})
		} else {
			weight := m.FontWeight
			if it.active {
				weight = m.ActiveFontWeight
			}
			canvas.DrawText(it.label, labelRect, m.FontSize, textCol,
				weight >= 600, widget.TextAlignLeft)
		}
	}

	// Focus ring.
	if it.focusVisible {
		draw.InsideBorder(canvas, bounds, m.ItemRadius, tok.SidebarRing, 1)
		draw.FocusRing(canvas, bounds, m.ItemRadius, draw.Alpha(tok.SidebarRing, metrics.RingAlpha))
	}
}

// Event handles hover, click, keyboard activation, and focus tracking.
func (it *SidebarMenuItemWidget) Event(ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.MouseEvent:
		return it.mouseEvent(ctx, ev)
	case *event.KeyEvent:
		if it.IsFocused() && ev.KeyType == event.KeyPress &&
			(ev.Key == event.KeyEnter || ev.Key == event.KeySpace) {
			if it.onClick != nil {
				it.onClick()
			}
			return true
		}
	}
	// NOTE: *event.FocusEvent is window-level (OS focus gained/lost, no target
	// widget) and is broadcast to every item; consuming it here marked all items
	// focus-visible whenever the window was focused (focus rings rendered as
	// solid boxes, gg#369). Per-widget focus is driven by ctx.RequestFocus.
	return false
}

func (it *SidebarMenuItemWidget) mouseEvent(ctx widget.Context, ev *event.MouseEvent) bool {
	switch ev.MouseType {
	case event.MouseEnter:
		it.hovered = true
		if ctx != nil {
			ctx.SetCursor(widget.CursorPointer)
			ctx.InvalidateRect(it.Bounds())
		}
		it.SetNeedsRedraw(true)
		return true
	case event.MouseLeave:
		it.hovered = false
		it.pressed = false
		if ctx != nil {
			ctx.SetCursor(widget.CursorDefault)
			ctx.InvalidateRect(it.Bounds())
		}
		it.SetNeedsRedraw(true)
		return true
	case event.MousePress:
		if ev.Button != event.ButtonLeft {
			return false
		}
		it.pressed = true
		it.pointerFocus = true
		if ctx != nil {
			ctx.RequestFocus(it)
		}
		it.pointerFocus = false
		return true
	case event.MouseRelease:
		wasPressed := it.pressed
		it.pressed = false
		if wasPressed && it.Bounds().Contains(ev.Position) {
			if it.onClick != nil {
				it.onClick()
			}
			return true
		}
	}
	return false
}

// SetFocused tracks focus-visible: the ring renders only when focus did not
// arrive from a mouse press.
func (it *SidebarMenuItemWidget) SetFocused(focused bool) {
	it.WidgetBase.SetFocused(focused)
	if focused {
		it.focusVisible = !it.pointerFocus
	} else {
		it.focusVisible = false
	}
	it.MarkRedrawLocal() // not SetNeedsRedraw: avoids context-lock re-entry in RequestFocus
}

// IsFocusable reports keyboard focusability.
func (it *SidebarMenuItemWidget) IsFocusable() bool {
	return it.IsVisible() && it.IsEnabled()
}

// Children returns nil; the menu item is a leaf.
func (it *SidebarMenuItemWidget) Children() []widget.Widget { return nil }

// Compile-time interface checks.
var (
	_ widget.Widget    = (*SidebarLayoutWidget)(nil)
	_ widget.Widget    = (*SidebarWidget)(nil)
	_ widget.Lifecycle = (*SidebarWidget)(nil)
	_ widget.Widget    = (*SidebarHeaderWidget)(nil)
	_ widget.Widget    = (*SidebarContentWidget)(nil)
	_ widget.Widget    = (*SidebarFooterWidget)(nil)
	_ widget.Widget    = (*SidebarGroupWidget)(nil)
	_ widget.Widget    = (*SidebarMenuItemWidget)(nil)
	_ widget.Focusable = (*SidebarMenuItemWidget)(nil)
)
