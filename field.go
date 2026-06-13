package graft

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/fonts"
	"github.com/TimLai666/graft/internal/textmetrics"
	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// The Field family is shadcn's form-layout primitive set (field.tsx):
// composition over Box stacks plus typography. A Field stacks a label, a
// control, an optional description, and an optional error with gap-3; a
// FieldGroup stacks fields with gap-7; a FieldSet groups fields under a
// FieldLegend with gap-6. None of these wrap a gogpu/ui core widget — they
// are graft-owned composites following the Card pattern, so colors that
// matter (description muted, error destructive) resolve from the active
// token set inside Typography.Draw and survive light/dark switches.

// FieldWidget is a single shadcn Field: a vertical (default) or horizontal
// stack of its label, control, description, and error with gap-3.
type FieldWidget struct {
	*primitives.BoxWidget
	horizontal bool
}

// Field stacks the given parts (FieldLabel, a control, FieldDescription,
// FieldError, ...) vertically with the field gap (gap-3 = 12).
//
// nil children are skipped so a FieldError() that resolved to nothing can be
// passed unconditionally.
func Field(children ...Widget) *FieldWidget {
	kids := compactWidgets(children)
	f := &FieldWidget{
		BoxWidget: primitives.VBox(kids...).
			Gap(metrics.Field.Gap).
			CrossAlign(primitives.CrossAxisStretch),
	}
	return f
}

// Horizontal lays the field out as a row with centered items
// (orientation="horizontal": flex-row items-center). Use it for a control +
// label pair such as a checkbox with its label.
func (f *FieldWidget) Horizontal() *FieldWidget {
	f.horizontal = true
	f.SetDirection(primitives.DirectionHorizontal)
	f.CrossAlign(primitives.CrossAxisCenter)
	return f
}

// FieldContentWidget is shadcn's FieldContent: a flex-1 column with gap-1.5
// used to stack a title/label and description next to a control in a
// horizontal field.
type FieldContentWidget struct {
	*primitives.BoxWidget
}

// FieldContent stacks the given parts vertically with gap-1.5 (6px).
func FieldContent(children ...Widget) *FieldContentWidget {
	return &FieldContentWidget{
		primitives.VBox(compactWidgets(children)...).
			Gap(metrics.Field.ContentGap).
			CrossAlign(primitives.CrossAxisStretch),
	}
}

// FieldGroupWidget is shadcn's FieldGroup: a full-width column stacking
// fields with gap-7 (28px).
type FieldGroupWidget struct {
	*primitives.BoxWidget
}

// FieldGroup stacks the given fields vertically with the group gap
// (gap-7 = 28).
func FieldGroup(children ...Widget) *FieldGroupWidget {
	return &FieldGroupWidget{
		primitives.VBox(compactWidgets(children)...).
			Gap(metrics.Field.GroupGap).
			CrossAlign(primitives.CrossAxisStretch),
	}
}

// FieldSetWidget is shadcn's FieldSet: a column grouping a legend and its
// fields with gap-6 (24px).
type FieldSetWidget struct {
	*primitives.BoxWidget
}

// FieldSet groups a FieldLegend and fields vertically with gap-6 (24px). The
// legend already carries its own mb-3 spacing.
func FieldSet(children ...Widget) *FieldSetWidget {
	return &FieldSetWidget{
		primitives.VBox(compactWidgets(children)...).
			Gap(metrics.Field.SetGap).
			CrossAlign(primitives.CrossAxisStretch),
	}
}

// FieldLabel creates a field label: 14px / weight 500 / leading-snug in the
// foreground color (Label + leading-snug). It reuses the shadcn Label look.
func FieldLabel(text string) *LabelWidget {
	l := Label(text)
	l.LineHeight(metrics.Field.LabelLineHeight)
	return l
}

// FieldTitle creates a field title: 14px / weight 500 / leading-snug. Unlike
// FieldLabel it is a plain styled div (no label semantics) used as the
// heading inside a FieldContent column.
func FieldTitle(text string) *TypographyWidget {
	return styled(text, metrics.Field.LabelFontSize, metrics.Field.LabelFontWeight, metrics.Field.LabelLineHeight)
}

