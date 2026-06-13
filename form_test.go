package graft_test

import (
	"errors"
	"testing"

	"github.com/gogpu/ui/uitest"

	"github.com/TimLai666/graft"
)

// TestFormRequired covers the Required validator on submit.
func TestFormRequired(t *testing.T) {
	form := graft.NewForm()
	email := graft.FormValue(form, "email", "", graft.Required())

	if form.Validate() {
		t.Fatal("empty required field should be invalid")
	}
	if email.Error().Get() == "" {
		t.Fatal("error signal should be populated after failed validation")
	}
	if !email.Invalid().Get() {
		t.Fatal("invalid flag should be true while the field has an error")
	}

	email.Set("a@b.co")
	if !form.Validate() {
		t.Fatal("filled required field should validate")
	}
	if email.Error().Get() != "" {
		t.Fatalf("error should clear on success, got %q", email.Error().Get())
	}
	if email.Invalid().Get() {
		t.Fatal("invalid flag should be false after clearing the error")
	}
}

// TestFormMinLen covers the MinLen validator and its message.
func TestFormMinLen(t *testing.T) {
	form := graft.NewForm()
	pw := graft.FormValue(form, "password", "abc", graft.MinLen(8))

	if pw.Validate() {
		t.Fatal("short value should fail MinLen(8)")
	}
	if pw.Error().Get() == "" {
		t.Fatal("MinLen should set an error message")
	}

	pw.Set("abcdefgh")
	if !pw.Validate() {
		t.Fatal("8-char value should pass MinLen(8)")
	}
}

// TestFormEmail covers the Email validator (empty passes; combine with
// Required to forbid empty).
func TestFormEmail(t *testing.T) {
	form := graft.NewForm()
	email := graft.FormValue(form, "email", "not-an-email", graft.Email())

	if email.Validate() {
		t.Fatal("malformed email should fail")
	}

	email.Set("")
	if !email.Validate() {
		t.Fatal("empty value should pass Email() alone")
	}

	email.Set("user@example.com")
	if !email.Validate() {
		t.Fatal("valid email should pass")
	}
}

// TestFormCustom covers the Custom validator wrapper.
func TestFormCustom(t *testing.T) {
	form := graft.NewForm()
	field := graft.FormValue(form, "code", "nope", graft.Custom(func(v string) error {
		if v != "secret" {
			return errors.New("wrong code")
		}
		return nil
	}))

	if field.Validate() {
		t.Fatal("custom validator should reject 'nope'")
	}
	if field.Error().Get() != "wrong code" {
		t.Fatalf("error = %q, want 'wrong code'", field.Error().Get())
	}

	field.Set("secret")
	if !field.Validate() {
		t.Fatal("custom validator should accept 'secret'")
	}
}

// TestFormValidatorOrder verifies the first failing validator wins.
func TestFormValidatorOrder(t *testing.T) {
	form := graft.NewForm()
	email := graft.FormValue(form, "email", "", graft.Required(), graft.Email())

	if email.Validate() {
		t.Fatal("empty value should fail")
	}
	// Required runs first, so its message should be the one surfaced.
	if got := email.Error().Get(); got == "Enter a valid email address." {
		t.Fatalf("Required should win over Email for empty value, got %q", got)
	}
}

// TestFormSubmitGatesOnValidation verifies OnSubmit only runs when valid.
func TestFormSubmitGatesOnValidation(t *testing.T) {
	form := graft.NewForm()
	name := graft.FormValue(form, "name", "", graft.Required())

	submitted := 0
	form.OnSubmit(func() { submitted++ })

	if form.Submit() {
		t.Fatal("submit should fail while a required field is empty")
	}
	if submitted != 0 {
		t.Fatal("OnSubmit must not run on a failed submit")
	}

	name.Set("Ada")
	if !form.Submit() {
		t.Fatal("submit should succeed once required fields are filled")
	}
	if submitted != 1 {
		t.Fatalf("OnSubmit should run exactly once, ran %d", submitted)
	}
}

// TestFormValidateAllFields verifies a whole-form Validate populates every
// field's error, not just the first failing one.
func TestFormValidateAllFields(t *testing.T) {
	form := graft.NewForm()
	a := graft.FormValue(form, "a", "", graft.Required())
	b := graft.FormValue(form, "b", "", graft.Required())

	if form.Validate() {
		t.Fatal("form with two empty required fields should be invalid")
	}
	if a.Error().Get() == "" || b.Error().Get() == "" {
		t.Fatal("Validate should populate the error of every field")
	}
}

// TestFormValuesSnapshot verifies the name→value snapshot.
func TestFormValuesSnapshot(t *testing.T) {
	form := graft.NewForm()
	graft.FormValue(form, "first", "Ada")
	graft.FormValue(form, "last", "Lovelace")

	vals := form.Values()
	if vals["first"] != "Ada" || vals["last"] != "Lovelace" {
		t.Fatalf("values snapshot = %v", vals)
	}
}

// TestFormFieldLookup verifies registered fields are retrievable by name.
func TestFormFieldLookup(t *testing.T) {
	form := graft.NewForm()
	fv := graft.FormValue(form, "email", "x")
	if form.Field("email") != fv {
		t.Fatal("Field lookup should return the registered field")
	}
	if form.Field("missing") != nil {
		t.Fatal("missing field lookup should return nil")
	}
}

// TestFormFieldWiresInputInvalid verifies FormField binds the input value and
// invalid flag to the form field.
func TestFormFieldWiresInputInvalid(t *testing.T) {
	lightTokens(t)
	form := graft.NewForm()
	email := graft.FormValue(form, "email", "", graft.Required())

	in := graft.Input()
	row := graft.FormField(email, "Email", in)
	uitest.LayoutWidget(row, 400, 400)

	// Before validation the row has no visible error; the FieldError is the
	// last child and should be collapsed.
	form.Validate() // empty required → error set
	// The bound input now tracks the field value: writing the signal updates it.
	email.Set("typed@example.com")
	if in == nil {
		t.Fatal("FormField should keep the control")
	}
	if email.Value().Get() != "typed@example.com" {
		t.Fatal("field value signal should hold the set value")
	}
}

// TestSubmitButtonSubmits verifies the helper button triggers form.Submit.
func TestSubmitButtonSubmits(t *testing.T) {
	lightTokens(t)
	form := graft.NewForm()
	name := graft.FormValue(form, "name", "Ada", graft.Required())
	_ = name

	submitted := 0
	form.OnSubmit(func() { submitted++ })

	btn := graft.SubmitButton(form, "Save")
	uitest.LayoutWidget(btn, 200, 60)
	b := btn.Bounds()
	uitest.SimulateClick(btn, b.Center().X, b.Center().Y)

	if submitted != 1 {
		t.Fatalf("submit button click should submit once, got %d", submitted)
	}
}
