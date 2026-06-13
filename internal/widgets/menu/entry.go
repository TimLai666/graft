// Package menu is the shared shadcn menu engine: an overlay panel widget
// whose rows are arbitrary menu entries (item, checkbox-item, radio-item,
// label, separator) following the shadcn dropdown-menu anatomy. It is
// reusable by DropdownMenu, ContextMenu, and Menubar drop-downs.
//
// The engine is graft-OWNED (painter-less): it draws directly via
// internal/draw + metrics + theme tokens, exactly like the other Class-C
// widgets. It exposes a clean, declarative API — build entries, hand them
// to NewPanel, and the panel renders rows and handles keyboard navigation
// (up/down/home/end skip disabled, enter selects, esc closes).
//
// Icons are passed in as icon.IconData so the engine does not depend on
// graft's icons package; callers (dropdownmenu.go) supply the lucide check,
// circle, and per-item icons.
//
// Sub-menus are deferred for v1.
package menu

import "github.com/gogpu/ui/icon"

// Kind classifies a menu entry row.
type Kind uint8

// Entry kinds.
const (
	// KindItem is a selectable action row.
	KindItem Kind = iota
	// KindCheckbox is a toggleable checkbox row (check indicator at left).
	KindCheckbox
	// KindRadio is a radio row (dot indicator at left).
	KindRadio
	// KindLabel is a non-interactive section label row.
	KindLabel
	// KindSeparator is a divider row.
	KindSeparator
)

// Entry is one row in a menu panel. Concrete entries are built with NewItem,
// NewCheckboxItem, NewRadioItem, NewLabel, and NewSeparator.
type Entry interface {
	// kind returns the entry kind.
	kind() Kind
	// selectable reports whether keyboard navigation may land on this row
	// (an enabled item/checkbox/radio).
	selectable() bool
}

// ItemEntry is a selectable action row.
type ItemEntry struct {
	Label       string
	IconData    icon.IconData
	HasIcon     bool
	Shortcut    string
	Destructive bool
	IsDisabled  bool
	IsInset     bool
	OnSelectFn  func()
}

// NewItem creates a selectable action item.
func NewItem(label string) *ItemEntry { return &ItemEntry{Label: label} }

// Icon sets the leading icon (16px, muted-foreground).
func (e *ItemEntry) Icon(ic icon.IconData) *ItemEntry {
	e.IconData = ic
	e.HasIcon = true
	return e
}

// SetShortcut sets the right-aligned shortcut text (12px muted).
func (e *ItemEntry) SetShortcut(s string) *ItemEntry { e.Shortcut = s; return e }

// SetDestructive marks the item as destructive (text-destructive + tinted
// focus fill).
func (e *ItemEntry) SetDestructive() *ItemEntry { e.Destructive = true; return e }

// SetDisabled marks the item disabled (faded, not navigable).
func (e *ItemEntry) SetDisabled(v bool) *ItemEntry { e.IsDisabled = v; return e }

// SetInset shifts the label right (pl-8) to align with items that have
// indicators.
func (e *ItemEntry) SetInset() *ItemEntry { e.IsInset = true; return e }

// OnSelect registers the selection callback.
func (e *ItemEntry) OnSelect(fn func()) *ItemEntry { e.OnSelectFn = fn; return e }

func (e *ItemEntry) kind() Kind       { return KindItem }
func (e *ItemEntry) selectable() bool { return !e.IsDisabled }

// CheckboxEntry is a toggleable checkbox row.
type CheckboxEntry struct {
	Label      string
	Checked    bool
	IsDisabled bool
	CheckIcon  icon.IconData
	HasCheck   bool
	OnChangeFn func(bool)
}

// NewCheckboxItem creates a checkbox row. The check indicator icon is
// supplied so the engine stays icon-package-independent.
func NewCheckboxItem(label string, checkIcon icon.IconData) *CheckboxEntry {
	return &CheckboxEntry{Label: label, CheckIcon: checkIcon, HasCheck: true}
}

// SetChecked sets the checked state.
func (e *CheckboxEntry) SetChecked(v bool) *CheckboxEntry { e.Checked = v; return e }

// SetDisabled marks the checkbox disabled.
func (e *CheckboxEntry) SetDisabled(v bool) *CheckboxEntry { e.IsDisabled = v; return e }

// OnChange registers the toggle callback.
func (e *CheckboxEntry) OnChange(fn func(bool)) *CheckboxEntry { e.OnChangeFn = fn; return e }

func (e *CheckboxEntry) kind() Kind       { return KindCheckbox }
func (e *CheckboxEntry) selectable() bool { return !e.IsDisabled }

// RadioEntry is a single radio row within a radio group.
type RadioEntry struct {
	Value      string
	Label      string
	Selected   bool
	IsDisabled bool
	DotIcon    icon.IconData
	HasDot     bool
	OnSelectFn func(value string)
}

// NewRadioItem creates a radio row. The dot indicator icon is supplied so
// the engine stays icon-package-independent.
func NewRadioItem(value, label string, dotIcon icon.IconData) *RadioEntry {
	return &RadioEntry{Value: value, Label: label, DotIcon: dotIcon, HasDot: true}
}

// SetSelected sets the selected state.
func (e *RadioEntry) SetSelected(v bool) *RadioEntry { e.Selected = v; return e }

// SetDisabled marks the radio disabled.
func (e *RadioEntry) SetDisabled(v bool) *RadioEntry { e.IsDisabled = v; return e }

// OnSelect registers the selection callback (receives the entry value).
func (e *RadioEntry) OnSelect(fn func(value string)) *RadioEntry { e.OnSelectFn = fn; return e }

func (e *RadioEntry) kind() Kind       { return KindRadio }
func (e *RadioEntry) selectable() bool { return !e.IsDisabled }

// LabelEntry is a non-interactive section label row.
type LabelEntry struct {
	Text    string
	IsInset bool
}

// NewLabel creates a section label row.
func NewLabel(text string) *LabelEntry { return &LabelEntry{Text: text} }

// SetInset shifts the label right (pl-8).
func (e *LabelEntry) SetInset() *LabelEntry { e.IsInset = true; return e }

func (e *LabelEntry) kind() Kind       { return KindLabel }
func (e *LabelEntry) selectable() bool { return false }

// SeparatorEntry is a divider row.
type SeparatorEntry struct{}

// NewSeparator creates a divider row.
func NewSeparator() *SeparatorEntry { return &SeparatorEntry{} }

func (e *SeparatorEntry) kind() Kind       { return KindSeparator }
func (e *SeparatorEntry) selectable() bool { return false }