// FieldDescription creates a field description: 14px / weight 400 /
// leading-normal in the muted-foreground token.
func FieldDescription(text string) *TypographyWidget {
	return styled(text, metrics.Field.DescriptionFontSize, metrics.Field.DescriptionFontWeight, metrics.Field.DescriptionLineHeight).
		Muted()
}

// FieldLegendWidget is shadcn's FieldLegend: a weight-500 heading for a
// FieldSet carrying its mb-3 spacing below the text.
type FieldLegendWidget struct {
	*primitives.BoxWidget
	text *TypographyWidget
}

// FieldLegend creates a field-set legend: weight 500, 16px ("legend"
// variant) with mb-3 spacing below it.
func FieldLegend(text string) *FieldLegendWidget {
	return newLegend(text, metrics.Field.LegendFontSize)
}

// FieldLegendLabel creates a field-set legend in the smaller "label" variant
// (14px) with mb-3 spacing below.
func FieldLegendLabel(text string) *FieldLegendWidget {
	return newLegend(text, metrics.Field.LegendLabelFontSize)
}

func newLegend(text string, size float32) *FieldLegendWidget {
	t := styled(text, size, metrics.Field.LegendFontWeight, size)
	return &FieldLegendWidget{
		BoxWidget: primitives.VBox(t).PaddingBottom(metrics.Field.LegendMarginBottom),
		text:      t,
	}
}

// Theme pins a specific theme.
func (l *FieldLegendWidget) Theme(th *theme.Theme) *FieldLegendWidget {
	l.text.Theme(th)
	return l
}

// FieldErrorWidget is shadcn's FieldError: 14px destructive text shown only
// when an error message is present. It can be driven by a read-only signal
// (Form wires this) so it appears/disappears as validation state changes.
type FieldErrorWidget struct {
	*TypographyWidget
	sig state.ReadonlySignal[string]
}

// FieldError creates a field error message: 14px / weight 400 in the
// destructive token. An empty message renders nothing (the row collapses).
func FieldError(message string) *FieldErrorWidget {
	t := styled(message, metrics.Field.ErrorFontSize, metrics.Field.ErrorFontWeight, metrics.Field.ErrorLineHeight).
		Destructive()
	e := &FieldErrorWidget{TypographyWidget: t}
	e.syncVisible(message)
	return e
}

// BindError drives the error text from a read-only signal. When the message
// is empty the row is hidden; the binding is registered in Mount. Form's
// FormField uses this to surface validation errors.
func (e *FieldErrorWidget) BindError(sig state.ReadonlySignal[string]) *FieldErrorWidget {
	e.sig = sig
	e.SetContent(sig.Get())
	e.syncVisible(sig.Get())
	return e
}

// Theme pins a specific theme.
func (e *FieldErrorWidget) Theme(th *theme.Theme) *FieldErrorWidget {
	e.TypographyWidget.Theme(th)
	return e
}

func (e *FieldErrorWidget) syncVisible(msg string) {
	e.SetVisible(msg != "")
}

// Layout collapses the error to zero size when there is no message.
func (e *FieldErrorWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	if e.sig != nil {
		msg := e.sig.Get()
		if msg != e.Content() {
			e.SetContent(msg)
		}
		e.syncVisible(msg)
	}
	if e.Content() == "" {
		e.SetBounds(geometry.FromPointSize(e.Position(), geometry.Sz(0, 0)))
		return geometry.Sz(0, 0)
	}
	return e.TypographyWidget.Layout(ctx, c)
}

// Mount binds the error signal so external validation repaints the row.
func (e *FieldErrorWidget) Mount(ctx widget.Context) {
	if sched := ctx.Scheduler(); sched != nil && e.sig != nil {
		e.AddBinding(state.BindToScheduler(e.sig, e, sched))
	}
}

// FieldSeparatorWidget is shadcn's FieldSeparator: a 20px row with a 1px
// rule across its vertical center and optional centered text that masks the
// rule with the page background (the "or" divider).
type FieldSeparatorWidget struct {
	widget.WidgetBase

	label string
	theme *theme.Theme
}

