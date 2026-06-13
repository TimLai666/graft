package graft

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
)

// minLenMessage formats the MinLen error message.
func minLenMessage(n int) string {
	return "Must be at least " + strconv.Itoa(n) + " characters."
}

// graft's Form is the Go-idiomatic replacement for react-hook-form
// (DESIGN.md §4). Instead of a controller hook it is a tiny signal-backed
// state container: each field carries a value signal, a validator list, and
// an error signal. Submitting (or calling Validate) runs every field's
// validators and writes their error signals, which Field/FieldError and the
// control's Invalid binding observe — so the UI updates reactively without
// any imperative wiring. The whole package is pure logic plus composition
// over the Field family; it renders nothing of its own.

// Validator reports a validation error for a field value, or nil when the
// value is acceptable. Validators run in registration order and the first
// non-nil error wins.
type Validator func(value string) error

// validationError is the error type returned by the built-in validators so
// the message can be surfaced verbatim in a FieldError.
type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }

// newValidationError builds a Validator error with the given message.
func newValidationError(msg string) error { return &validationError{msg: msg} }

// Required fails when the value is empty or only whitespace.
func Required() Validator {
	return func(v string) error {
		if strings.TrimSpace(v) == "" {
			return newValidationError("This field is required.")
		}
		return nil
	}
}

// MinLen fails when the value has fewer than n runes (after no trimming;
// length is measured on the raw value, matching HTML minlength).
func MinLen(n int) Validator {
	return func(v string) error {
		if len([]rune(v)) < n {
			return newValidationError(minLenMessage(n))
		}
		return nil
	}
}

// emailPattern is a pragmatic email check: non-empty local part, an @, a
// dot-separated domain. It intentionally mirrors the loose validation web
// forms use rather than RFC 5322 in full.
var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// Email fails when the value is non-empty and not a valid-looking email
// address. An empty value passes (combine with Required to forbid empty).
func Email() Validator {
	return func(v string) error {
		if v == "" {
			return nil
		}
		if !emailPattern.MatchString(v) {
			return newValidationError("Enter a valid email address.")
		}
		return nil
	}
}

// Custom adapts an arbitrary func(string) error into a Validator.
func Custom(fn func(string) error) Validator { return Validator(fn) }

// FormValueState is a single form field's reactive state: its value signal,
// its error-message signal, a derived invalid flag, and the validators that
// connect them. Controls Bind to Value(); Field/FieldError observe Error()
// and Invalid().
type FormValueState struct {
	name       string
	value      state.Signal[string]
	errMsg     state.Signal[string]
	invalid    state.ReadonlySignal[bool]
	validators []Validator
}

// FormState holds a form's registered fields and submit handler. Construct
// it with NewForm and add fields with FormValue.
type FormState struct {
	fields   []*FormValueState
	byName   map[string]*FormValueState
	onSubmit func()
}

// NewForm creates an empty form-state container.
func NewForm() *FormState {
	return &FormState{byName: map[string]*FormValueState{}}
}

// FormValue registers a field on the form and returns its reactive state. The
// value starts at initial; validators run on Validate/Submit in order.
func FormValue(form *FormState, name, initial string, validators ...Validator) *FormValueState {
	errSig := state.NewSignal("")
	fv := &FormValueState{
		name:       name,
		value:      state.NewSignal(initial),
		errMsg:     errSig,
		validators: validators,
	}
	// invalid is derived: a non-empty error message means the field is invalid.
	fv.invalid = state.NewComputed(func() bool {
		return errSig.Get() != ""
	}, errSig.AsReadonly())

	if form != nil {
		form.fields = append(form.fields, fv)
		if form.byName != nil {
			form.byName[name] = fv
		}
	}
	return fv
}

// Name returns the field's registration name.
func (fv *FormValueState) Name() string { return fv.name }

// Value returns the writable value signal; bind a control to it.
func (fv *FormValueState) Value() state.Signal[string] { return fv.value }

// Get returns the current value.
func (fv *FormValueState) Get() string { return fv.value.Get() }

// Set writes a new value.
func (fv *FormValueState) Set(v string) { fv.value.Set(v) }

// Error returns the read-only error-message signal (empty when valid).
func (fv *FormValueState) Error() state.ReadonlySignal[string] { return fv.errMsg.AsReadonly() }

// Invalid returns a read-only flag that is true while the field has an error;
// wire it into a control's BindInvalid.
func (fv *FormValueState) Invalid() state.ReadonlySignal[bool] { return fv.invalid }

