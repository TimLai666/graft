package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/internal/draw"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// AlertDialog creates a modal alert-dialog host: same chassis as Dialog but
// with NO close button and NO backdrop-click dismissal (Esc still closes), per
// shadcn (DESIGN.md §4). Pair it with a DialogTrigger flipping the same signal.
//
// The content's close button is hidden automatically; an AlertDialog must offer
// explicit Action/Cancel buttons in its footer.
func AlertDialog(content *DialogContentWidget) *DialogWidget {
	content.HideClose()
	d := Dialog(content)
	d.dismissOnBackdrop = false
	return d
}

// AlertDialogContent assembles the alert-dialog card. It mirrors DialogContent
// and always hides the close button.
func AlertDialogContent(children ...widget.Widget) *DialogContentWidget {
	return DialogContent(children...).HideClose()
}

// AlertDialogHeader stacks title + description vertically with gap 8
// (mirrors DialogHeader).
func AlertDialogHeader(children ...widget.Widget) *DialogSectionWidget {
	return DialogHeader(children...)
}

// AlertDialogTitle renders the alert title: 18px / 600 / leading-none
// (mirrors DialogTitle).
func AlertDialogTitle(text string) *TypographyWidget { return DialogTitle(text) }

// AlertDialogDescription renders the alert description: 14px muted
// (mirrors DialogDescription).
func AlertDialogDescription(text string) *TypographyWidget { return DialogDescription(text) }

// AlertDialogFooter lays the Action/Cancel buttons in a right-aligned row with
// gap 8 (mirrors DialogFooter).
func AlertDialogFooter(children ...widget.Widget) *DialogSectionWidget {
	return DialogFooter(children...)
}

// AlertDialogAction is the confirming button — a default (primary) button.
//
// Deviation: the design brief calls for returning a configured graft.Button,
// but the Button widget is not present in this batch's base. To keep batches
// disjoint, AlertDialogAction returns a small owned button widget styled like
// shadcn's default variant; once Button is merged, this can wrap it instead.
func AlertDialogAction(label string, onClick func()) *AlertDialogButtonWidget {
	return newAlertButton(label, true, onClick)
}

// AlertDialogCancel is the dismissing button — an outline button.
//
// Same deviation note as AlertDialogAction.
func AlertDialogCancel(label string, onClick func()) *AlertDialogButtonWidget {
	return newAlertButton(label, false, onClick)
}

// AlertDialogButtonWidget is a minimal shadcn-styled button used by the
// AlertDialog footer (default + outline variants). It exists only because the
// shared Button widget lives in another batch; it follows the same metrics
// (h36, px16, radius MD, font 14/500, shadow-xs on outline) and hover math.
type AlertDialogButtonWidget struct {
	widget.WidgetBase

	label   *TypographyWidget
	primary bool
	onClick func()
	theme   *theme.Theme

	hovered bool
}

func newAlertButton(label string, primary bool, onClick func()) *AlertDialogButtonWidget {
	b := &AlertDialogButtonWidget{
		label:   Text(label).FontSize(14).Weight(500).LineHeight(20).Align(widget.TextAlignCenter),
		primary: primary,
		onClick: onClick,
	}
	b.SetVisible(true)
	b.SetEnabled(true)
	return b
}

// Theme pins a specific theme.
func (b *AlertDialogButtonWidget) Theme(th *theme.Theme) *AlertDialogButtonWidget {
	b.theme = th
	b.label.Theme(th)
	return b
}

func (b *AlertDialogButtonWidget) resolvedTheme() *theme.Theme {
	if b.theme != nil {
		return b.theme
	}
	return CurrentTheme()
}

// alertButtonHeight is shadcn's default button height (h-9 = 36px).
const alertButtonHeight float32 = 36

// alertButtonPadX is the default button horizontal padding (px-4 = 16px).
const alertButtonPadX float32 = 16

// Layout sizes the button: text advance + 2*padX wide, fixed 36px tall.
func (b *AlertDialogButtonWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	ts := b.label.Layout(ctx, geometry.Loose(geometry.Sz(10000, 10000)))
	size := c.Constrain(geometry.Sz(ts.Width+2*alertButtonPadX, alertButtonHeight))
	b.SetBounds(geometry.FromPointSize(b.Position(), size))
	return size
}

// Draw paints the button surface (with hover math) and the label.
func (b *AlertDialogButtonWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !b.IsVisible() {
		return
	}
	th := b.resolvedTheme()
	tok := th.Active()
	bounds := b.Bounds()
	radius := th.RadiusMD()

	if b.primary {
		fill := tok.Primary
		if b.hovered {
			fill = draw.Alpha(tok.Primary, 0.9) // hover:bg-primary/90
		}
		canvas.DrawRoundRect(bounds, fill, radius)
		b.label.Color(tok.PrimaryForeground)
	} else {
		// Outline variant: shadow-xs, bg-background, 1px Border, accent hover.
		draw.Shadow(canvas, bounds, radius, metrics.ShadowXS)
		fill := tok.Background
		if b.hovered {
			fill = tok.Accent // hover:bg-accent
		}
		canvas.DrawRoundRect(bounds, fill, radius)
		draw.InsideBorder(canvas, bounds, radius, tok.Border, 1)
		if b.hovered {
			b.label.Color(tok.AccentForeground)
		} else {
			b.label.Color(tok.Foreground)
		}
	}

	textBounds := geometry.NewRect(
		bounds.Min.X+alertButtonPadX, bounds.Min.Y+(alertButtonHeight-20)/2,
		bounds.Width()-2*alertButtonPadX, 20,
	)
	b.label.Theme(th)
	b.label.SetBounds(textBounds)
	b.label.Draw(ctx, canvas)
}

// Event handles hover and click.
func (b *AlertDialogButtonWidget) Event(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	inside := b.Bounds().Contains(me.Position)
	switch me.MouseType {
	case event.MouseEnter, event.MouseMove:
		if inside != b.hovered {
			b.hovered = inside
			if inside {
				ctx.SetCursor(widget.CursorPointer)
			}
			b.SetNeedsRedraw(true)
			ctx.Invalidate()
		}
	case event.MouseLeave:
		if b.hovered {
			b.hovered = false
			b.SetNeedsRedraw(true)
		}
	case event.MousePress:
		if inside && me.Button == event.ButtonLeft {
			if b.onClick != nil {
				b.onClick()
			}
			return true
		}
	}
	return false
}

// Children returns the label.
func (b *AlertDialogButtonWidget) Children() []widget.Widget { return []widget.Widget{b.label} }

// Compile-time interface check.
var _ widget.Widget = (*AlertDialogButtonWidget)(nil)