// FieldSeparator creates a plain horizontal divider sized as a 20px row
// (h-5 -my-2) with the rule at its vertical center.
func FieldSeparator() *FieldSeparatorWidget {
	s := &FieldSeparatorWidget{theme: CurrentTheme()}
	s.SetVisible(true)
	s.SetEnabled(true)
	return s
}

// Label adds centered text (e.g. "Or continue with") that masks the rule
// with the background color, matching shadcn's data-content separator.
func (s *FieldSeparatorWidget) Label(text string) *FieldSeparatorWidget {
	s.label = text
	return s
}

// Theme pins a specific theme.
func (s *FieldSeparatorWidget) Theme(th *theme.Theme) *FieldSeparatorWidget {
	s.theme = th
	return s
}

func (s *FieldSeparatorWidget) resolvedTheme() *theme.Theme {
	if s.theme != nil {
		return s.theme
	}
	return CurrentTheme()
}

// Layout sizes the separator as a full-width 20px row.
func (s *FieldSeparatorWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	var w float32
	if c.HasBoundedWidth() {
		w = c.MaxWidth
	}
	size := c.Constrain(geometry.Sz(w, metrics.Field.SeparatorHeight))
	s.SetBounds(geometry.FromPointSize(s.Position(), size))
	return size
}

// Draw renders the centered rule and, when present, the masking label text.
func (s *FieldSeparatorWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !s.IsVisible() {
		return
	}
	th := s.resolvedTheme()
	tok := th.Active()
	bounds := s.Bounds()
	midY := bounds.Center().Y

	// 1px rule across the vertical center (Separator at top-1/2).
	rule := geometry.NewRect(bounds.Min.X, midY-metrics.Field.SeparatorThickness/2, bounds.Width(), metrics.Field.SeparatorThickness)
	canvas.DrawRect(rule, tok.Border)

	if s.label == "" {
		return
	}

	// Centered label masking the rule with the background, in muted text.
	family := fonts.Family(400)
	size := metrics.Field.ErrorFontSize
	tw := textmetrics.Width(family, size, s.label)
	boxW := tw + 2*metrics.Field.SeparatorTextGap
	boxX := bounds.Min.X + (bounds.Width()-boxW)/2
	mask := geometry.NewRect(boxX, bounds.Min.Y, boxW, bounds.Height())
	canvas.DrawRect(mask, tok.Background)

	textRect := geometry.NewRect(boxX+metrics.Field.SeparatorTextGap, midY-size/2, tw, size)
	if std, ok := canvas.(widget.StyledTextDrawer); ok {
		std.DrawStyledText(s.label, textRect, widget.TextStyle{
			FontFamily: family,
			FontSize:   size,
			Color:      tok.MutedForeground,
			Align:      widget.TextAlignLeft,
		})
		return
	}
	canvas.DrawText(s.label, textRect, size, tok.MutedForeground, false, widget.TextAlignLeft)
}

// Event ignores all input; separators are inert.
func (s *FieldSeparatorWidget) Event(widget.Context, event.Event) bool { return false }

// Children returns nil; the separator is a leaf.
func (s *FieldSeparatorWidget) Children() []widget.Widget { return nil }

// compactWidgets drops nil entries so callers can pass optional parts
// unconditionally (e.g. a FieldError that resolved to nothing).
func compactWidgets(in []Widget) []Widget {
	out := make([]Widget, 0, len(in))
	for _, w := range in {
		if w == nil {
			continue
		}
		out = append(out, w)
	}
	return out
}

// Compile-time interface checks.
var (
	_ widget.Widget = (*FieldWidget)(nil)
	_ widget.Widget = (*FieldContentWidget)(nil)
	_ widget.Widget = (*FieldGroupWidget)(nil)
	_ widget.Widget = (*FieldSetWidget)(nil)
	_ widget.Widget = (*FieldLegendWidget)(nil)
	_ widget.Widget = (*FieldErrorWidget)(nil)
	_ widget.Widget = (*FieldSeparatorWidget)(nil)
)
