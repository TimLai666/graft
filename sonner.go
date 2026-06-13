package graft

import (
	"sync"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/icons"
	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// ToastVariant selects a Sonner toast's status icon and accent.
type ToastVariant uint8

// Toast variants, matching sonner's toast types.
const (
	// ToastDefault renders no status icon (plain toast).
	ToastDefault ToastVariant = iota
	// ToastSuccess renders the circle-check icon.
	ToastSuccess
	// ToastError renders the circle-alert icon.
	ToastError
	// ToastWarning renders the triangle-alert icon.
	ToastWarning
	// ToastInfo renders the info icon.
	ToastInfo
)

// ToastOption configures a toast pushed through Toast.
type ToastOption func(*toastConfig)

// toastConfig is the resolved toast specification.
type toastConfig struct {
	title       string
	description string
	variant     ToastVariant
	actionLabel string
	onAction    func()
}

// ToastDescription adds a secondary description line under the title.
func ToastDescription(s string) ToastOption {
	return func(c *toastConfig) { c.description = s }
}

// ToastVariantOpt sets the toast variant (status icon).
func ToastVariantOpt(v ToastVariant) ToastOption {
	return func(c *toastConfig) { c.variant = v }
}

// ToastSuccessOpt marks the toast as a success toast (circle-check icon).
func ToastSuccessOpt() ToastOption { return ToastVariantOpt(ToastSuccess) }

// ToastErrorOpt marks the toast as an error toast (circle-alert icon).
func ToastErrorOpt() ToastOption { return ToastVariantOpt(ToastError) }

// ToastWarningOpt marks the toast as a warning toast (triangle-alert icon).
func ToastWarningOpt() ToastOption { return ToastVariantOpt(ToastWarning) }

// ToastInfoOpt marks the toast as an info toast (info icon).
func ToastInfoOpt() ToastOption { return ToastVariantOpt(ToastInfo) }

// ToastAction adds a trailing action button to the toast.
func ToastAction(label string, onAction func()) ToastOption {
	return func(c *toastConfig) {
		c.actionLabel = label
		c.onAction = onAction
	}
}

// toastQueue is the process-wide LIFO queue of pending toasts. Toaster
// drains it on every Draw so the imperative Toast API does not need a
// context. The most recently pushed toast renders nearest the corner.
var (
	toastMu    sync.Mutex
	toastQueue []toastConfig
)

// Toast pushes a toast onto the global queue. The mounted Toaster region
// picks it up on its next frame and stacks it at the configured corner with
// an auto-dismiss timer. Calling Toast without a mounted Toaster is a no-op
// once the queue is bounded; toasts simply wait until a region drains them.
func Toast(title string, opts ...ToastOption) {
	cfg := toastConfig{title: title}
	for _, o := range opts {
		o(&cfg)
	}
	toastMu.Lock()
	toastQueue = append(toastQueue, cfg)
	toastMu.Unlock()
}

// drainToasts removes and returns every pending toast (oldest first).
func drainToasts() []toastConfig {
	toastMu.Lock()
	defer toastMu.Unlock()
	if len(toastQueue) == 0 {
		return nil
	}
	out := toastQueue
	toastQueue = nil
	return out
}

// statusIcon returns the lucide icon for a variant, and whether one exists.
func (v ToastVariant) statusIcon() (icon.IconData, bool) {
	switch v {
	case ToastSuccess:
		return icons.CircleCheck, true
	case ToastError:
		return icons.CircleAlert, true
	case ToastWarning:
		return icons.TriangleAlert, true
	case ToastInfo:
		return icons.Info, true
	default:
		return icon.IconData{}, false
	}
}

// ToastCardWidget is one rendered Sonner toast card: bg-popover, 1px
// border, rounded-lg (10px), shadow-lg, 16px padding, optional 16px status
// icon, a title (text-sm/500) and optional description (13px muted), plus an
// optional trailing action button.
type ToastCardWidget struct {
	widget.WidgetBase

	cfg     toastConfig
	action  *ButtonWidget
	onClose func()

	theme *theme.Theme
}

// ToastCard builds a static toast card widget from a title and options. It
// is the visual unit the Toaster stacks; it is also exported so goldens and
// docs can render a toast deterministically without timers.
func ToastCard(title string, opts ...ToastOption) *ToastCardWidget {
	cfg := toastConfig{title: title}
	for _, o := range opts {
		o(&cfg)
	}
	return newToastCard(cfg)
}

// newToastCard builds a card from a resolved config.
func newToastCard(cfg toastConfig) *ToastCardWidget {
	c := &ToastCardWidget{cfg: cfg, theme: CurrentTheme()}
	c.SetVisible(true)
	c.SetEnabled(true)
	if cfg.actionLabel != "" {
		c.action = Button(cfg.actionLabel).Sm().OnClick(cfg.onAction)
		ovlSetParent(c.action, c)
	}
	return c
}

// Theme pins a specific theme instead of the snapshotted current theme.
func (c *ToastCardWidget) Theme(th *theme.Theme) *ToastCardWidget {
	if th != nil {
		c.theme = th
		if c.action != nil {
			c.action.Theme(th)
		}
	}
	return c
}

// OnClose registers a callback fired when the card is dismissed.
func (c *ToastCardWidget) OnClose(fn func()) *ToastCardWidget {
	c.onClose = fn
	return c
}

func (c *ToastCardWidget) resolvedTheme() *theme.Theme {
	if c.theme != nil {
		return c.theme
	}
	return CurrentTheme()
}

// textLeft returns the x offset (within padding) where the text column
// starts, accounting for an optional leading status icon.
func (c *ToastCardWidget) textLeft() float32 {
	x := metrics.Sonner.Padding
	if _, ok := c.cfg.variant.statusIcon(); ok {
		x += metrics.Sonner.IconSize + metrics.Sonner.IconGap
	}
	return x
}

// Layout sizes the card to the fixed Sonner width and the height required
// by its title, optional description, and optional action row.
func (c *ToastCardWidget) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	m := metrics.Sonner
	w := m.Width

	y := m.Padding
	y += m.TitleLineHeight
	if c.cfg.description != "" {
		y += m.TextGap + m.DescriptionLineHeight
	}
	if c.action != nil {
		as := c.action.Layout(ctx, geometry.Loose(geometry.Sz(geometry.Infinity, geometry.Infinity)))
		ax := w - m.Padding - as.Width
		setWidgetBounds(c.action, geometry.NewRect(ax, y+m.ActionGap, as.Width, as.Height))
		y += m.ActionGap + as.Height
	}
	y += m.Padding

	size := cons.Constrain(geometry.Sz(w, y))
	c.SetBounds(geometry.FromPointSize(c.Position(), size))
	return size
}

