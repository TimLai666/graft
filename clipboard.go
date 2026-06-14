package graft

import "sync"

// Clipboard provider.
//
// graft's core module is deliberately free of any GPU/windowing dependency (it
// never imports gogpu/gogpu), so it cannot reach the OS clipboard on its own.
// Widgets that need copy/cut/paste therefore go through a package-level provider
// that defaults to a safe in-memory string. This keeps the core importable for
// headless and offscreen use (and tests) while letting a host wire the real OS
// clipboard in.
//
// graftapp installs an OS-backed provider (the gogpu App's ClipboardRead /
// ClipboardWrite) in its launcher before the render loop starts. When no
// provider is installed the in-memory default is used, so behavior is identical
// in tests and headless runs.
//
// LIMITATION: this provider is only consulted by graft-OWNED widgets (currently
// Textarea). Input / InputGroup wrap gogpu's core/textfield, which uses its own
// internal clipboard that is not reachable without forking that core widget, so
// they are unaffected by SetClipboard.

var (
	clipboardMu sync.RWMutex

	// inMemoryClipboard backs the default provider.
	inMemoryClipboard string

	// clipboardRead / clipboardWrite are the active provider. They default to
	// the in-memory string and are replaced by SetClipboard.
	clipboardRead  = defaultClipboardRead
	clipboardWrite = defaultClipboardWrite
)

func defaultClipboardRead() string {
	clipboardMu.RLock()
	defer clipboardMu.RUnlock()
	return inMemoryClipboard
}

func defaultClipboardWrite(s string) {
	clipboardMu.Lock()
	defer clipboardMu.Unlock()
	inMemoryClipboard = s
}

// SetClipboard installs an OS-backed (or otherwise custom) clipboard provider
// for graft-owned widgets such as Textarea. read returns the current clipboard
// text; write replaces it. Passing nil for either argument restores that side's
// in-memory default, so a partial install never panics.
//
// The core package defaults to an in-memory clipboard so it works headless and
// in tests; graftapp installs the OS clipboard in its launcher.
func SetClipboard(read func() string, write func(string)) {
	clipboardMu.Lock()
	defer clipboardMu.Unlock()
	if read != nil {
		clipboardRead = read
	} else {
		clipboardRead = defaultClipboardRead
	}
	if write != nil {
		clipboardWrite = write
	} else {
		clipboardWrite = defaultClipboardWrite
	}
}

// clipboardGet returns the current clipboard text via the active provider.
func clipboardGet() string {
	clipboardMu.RLock()
	read := clipboardRead
	clipboardMu.RUnlock()
	return read()
}

// clipboardSet writes text to the clipboard via the active provider.
func clipboardSet(s string) {
	clipboardMu.RLock()
	write := clipboardWrite
	clipboardMu.RUnlock()
	write(s)
}
