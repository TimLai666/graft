package graft_test

import (
	"testing"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft"
	"github.com/TimLai666/graft/internal/gtest"
	"github.com/TimLai666/graft/metrics"
)

// TestFieldDescriptionColor pins the description to muted-foreground at 14px.
func TestFieldDescriptionColor(t *testing.T) {
	tok := lightTokens(t)
	d := graft.FieldDescription("Enter your email below")
	uitest.LayoutWidget(d, 400, 200)
	c := uitest.DrawWidget(d)
	if len(c.StyledTexts) != 1 {
		t.Fatalf("styled texts = %d, want 1", len(c.StyledTexts))
	}
	if c.StyledTexts[0].Style.Color != tok.MutedForeground {
		t.Errorf("description color = %v, want muted-foreground %v", c.StyledTexts[0].Style.Color, tok.MutedForeground)
	}
	if c.StyledTexts[0].Style.FontSize != metrics.Field.DescriptionFontSize {
		t.Errorf("description size = %v, want %v", c.StyledTexts[0].Style.FontSize, metrics.Field.DescriptionFontSize)
	}
}

// TestFieldErrorColor pins a non-empty error to the destructive token.
func TestFieldErrorColor(t *testing.T) {
	tok := lightTokens(t)
	e := graft.FieldError("Email is required")
	uitest.LayoutWidget(e, 400, 200)
	c := uitest.DrawWidget(e)
	if len(c.StyledTexts) != 1 {
		t.Fatalf("styled texts = %d, want 1", len(c.StyledTexts))
	}
	if c.StyledTexts[0].Style.Color != tok.Destructive {
		t.Errorf("error color = %v, want destructive %v", c.StyledTexts[0].Style.Color, tok.Destructive)
	}
}

// TestFieldErrorEmptyCollapses verifies an empty error draws nothing and
// reports zero size.
func TestFieldErrorEmptyCollapses(t *testing.T) {
	lightTokens(t)
	e := graft.FieldError("")
	size := uitest.LayoutWidget(e, 400, 200)
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("empty error size = %v, want 0×0", size)
	}
	c := uitest.DrawWidget(e)
	if len(c.StyledTexts)+len(c.Texts) != 0 {
		t.Errorf("empty error drew text: %d styled, %d plain", len(c.StyledTexts), len(c.Texts))
	}
}

// TestFieldErrorBindShows verifies a bound error appears when the signal
// receives a message and hides when cleared.
func TestFieldErrorBindShows(t *testing.T) {
	lightTokens(t)
	sig := state.NewSignal("")
	e := graft.FieldError("").BindError(sig)

	size := uitest.LayoutWidget(e, 400, 200)
	if size.Height != 0 {
		t.Fatalf("error should be collapsed while empty, got %v", size)
	}

	sig.Set("Required")
	size = uitest.LayoutWidget(e, 400, 200)
	if size.Height == 0 {
		t.Fatalf("error should be visible after setting a message")
	}
	c := uitest.DrawWidget(e)
	if len(c.StyledTexts) != 1 {
		t.Fatalf("styled texts after set = %d, want 1", len(c.StyledTexts))
	}

	sig.Set("")
	size = uitest.LayoutWidget(e, 400, 200)
	if size.Height != 0 {
		t.Fatalf("error should re-collapse after clearing, got %v", size)
	}
}

// TestFieldLabelLeadingSnug pins the field label line box to leading-snug.
func TestFieldLabelLeadingSnug(t *testing.T) {
	lightTokens(t)
	l := graft.FieldLabel("Email")
	size := uitest.LayoutWidget(l, 400, 200)
	if size.Height != metrics.Field.LabelLineHeight {
		t.Errorf("label height = %v, want leading-snug %v", size.Height, metrics.Field.LabelLineHeight)
	}
}

// TestFieldStackGap pins the vertical field gap (gap-3 = 12) between the
// label and the control by checking the laid-out child positions.
func TestFieldStackGap(t *testing.T) {
	lightTokens(t)
	label := graft.FieldLabel("Email")
	control := graft.Input()
	f := graft.Field(label, control)
	uitest.LayoutWidget(f, 400, 400)

	gap := control.Bounds().Min.Y - label.Bounds().Max.Y
	if gap != metrics.Field.Gap {
		t.Errorf("field gap = %v, want %v", gap, metrics.Field.Gap)
	}
}

// TestFieldNilChildSkipped verifies a nil child is dropped so optional parts
// can be passed unconditionally.
func TestFieldNilChildSkipped(t *testing.T) {
	lightTokens(t)
	f := graft.Field(graft.FieldLabel("Email"), nil, graft.Input())
	if got := len(f.Children()); got != 2 {
		t.Errorf("children = %d, want 2 (nil dropped)", got)
	}
}

// TestGoldenField renders the Field family in light and dark.
func TestGoldenField(t *testing.T) {
	gtest.GoldenLightDark(t, "field-basic", func() widget.Widget {
		return primitives.VBox(
			graft.Field(
				graft.FieldLabel("Email"),
				graft.Input().Placeholder("you@example.com"),
				graft.FieldDescription("We will never share your email."),
			),
		).Padding(24).Width(360)
	})

	gtest.GoldenLightDark(t, "field-error", func() widget.Widget {
		return primitives.VBox(
			graft.Field(
				graft.FieldLabel("Password"),
				graft.Input().Password().Invalid(true),
				graft.FieldError("Password is too short."),
			),
		).Padding(24).Width(360)
	})

	gtest.GoldenLightDark(t, "field-set", func() widget.Widget {
		return primitives.VBox(
			graft.FieldSet(
				graft.FieldLegend("Address"),
				graft.Field(
					graft.FieldLabel("Street"),
					graft.Input(),
				),
				graft.Field(
					graft.FieldLabel("City"),
					graft.Input(),
				),
			),
		).Padding(24).Width(360)
	})

	gtest.GoldenLightDark(t, "field-separator", func() widget.Widget {
		return primitives.VBox(
			graft.FieldSeparator().Label("Or continue with"),
		).Padding(24).Width(360)
	})
}
