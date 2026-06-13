package graft

import (
	"github.com/gogpu/ui/core/textfield"
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

// InputGroupWidget is shadcn's InputGroup: a single bordered, rounded container
// (the Input chrome) that hosts a borderless text field plus leading and/or
// trailing addons — an icon, a short text label, or a small button — sharing
// one focus ring (shadcn/ui input-group).
//
// WRAP DECISION (DESIGN.md §3.2): graft-OWNED composite. It draws the container
// chrome (shadow-xs, dark Input/30 fill, 1px Input border, focus/invalid ring)
// once and lays addons + a borderless core/textfield inside. The inner field is
// a wrapped core/textfield (text-editing machinery is substantial, same as
// Input) but painted by a private borderless painter that draws only the text,
// placeholder, selection, and caret — the container owns the border.
type InputGroupWidget struct {
	widget.WidgetBase

	field *textfield.Widget
	theme *theme.Theme

	leading  *InputGroupAddonWidget
	trailing *InputGroupAddonWidget

	userOpts []textfield.Option
	value    string
	hasValue bool

	invalid  bool
	disabled bool
	width    float32
}

// addonKind selects what an addon renders.
type addonKind uint8

const (
	addonIcon addonKind = iota
	addonText
	addonButton
)

// InputGroupAddonWidget is a leading/trailing element inside an InputGroup: an
// icon, a text label, or a button.
type InputGroupAddonWidget struct {
	kind   addonKind
	icon   icon.IconData
	text   string
	button *ButtonWidget
}

// InputGroupAddon creates a 16px muted icon addon.
func InputGroupAddon(ic icon.IconData) *InputGroupAddonWidget {
	return &InputGroupAddonWidget{kind: addonIcon, icon: ic}
}

// InputGroupText creates a muted text addon (e.g. a unit suffix like "kg").
func InputGroupText(s string) *InputGroupAddonWidget {
	return &InputGroupAddonWidget{kind: addonText, text: s}
}

// InputGroupButton creates a button addon (rendered at sm size inside the
// group). The button is reconfigured to the sm size.
func InputGroupButton(b *ButtonWidget) *InputGroupAddonWidget {
	if b != nil {
		b.Sm()
	}
	return &InputGroupAddonWidget{kind: addonButton, button: b}
}

// InputGroup creates an input group hosting a borderless text field.
func InputGroup() *InputGroupWidget {
	g := &InputGroupWidget{theme: CurrentTheme()}
	g.SetVisible(true)
	g.SetEnabled(true)
	g.rebuild()
	return g
}

// Leading sets the addon rendered before the text field.
func (g *InputGroupWidget) Leading(a *InputGroupAddonWidget) *InputGroupWidget {
	g.leading = a
	if a != nil && a.button != nil {
		a.button.SetParent(g)
	}
	return g
}

// Trailing sets the addon rendered after the text field.
func (g *InputGroupWidget) Trailing(a *InputGroupAddonWidget) *InputGroupWidget {
	g.trailing = a
	if a != nil && a.button != nil {
		a.button.SetParent(g)
	}
	return g
}

// Placeholder sets the empty-state placeholder text.
func (g *InputGroupWidget) Placeholder(s string) *InputGroupWidget {
	g.userOpts = append(g.userOpts, textfield.Placeholder(s))
	g.rebuild()
	return g
}

// Value sets the initial (uncontrolled) text value.
func (g *InputGroupWidget) Value(s string) *InputGroupWidget {
	g.value = s
	g.hasValue = true
	g.rebuild()
	return g
}

// Bind makes the field controlled by sig.
func (g *InputGroupWidget) Bind(sig state.Signal[string]) *InputGroupWidget {
	g.userOpts = append(g.userOpts, textfield.ValueSignal(sig))
	g.rebuild()
	return g
}

// OnChange registers a callback fired on every text change.
func (g *InputGroupWidget) OnChange(fn func(string)) *InputGroupWidget {
	g.userOpts = append(g.userOpts, textfield.OnChange(fn))
	g.rebuild()
	return g
}

// Disabled sets the disabled state.
func (g *InputGroupWidget) Disabled(v bool) *InputGroupWidget {
	g.disabled = v
	g.userOpts = append(g.userOpts, textfield.Disabled(v))
	g.rebuild()
	return g
}

// Invalid sets the aria-invalid state (destructive border + ring).
func (g *InputGroupWidget) Invalid(v bool) *InputGroupWidget {
	g.invalid = v
	return g
}

// W sets an explicit container width in px.
func (g *InputGroupWidget) W(px float32) *InputGroupWidget {
	g.width = px
	return g
}

// Theme pins a specific theme.
func (g *InputGroupWidget) Theme(th *theme.Theme) *InputGroupWidget {
	if th != nil {
		g.theme = th
		g.rebuild()
	}
	return g
}

// rebuild replaces the inner field with the borderless painter plus the
// accumulated caller options.
func (g *InputGroupWidget) rebuild() {
	opts := make([]textfield.Option, 0, len(g.userOpts)+1)
	opts = append(opts, textfield.PainterOpt(inputGroupFieldPainter{theme: g.theme, disabled: g.disabled}))
	opts = append(opts, g.userOpts...)
	g.field = textfield.New(opts...)
	if g.hasValue {
		g.field.SetText(g.value)
	}
}

// Text returns the current field text.
func (g *InputGroupWidget) Text() string { return g.field.Text() }

// leadingWidth returns the px consumed by the leading addon + its gap, or 0.
func (g *InputGroupWidget) addonWidth(a *InputGroupAddonWidget, ctx widget.Context) float32 {
	if a == nil {
		return 0
	}
	switch a.kind {
	case addonIcon:
		return metrics.InputGroup.IconSize
	case addonText:
		return textmetrics.Width(g.fontFamily(), metrics.InputGroup.FontSize, a.text)
	case addonButton:
		if a.button == nil {
			return 0
		}
		sz := a.button.Layout(ctx, geometry.Loose(geometry.Sz(geometry.Infinity, metrics.InputGroup.Height)))
		return sz.Width
	}
	return 0
}

func (g *InputGroupWidget) fontFamily() string {
	if g.theme.FontSans != theme.DefaultFontSans {
		return g.theme.FontSans
	}
	return fonts.Family(metrics.InputGroup.FontWeight)
}

// Layout sizes the container (h-9, requested/available width) and positions the
// addons + the borderless field inside.
func (g *InputGroupWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	w := g.width
	if w <= 0 {
		w = c.MaxWidth
		if w <= 0 || w >= geometry.Infinity {
			w = 240
		}
	}
	h := metrics.InputGroup.Height
	size := c.Constrain(geometry.Sz(w, h))
	bounds := geometry.FromPointSize(g.Position(), size)
	g.SetBounds(bounds)

	pad := metrics.InputGroup.PadX
	gap := metrics.InputGroup.Gap
	innerLeft := bounds.Min.X + pad
	innerRight := bounds.Max.X - pad

	if lw := g.addonWidth(g.leading, ctx); lw > 0 {
		g.placeAddon(g.leading, geometry.NewRect(innerLeft, bounds.Min.Y, lw, h), h)
		innerLeft += lw + gap
	}
	if tw := g.addonWidth(g.trailing, ctx); tw > 0 {
		g.placeAddon(g.trailing, geometry.NewRect(innerRight-tw, bounds.Min.Y, tw, h), h)
		innerRight -= tw + gap
	}

	fieldW := innerRight - innerLeft
	if fieldW < 0 {
		fieldW = 0
	}
	g.field.Layout(ctx, geometry.Tight(geometry.Sz(fieldW, h)))
	g.field.SetBounds(geometry.NewRect(innerLeft, bounds.Min.Y, fieldW, h))
	return size
}

// placeAddon centers an addon (button only needs bounds; icon/text are drawn
// directly) vertically within the container.
func (g *InputGroupWidget) placeAddon(a *InputGroupAddonWidget, rect geometry.Rect, h float32) {
	if a.kind == addonButton && a.button != nil {
		bh := a.button.Bounds().Height()
		y := rect.Min.Y + (h-bh)/2
		a.button.SetBounds(geometry.NewRect(rect.Min.X, y, a.button.Bounds().Width(), bh))
	}
}

// Draw paints the container chrome, then the addons and inner field.
func (g *InputGroupWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !g.IsVisible() {
		return
	}
	th := g.theme
	tok := th.Active()
	dark := th.IsDark()
	bounds := g.Bounds()
	radius := th.RadiusMD()
	disabled := g.disabled

	if !disabled {
		draw.Shadow(canvas, bounds, radius, metrics.ShadowXS)
	}
	if dark {
		fill := draw.MulAlpha(tok.Input, metrics.InputGroup.DarkFillAlpha)
		canvas.DrawRoundRect(bounds, draw.Fade(fill, disabled), radius)
	}

	borderColor := tok.Input
	switch {
	case g.invalid:
		borderColor = tok.Destructive
		a := metrics.InvalidRingAlphaLight
		if dark {
			a = metrics.InvalidRingAlphaDark
		}
		draw.FocusRing(canvas, bounds, radius, draw.Alpha(tok.Destructive, a))
	case g.field.IsFocused() && !disabled:
		borderColor = tok.Ring
		draw.FocusRing(canvas, bounds, radius, draw.Alpha(tok.Ring, metrics.RingAlpha))
	}
	draw.InsideBorder(canvas, bounds, radius, draw.Fade(borderColor, disabled), metrics.InputGroup.BorderWidth)

	// Inner field.
	widget.StampScreenOrigin(g.field, canvas)
	widget.DrawChild(g.field, ctx, canvas)

	// Addons.
	g.drawAddon(ctx, canvas, g.leading, tok, disabled)
	g.drawAddon(ctx, canvas, g.trailing, tok, disabled)
}

// drawAddon paints one addon at its laid-out bounds.
func (g *InputGroupWidget) drawAddon(ctx widget.Context, canvas widget.Canvas, a *InputGroupAddonWidget, tok *theme.Tokens, disabled bool) {
	if a == nil {
		return
	}
	switch a.kind {
	case addonIcon:
		ir := g.addonRect(a)
		size := metrics.InputGroup.IconSize
		iconRect := geometry.NewRect(ir.Min.X, ir.Min.Y+(ir.Height()-size)/2, size, size)
		icon.Draw(canvas, a.icon, iconRect, draw.Fade(tok.MutedForeground, disabled))
	case addonText:
		ir := g.addonRect(a)
		drawTextareaText(canvas, a.text, ir, g.fontFamily(), metrics.InputGroup.FontSize, draw.Fade(tok.MutedForeground, disabled))
	case addonButton:
		if a.button != nil {
			widget.StampScreenOrigin(a.button, canvas)
			widget.DrawChild(a.button, ctx, canvas)
		}
	}
}

// addonRect recomputes an addon's drawing rect (icon/text addons are not
// widgets with bounds; the layout positions are derived the same way as in
// Layout).
func (g *InputGroupWidget) addonRect(a *InputGroupAddonWidget) geometry.Rect {
	bounds := g.Bounds()
	pad := metrics.InputGroup.PadX
	h := bounds.Height()
	if a == g.leading {
		w := g.staticAddonWidth(a)
		return geometry.NewRect(bounds.Min.X+pad, bounds.Min.Y, w, h)
	}
	w := g.staticAddonWidth(a)
	return geometry.NewRect(bounds.Max.X-pad-w, bounds.Min.Y, w, h)
}

// staticAddonWidth returns the width of an icon/text addon (no layout context
// needed).
func (g *InputGroupWidget) staticAddonWidth(a *InputGroupAddonWidget) float32 {
	switch a.kind {
	case addonIcon:
		return metrics.InputGroup.IconSize
	case addonText:
		return textmetrics.Width(g.fontFamily(), metrics.InputGroup.FontSize, a.text)
	}
	return 0
}

// Event forwards mouse/keyboard input to the inner field and addon button.
func (g *InputGroupWidget) Event(ctx widget.Context, e event.Event) bool {
	if g.disabled {
		return false
	}
	// Button addons consume their own clicks when the cursor is over them.
	// MouseLeave is forwarded unconditionally (to clear hover), but we don't
	// let a button consume it — the field also needs to see it.
	for _, a := range []*InputGroupAddonWidget{g.leading, g.trailing} {
		if a != nil && a.kind == addonButton && a.button != nil {
			if me, ok := e.(*event.MouseEvent); ok {
				if me.MouseType == event.MouseLeave {
					a.button.Event(ctx, e)
				} else if a.button.Bounds().Contains(me.Position) {
					if a.button.Event(ctx, e) {
						return true
					}
				}
			}
		}
	}
	return g.field.Event(ctx, e)
}

// IsFocusable reports whether the group's field can take focus.
func (g *InputGroupWidget) IsFocusable() bool {
	return g.IsVisible() && !g.disabled && g.field.IsFocusable()
}

// SetFocused forwards focus to the inner field.
func (g *InputGroupWidget) SetFocused(focused bool) {
	g.field.SetFocused(focused)
	g.WidgetBase.SetFocused(focused)
	g.MarkRedrawLocal() // not SetNeedsRedraw: avoids context-lock re-entry in RequestFocus
}

// Mount mounts the inner field.
func (g *InputGroupWidget) Mount(ctx widget.Context) {
	g.field.Mount(ctx)
}

// Unmount unmounts the inner field.
func (g *InputGroupWidget) Unmount() {
	g.field.Unmount()
}

// Children returns the inner field (and button addons).
func (g *InputGroupWidget) Children() []widget.Widget {
	out := []widget.Widget{g.field}
	for _, a := range []*InputGroupAddonWidget{g.leading, g.trailing} {
		if a != nil && a.kind == addonButton && a.button != nil {
			out = append(out, a.button)
		}
	}
	return out
}

// inputGroupFieldPainter draws the borderless inner text field of an
// InputGroup: text, placeholder, selection, and caret only — the container
// supplies the border, background, and ring.
type inputGroupFieldPainter struct {
	theme    *theme.Theme
	disabled bool
}

// PaintTextField renders only the inner content (no border/background/shadow).
func (p inputGroupFieldPainter) PaintTextField(canvas widget.Canvas, st textfield.PaintState) {
	if st.Bounds.IsEmpty() {
		return
	}
	tok := p.theme.Active()
	disabled := p.disabled || st.Disabled
	family := fonts.Family(metrics.InputGroup.FontWeight)
	size := metrics.InputGroup.FontSize

	content := st.Bounds
	canvas.PushClip(content)
	defer canvas.PopClip()

	if st.Text == "" {
		if st.Placeholder != "" {
			inputGroupDrawText(canvas, st.Placeholder, content, family, size, draw.Fade(tok.MutedForeground, disabled))
		}
		inputGroupCaret(canvas, st, content, "", family, size, tok, disabled)
		return
	}

	if st.SelectStart != st.SelectEnd && !disabled {
		runes := []rune(st.Text)
		s, e := st.SelectStart, st.SelectEnd
		if s > e {
			s, e = e, s
		}
		if s < 0 {
			s = 0
		}
		if e > len(runes) {
			e = len(runes)
		}
		x1 := content.Min.X + inputGroupMeasure(canvas, string(runes[:s]), family, size)
		x2 := content.Min.X + inputGroupMeasure(canvas, string(runes[:e]), family, size)
		if x2 > content.Max.X {
			x2 = content.Max.X
		}
		canvas.DrawRect(geometry.NewRect(x1, content.Min.Y, x2-x1, content.Height()), tok.Primary)
	}

	inputGroupDrawText(canvas, st.Text, content, family, size, draw.Fade(tok.Foreground, disabled))
	inputGroupCaret(canvas, st, content, st.Text, family, size, tok, disabled)
}

// inputGroupDrawText draws vertically-centered left-aligned text.
func inputGroupDrawText(canvas widget.Canvas, text string, bounds geometry.Rect, family string, size float32, col widget.Color) {
	if text == "" {
		return
	}
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(text, bounds, widget.TextStyle{FontFamily: family, FontSize: size, Color: col, Align: widget.TextAlignLeft})
		return
	}
	canvas.DrawText(text, bounds, size, col, false, widget.TextAlignLeft)
}

func inputGroupMeasure(canvas widget.Canvas, text, family string, size float32) float32 {
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		return std.MeasureStyledText(text, widget.TextStyle{FontFamily: family, FontSize: size})
	}
	return canvas.MeasureText(text, size, false)
}

func inputGroupCaret(canvas widget.Canvas, st textfield.PaintState, content geometry.Rect, display, family string, size float32, tok *theme.Tokens, disabled bool) {
	if !st.Focused || disabled || st.SelectStart != st.SelectEnd {
		return
	}
	runes := []rune(display)
	pos := st.CursorPos
	if pos < 0 {
		pos = 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}
	x := content.Min.X + inputGroupMeasure(canvas, string(runes[:pos]), family, size)
	if x > content.Max.X {
		x = content.Max.X
	}
	canvas.DrawRect(geometry.NewRect(x, content.Min.Y, metrics.InputGroup.CaretWidth, content.Height()), tok.Foreground)
}

// Compile-time interface checks.
var (
	_ widget.Widget     = (*InputGroupWidget)(nil)
	_ widget.Focusable  = (*InputGroupWidget)(nil)
	_ widget.Lifecycle  = (*InputGroupWidget)(nil)
	_ textfield.Painter = inputGroupFieldPainter{}
)
