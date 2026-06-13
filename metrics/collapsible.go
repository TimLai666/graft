package metrics

// Collapsible holds the pixel/animation constants for the shadcn Collapsible
// component (docs/research/03-shadcn-pixel-spec.md; Collapsible itself ships
// unstyled in shadcn — it is a Radix primitive wrapper with no chrome of its
// own. The trigger and content are user-supplied widgets; graft only owns the
// open/closed show-hide of the content and the collapse animation timing.)
//
// shadcn anatomy: Collapsible (root) / CollapsibleTrigger (asChild button) /
// CollapsibleContent (the region that expands and collapses). The content
// region animates its height between 0 and its natural height; shadcn uses the
// Radix `--radix-collapsible-content-height` CSS var with the same
// collapsible-down/up keyframes Accordion uses (0.2s ease-out).
var Collapsible = struct {
	// DurationMS is the content height expand/collapse animation duration in
	// milliseconds (collapsible-down/up = 0.2s ease-out, shared with
	// Accordion). Goldens render the settled state; this records the spec.
	DurationMS float32

	// ContentGap is the default vertical gap inserted between the trigger and
	// the content in graft's owned layout. shadcn leaves spacing to the user
	// (the component has no padding); graft keeps 0 so the content sits
	// directly under the trigger unless the caller wraps it.
	ContentGap float32
}{
	DurationMS: 200, // 0.2s ease-out
	ContentGap: 0,   // unstyled root; no built-in spacing
}
