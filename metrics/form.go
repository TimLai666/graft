package metrics

// Form metrics. The graft Form is the Go-idiomatic replacement for
// react-hook-form (DESIGN.md §4): it is pure validation logic plus layout
// composition over the Field family, so it owns no novel pixel values. The
// one layout constant is the default vertical gap between a form's children,
// which matches a FieldGroup (gap-7 = 28) so stacked FormFields breathe the
// same way the shadcn form examples do.
const (
	// FormGap is the default vertical gap between form children in px,
	// matching FieldGroup (gap-7 = 28).
	FormGap float32 = 28
)