// Validate runs the field's validators against its current value, writes the
// resulting message into the error signal, and reports whether it passed.
func (fv *FormValueState) Validate() bool {
	v := fv.value.Get()
	for _, vd := range fv.validators {
		if vd == nil {
			continue
		}
		if err := vd(v); err != nil {
			fv.errMsg.Set(err.Error())
			return false
		}
	}
	fv.errMsg.Set("")
	return true
}

// ClearError resets the field's error message.
func (fv *FormValueState) ClearError() { fv.errMsg.Set("") }

// Field looks up a previously registered field by name (nil if absent).
func (f *FormState) Field(name string) *FormValueState {
	if f == nil || f.byName == nil {
		return nil
	}
	return f.byName[name]
}

// Values returns a name→value snapshot of every registered field.
func (f *FormState) Values() map[string]string {
	out := make(map[string]string, len(f.fields))
	for _, fv := range f.fields {
		out[fv.name] = fv.value.Get()
	}
	return out
}

// Validate runs every field's validators (all of them, so each error signal
// is populated) and reports whether the whole form is valid.
func (f *FormState) Validate() bool {
	ok := true
	for _, fv := range f.fields {
		if !fv.Validate() {
			ok = false
		}
	}
	return ok
}

// OnSubmit registers the handler run by Submit when the form validates.
func (f *FormState) OnSubmit(fn func()) { f.onSubmit = fn }

// Submit validates the form and, when valid, invokes the submit handler.
// Returns whether the form was valid.
func (f *FormState) Submit() bool {
	if !f.Validate() {
		return false
	}
	if f.onSubmit != nil {
		f.onSubmit()
	}
	return true
}

// FormWidget is the visual root of a form: a vertical stack of FormFields and
// actions. It is a thin composite over a Box; submission logic lives on the
// FormState it carries.
type FormWidget struct {
	*primitives.BoxWidget
	form *FormState
}

// Form lays the given children (FormFields, buttons, ...) out vertically with
// the default form gap and associates them with the form state.
func Form(form *FormState, children ...Widget) *FormWidget {
	return &FormWidget{
		BoxWidget: primitives.VBox(compactWidgets(children)...).
			Gap(metrics.FormGap).
			CrossAlign(primitives.CrossAxisStretch),
		form: form,
	}
}

// Gap overrides the vertical spacing between form children.
func (f *FormWidget) Gap(px float32) *FormWidget {
	f.BoxWidget.Gap(px)
	return f
}

// OnSubmit registers the form's submit handler (sugar for
// form.OnSubmit(fn)).
func (f *FormWidget) OnSubmit(fn func()) *FormWidget {
	if f.form != nil {
		f.form.OnSubmit(fn)
	}
	return f
}

// State returns the underlying form state.
func (f *FormWidget) State() *FormState { return f.form }

// invalidBinder is implemented by controls that can surface a form field's
// invalid flag (Input, and any future graft control with aria-invalid).
type invalidBinder interface {
	BindInvalid(sig state.ReadonlySignal[bool]) *InputWidget
}

// FormField wires a form field's reactive state to a Field row: it binds the
// control to the field value (and its invalid flag when the control supports
// it), labels it, and appends a FieldError bound to the field's error signal.
//
// The control is bound to the value only if it is a graft control that
// exposes Bind/BindInvalid; otherwise it is laid out as-is (callers can wire
// the binding themselves). Optional extras (e.g. a FieldDescription) are
// inserted between the control and the error.
func FormField(fv *FormValueState, label string, control Widget, extras ...Widget) *FieldWidget {
	if in, ok := control.(*InputWidget); ok {
		in.Bind(fv.Value())
		in.BindInvalid(fv.Invalid())
	} else if b, ok := control.(invalidBinder); ok {
		b.BindInvalid(fv.Invalid())
	}

	parts := make([]Widget, 0, len(extras)+3)
	if label != "" {
		parts = append(parts, FieldLabel(label))
	}
	parts = append(parts, control)
	parts = append(parts, extras...)
	parts = append(parts, FieldError("").BindError(fv.Error()))
	return Field(parts...)
}

// SubmitButton builds a primary button that submits the form on click.
// button.go is not modified; this composes the existing Button API. Pass
// extra children (e.g. an icon) just like Button.
func SubmitButton(form *FormState, label string, children ...Widget) *ButtonWidget {
	return Button(label, children...).OnClick(func() {
		if form != nil {
			form.Submit()
		}
	})
}

// Compile-time interface checks.
var (
	_ widget.Widget = (*FormWidget)(nil)
)