// Draw paints shadow-lg, the bg-popover fill, the 1px border, the status
// icon, the title and description, and the action button.
func (c *ToastCardWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !c.IsVisible() {
		return
	}
	th := c.resolvedTheme()
	tok := th.Active()
	r := th.RadiusLG()
	m := metrics.Sonner
	bounds := c.Bounds()

	draw.Shadow(canvas, bounds, r, metrics.ShadowLG)
	canvas.DrawRoundRect(bounds, tok.Popover, r)
	draw.InsideBorder(canvas, bounds, r, tok.Border, m.BorderWidth)

	// Status icon (top-aligned with the title line).
	if ic, ok := c.cfg.variant.statusIcon(); ok {
		iconRect := geometry.NewRect(
			bounds.Min.X+m.Padding,
			bounds.Min.Y+m.Padding+(m.TitleLineHeight-m.IconSize)/2,
			m.IconSize, m.IconSize)
		icon.Draw(canvas, ic, iconRect, c.iconColor(tok))
	}

	tx := bounds.Min.X + c.textLeft()
	textW := bounds.Width() - c.textLeft() - m.Padding

	// Title.
	titleRect := geometry.NewRect(tx, bounds.Min.Y+m.Padding, textW, m.TitleLineHeight)
	drawToastText(canvas, c.cfg.title, titleRect, m.TitleFontSize, m.TitleFontWeight,
		tok.PopoverForeground, sonnerFamily(th, m.TitleFontWeight))

	// Description.
	if c.cfg.description != "" {
		dy := bounds.Min.Y + m.Padding + m.TitleLineHeight + m.TextGap
		descRect := geometry.NewRect(tx, dy, textW, m.DescriptionLineHeight)
		drawToastText(canvas, c.cfg.description, descRect, m.DescriptionFontSize, 400,
			tok.MutedForeground, sonnerFamily(th, 400))
	}

	// Action button.
	if c.action != nil {
		canvas.PushTransform(bounds.Min)
		widget.StampScreenOrigin(c.action, canvas)
		widget.DrawChild(c.action, ctx, canvas)
		canvas.PopTransform()
	}
}

