package graft

import (
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// InputWidget is the shadcn Input: a single-line text field rendered in the
// shadcn style (docs/research/03-shadcn-pixel-spec.md §5 "Input").
//
// WRAP DECISION (DESIGN.md §3.2 / §4): Input wraps the gogpu/ui core
// textfield widget because text editing (cursor, selection, insertion,
// password masking, two-way signal sync, validation, Enter submit) is
// substantial machinery. The shadcn visuals come from painters.TextField
// (painters/textfield.go). The one thing the core widget cannot do is the
// 36px height: its Layout hardcodes 48px and offers no PaddingXY layout hook
// (Report 1 §7.6 trick is unavailable here), so InputWidget overrides Layout
// to pin h-9 = 36px while delegating Draw/Event/focus/lifecycle to the inner
// widget. The aria-invalid state is driven through the field's validation
// channel (a sentinel validator gated by the invalid flag) so the painter's
// PaintState.HasError lights up without a per-widget painter field.
//
// The inner field is rebuilt as fluent setters are chained, because the core
// textfield config is write-once at New. Rebuilds are cheap (no GPU state)
// and only happen during the builder chain, before the widget mounts.
type InputWidget struct {
	*textfield.Widget

	theme *theme.Theme

	// userOpts are the caller-supplied options accumulated by the fluent
	// setters, replayed (after the base painter + invalid validator) on every
	// rebuild.
	userOpts []textfield.Option

	value    string // initial text, preserved across rebuilds
	hasValue bool

	// invalid is the static aria-invalid flag; invalidSig (when set) overrides
	// it. Both feed the sentinel validator so the inner field's HasError, and
	// thus the painter's invalid styling, tracks the graft Invalid API.
	invalid    bool
	invalidSig state.ReadonlySignal[bool]

	width    float32 // explicit width in px (.W); 0 = fill available
	disabled bool
}

// Input creates an empty shadcn text input snapshotting the current theme.
func Input() *InputWidget {
	i := &InputWidget{theme: CurrentTheme()}
	i.rebuild()
	return i
}

// Placeholder sets the empty-state placeholder text.
func (i *InputWidget) Placeholder(s string) *InputWidget {
	return i.addOption(textfield.Placeholder(s))
}

// Value sets the initial (uncontrolled) text value.
func (i *InputWidget) Value(s string) *InputWidget {
	i.value = s
	i.hasValue = true
	i.rebuild()
	return i
}

// Bind makes the input controlled by sig: it renders sig's value and writes
// edits back to it. The binding is registered in Mount.
func (i *InputWidget) Bind(sig state.Signal[string]) *InputWidget {
	return i.addOption(textfield.ValueSignal(sig))
}

// Password masks typed characters with bullets.
func (i *InputWidget) Password() *InputWidget {
	return i.addOption(textfield.InputTypeOpt(textfield.TypePassword))
}

// Disabled sets the disabled state (faded, not focusable, ignores input).
func (i *InputWidget) Disabled(v bool) *InputWidget {
	i.disabled = v
	return i.addOption(textfield.Disabled(v))
}

// Invalid sets the aria-invalid state (destructive border + ring).
func (i *InputWidget) Invalid(v bool) *InputWidget {
	i.invalid = v
	i.revalidate()
	return i
}

// BindInvalid drives the aria-invalid state from a read-only signal (Form
// wires this). The binding is registered in Mount.
func (i *InputWidget) BindInvalid(sig state.ReadonlySignal[bool]) *InputWidget {
	i.invalidSig = sig
	i.revalidate()
	return i
}

// OnChange registers a callback fired on every text change.
func (i *InputWidget) OnChange(fn func(string)) *InputWidget {
	return i.addOption(textfield.OnChange(fn))
}

// OnSubmit registers a callback fired when the user presses Enter.
func (i *InputWidget) OnSubmit(fn func(string)) *InputWidget {
	return i.addOption(textfield.OnSubmit(fn))
}

// W sets an explicit width in px (otherwise the input fills available width).
func (i *InputWidget) W(px float32) *InputWidget {
	i.width = px
	return i
}

// Theme pins a specific theme instead of the process-wide current theme.
func (i *InputWidget) Theme(th *theme.Theme) *InputWidget {
	i.theme = th
	i.rebuild()
	return i
}

// addOption appends a caller option and rebuilds the inner field.
func (i *InputWidget) addOption(opt textfield.Option) *InputWidget {
	i.userOpts = append(i.userOpts, opt)
	i.rebuild()
	return i
}

// resolveInvalid returns the effective aria-invalid state (signal wins).
func (i *InputWidget) resolveInvalid() bool {
	if i.invalidSig != nil {
		return i.invalidSig.Get()
	}
	return i.invalid
}

// revalidate re-runs the inner field's validation so the sentinel invalid
// validator updates HasError immediately after the flag changes.
func (i *InputWidget) revalidate() {
	if i.Widget != nil {
		i.Widget.SetText(i.Widget.Text())
	}
}

// rebuild replaces the inner field with the base options (graft painter +
// invalid validator) followed by every accumulated caller option, then
// restores any uncontrolled initial value.
func (i *InputWidget) rebuild() {
	opts := make([]textfield.Option, 0, len(i.userOpts)+2)
	opts = append(opts,
		textfield.PainterOpt(PaintersFor(i.theme).TextField),
		textfield.Validation(func(string) string {
			if i.resolveInvalid() {
				return "invalid"
			}
			return ""
		}),
	)
	opts = append(opts, i.userOpts...)
	i.Widget = textfield.New(opts...)
	if i.hasValue {
		i.Widget.SetText(i.value)
	}
}

// Layout pins the control to h-9 (36px) and the requested or available width,
// overriding the inner widget's fixed 48px. The inner widget's bounds are set
// to match so its Draw and hit-testing align.
func (i *InputWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	w := i.width
	if w <= 0 {
		w = c.MaxWidth
		if w <= 0 || w >= geometry.Infinity {
			w = 200 // sensible default when offered unbounded width
		}
	}
	size := c.Constrain(geometry.Sz(w, metrics.Input.Height))
	bounds := geometry.FromPointSize(i.Position(), size)
	i.SetBounds(bounds)
	return size
}

// Mount binds the value (via the inner field) and invalid signals.
func (i *InputWidget) Mount(ctx widget.Context) {
	i.Widget.Mount(ctx)
	if sched := ctx.Scheduler(); sched != nil && i.invalidSig != nil {
		i.AddBinding(state.BindToScheduler(i.invalidSig, i, sched))
	}
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*InputWidget)(nil)
	_ widget.Focusable = (*InputWidget)(nil)
	_ widget.Lifecycle = (*InputWidget)(nil)
)
