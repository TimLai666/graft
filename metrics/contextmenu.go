package metrics

// ContextMenu shares the shadcn menu family metrics with DropdownMenu and the
// Menubar drop-downs — see metrics.Menu for the transcribed Tailwind classes
// (panel min-w-[8rem], p-1, rounded-md, 1px border, shadow-md; items px-2
// py-1.5 rounded-sm text-sm, etc.). shadcn's context-menu.tsx uses the same
// class strings as dropdown-menu.tsx.
//
// The only ContextMenu-specific behaviour is placement: the panel opens AT THE
// CURSOR on right-click (event.MousePress with Button==ButtonRight), anchored
// to the pointer's GlobalPosition rather than below a trigger. The constants
// here document that anchoring; row/surface metrics route through metrics.Menu.
const (
	// ContextMenuCursorGap is the gap in px between the cursor and the panel's
	// top-left corner when the menu opens at the pointer. shadcn anchors the
	// panel essentially at the cursor (Radix places it with a small offset); 2px
	// keeps the first row from sitting directly under the pointer.
	ContextMenuCursorGap float32 = 2
)