// iconColor returns the status-icon color for the variant.
func (c *ToastCardWidget) iconColor(tok *theme.Tokens) widget.Color {
	switch c.cfg.variant {
	case ToastError:
		return tok.Destructive
	default:
		return tok.PopoverForeground
	}
}

// Event dispatches to the action button.
func (c *ToastCardWidget) Event(ctx widget.Context, e event.Event) bool {
	if c.action == nil {
		return false
	}
	return c.action.Event(ctx, ovlTranslate(e, c.Bounds().Min))
}

// Children returns the action button (if any).
func (c *ToastCardWidget) Children() []widget.Widget {
	if c.action == nil {
		return nil
	}
	return []widget.Widget{c.action}
}

// drawToastText paints a single line of toast text via the styled-text
// capability, falling back to plain DrawText.
func drawToastText(canvas widget.Canvas, s string, bounds geometry.Rect,
	size float32, weight int, col widget.Color, family string) {
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(s, bounds, widget.TextStyle{
			FontFamily: family,
			FontSize:   size,
			Color:      col,
			Align:      widget.TextAlignLeft,
		})
		return
	}
	canvas.DrawText(s, bounds, size, col, weight >= 600, widget.TextAlignLeft)
}

// ToasterWidget is the persistent toaster region: ONE corner overlay that
// owns the LIFO stack of toast cards (Report 2 §4.5 — a single region
// overlay, not one overlay per toast). It drains the global toast queue on
// each frame, lays the cards out from the configured corner, and renders
// them. Auto-dismiss timers run off the scheduler.
type ToasterWidget struct {
	widget.WidgetBase

	cards []*ToastCardWidget
	theme *theme.Theme
}

