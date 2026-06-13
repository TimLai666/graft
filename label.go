package graft

import (
	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/widget"

	"github.com/TimLai666/graft/metrics"
	"github.com/TimLai666/graft/theme"
)

// LabelWidget is the shadcn Label: 14px / weight 500 / leading-none text
// in the foreground color.
//
// Architecture decision: graft-owned thin wrapper around TypographyWidget
// (the component is pure styled text); it keeps its own type and file per
// the one-file-per-shadcn-registry-entry rule.
type LabelWidget struct {
	*TypographyWidget

	forTarget Widget
}

// Label creates a shadcn form label (text-sm leading-none font-medium).
func Label(text string) *LabelWidget {
	return &LabelWidget{
		// lineHeight 0 = leading-none (line box equals font size).
		TypographyWidget: styled(text, metrics.Label.FontSize, metrics.Label.FontWeight, 0),
	}
}

// For associates the label with a form control. v1 stores the target for
// future accessibility wiring only (no visual effect).
func (l *LabelWidget) For(w Widget) *LabelWidget {
	l.forTarget = w
	return l
}

// Disabled renders the label at 50% opacity
// (group-data-[disabled=true]:opacity-50).
func (l *LabelWidget) Disabled(v bool) *LabelWidget {
	if v {
		l.Opacity(metrics.DisabledOpacity)
	} else {
		l.Opacity(1)
	}
	return l
}

// Theme pins a specific theme instead of the process-wide current theme.
func (l *LabelWidget) Theme(th *theme.Theme) *LabelWidget {
	l.TypographyWidget.Theme(th)
	return l
}

// AccessibilityRole returns the label role.
func (l *LabelWidget) AccessibilityRole() a11y.Role { return a11y.RoleLabel }

// AccessibilityLabel returns the label text.
func (l *LabelWidget) AccessibilityLabel() string { return l.Content() }

// AccessibilityHint returns no hint.
func (l *LabelWidget) AccessibilityHint() string { return "" }

// AccessibilityValue returns no value.
func (l *LabelWidget) AccessibilityValue() string { return "" }

// AccessibilityState reports an empty state; labels are inert.
func (l *LabelWidget) AccessibilityState() a11y.State { return a11y.State{} }

// AccessibilityActions returns no actions.
func (l *LabelWidget) AccessibilityActions() []a11y.Action { return nil }

// Compile-time interface checks.
var (
	_ widget.Widget   = (*LabelWidget)(nil)
	_ a11y.Accessible = (*LabelWidget)(nil)
)