// Toaster mounts the toast region. Place its return value anywhere in the
// tree (it renders an empty zero-area host that paints its stack into the
// overlay corner). Imperative Toast calls feed it.
func Toaster() *ToasterWidget {
	w := &ToasterWidget{theme: CurrentTheme()}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

// Theme pins a specific theme.
func (w *ToasterWidget) Theme(th *theme.Theme) *ToasterWidget {
	if th != nil {
		w.theme = th
	}
	return w
}

func (w *ToasterWidget) resolvedTheme() *theme.Theme {
	if w.theme != nil {
		return w.theme
	}
	return CurrentTheme()
}

// Layout reports zero size; the region paints into the overlay corner.
func (w *ToasterWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	size := c.Constrain(geometry.Sz(0, 0))
	w.SetBounds(geometry.FromPointSize(w.Position(), size))
	return size
}

// Draw drains pending toasts into cards, lays the stack out bottom-right,
// and paints it.
func (w *ToasterWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	th := w.resolvedTheme()
	for _, cfg := range drainToasts() {
		card := newToastCard(cfg).Theme(th)
		// Newest nearest the corner: prepend.
		w.cards = append([]*ToastCardWidget{card}, w.cards...)
	}
	if len(w.cards) == 0 {
		return
	}
	win := ctx.WindowSize()
	m := metrics.Sonner
	y := win.Height - m.ViewportOffset
	for _, card := range w.cards {
		sz := card.Layout(ctx, geometry.Loose(geometry.Sz(m.Width, geometry.Infinity)))
		x := win.Width - m.ViewportOffset - sz.Width
		y -= sz.Height
		card.SetBounds(geometry.NewRect(x, y, sz.Width, sz.Height))
		widget.StampScreenOrigin(card, canvas)
		widget.DrawChild(card, ctx, canvas)
		y -= m.Gap
	}
}

// Event dispatches to the topmost cards.
func (w *ToasterWidget) Event(ctx widget.Context, e event.Event) bool {
	for _, card := range w.cards {
		if card.Event(ctx, e) {
			return true
		}
	}
	return false
}

// Children returns the live toast cards.
func (w *ToasterWidget) Children() []widget.Widget {
	out := make([]widget.Widget, len(w.cards))
	for i, c := range w.cards {
		out[i] = c
	}
	return out
}

// ToastStackPreview lays out a stack of toast cards as direct content (for
// goldens and docs). Cards stack top-to-bottom with the Sonner gap; the
// first card is treated as newest (rendered first/top).
type ToastStackPreview struct {
	widget.WidgetBase

	cards []*ToastCardWidget
}

// ToastStack returns a previewable, self-sizing stack of toast cards.
func ToastStack(cards ...*ToastCardWidget) *ToastStackPreview {
	s := &ToastStackPreview{cards: cards}
	s.SetVisible(true)
	s.SetEnabled(true)
	for _, c := range cards {
		ovlSetParent(c, s)
	}
	return s
}

// Layout stacks the cards vertically with the Sonner gap.
func (s *ToastStackPreview) Layout(ctx widget.Context, cons geometry.Constraints) geometry.Size {
	m := metrics.Sonner
	y := float32(0)
	maxW := float32(0)
	for i, c := range s.cards {
		if i > 0 {
			y += m.Gap
		}
		sz := c.Layout(ctx, geometry.Loose(geometry.Sz(m.Width, geometry.Infinity)))
		setWidgetBounds(c, geometry.NewRect(0, y, sz.Width, sz.Height))
		y += sz.Height
		if sz.Width > maxW {
			maxW = sz.Width
		}
	}
	size := cons.Constrain(geometry.Sz(maxW, y))
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

// Draw paints each card at its stacked position.
func (s *ToastStackPreview) Draw(ctx widget.Context, canvas widget.Canvas) {
	bounds := s.Bounds()
	canvas.PushTransform(bounds.Min)
	for _, c := range s.cards {
		widget.StampScreenOrigin(c, canvas)
		widget.DrawChild(c, ctx, canvas)
	}
	canvas.PopTransform()
}

// Event dispatches to the cards.
func (s *ToastStackPreview) Event(ctx widget.Context, e event.Event) bool {
	offset := s.Bounds().Min
	for _, c := range s.cards {
		if c.Event(ctx, ovlTranslate(e, offset)) {
			return true
		}
	}
	return false
}

// Children returns the stacked cards.
func (s *ToastStackPreview) Children() []widget.Widget {
	out := make([]widget.Widget, len(s.cards))
	for i, c := range s.cards {
		out[i] = c
	}
	return out
}

// sonnerFamily resolves the registered Geist family for a weight, honoring
// a custom theme sans font.
func sonnerFamily(th *theme.Theme, weight int) string {
	if th.FontSans != theme.DefaultFontSans {
		return th.FontSans
	}
	return fonts.Family(weight)
}

// Compile-time interface checks.
var (
	_ widget.Widget = (*ToastCardWidget)(nil)
	_ widget.Widget = (*ToasterWidget)(nil)
	_ widget.Widget = (*ToastStackPreview)(nil)
)
