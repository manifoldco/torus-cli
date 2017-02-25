package ui

import (
	"bytes"
	"fmt"
	"os"

	"github.com/chzyer/readline"
	"github.com/manifoldco/ansiwrap"

	"github.com/manifoldco/torus-cli/prefs"
	"github.com/manifoldco/torus-cli/promptui"
)

const (
	defaultCols = 80
	rightPad    = 2
)

var bold = promptui.Styler(promptui.FGBold)
var defUI *UI

// Init initializes a default global UI, accessible via the package functions.
func Init(preferences *prefs.Preferences) {
	defUI = &UI{
		Indent: 0,
		Cols:   screenWidth(),

		EnableProgress: preferences.Core.EnableProgress,
		EnableHints:    preferences.Core.EnableHints,
	}
}

// UI exposes methods for creating a terminal ui
type UI struct {
	Indent int

	// Cols holds the column width for text wrapping. For the default UI and
	// its children, It is either the width of the  terminal, or defaultCols,
	// minus rightPad.
	Cols int

	// EnableProgress is whether progress events should be displayed
	EnableProgress bool

	// EnableHints is whether hints should be displayed
	EnableHints bool
}

// Progress calls Progress on the default UI
func Progress(str string) { defUI.Progress(str) }

// Progress handles the ui output for progress events, when enabled
func (u *UI) Progress(str string) {
	if !u.EnableProgress {
		return
	}
	u.Line(str)
}

// Line calls Line on the default UI
func Line(format string, a ...interface{}) { defUI.Line(format, a...) }

// Line writes a formatted string followed by a newline to stdout. Output is
// word wrapped, and terminated by a newline.
func (u *UI) Line(format string, a ...interface{}) {
	u.LineIndent(0, format, a...)
}

// LineIndent calls LineIndent on the default UI
func LineIndent(indent int, format string, a ...interface{}) { defUI.LineIndent(indent, format, a...) }

// LineIndent writes a formatted string followed by a newline to stdout. Output
// is word wrapped, and terminated by a newline. All lines after the first are
// indented by indent number of spaces (in addition to the indenting enforced
// by this UI instance.
func (u *UI) LineIndent(indent int, format string, a ...interface{}) {
	o := fmt.Sprintf(format, a...)
	fmt.Fprintln(readline.Stdout, ansiwrap.WrapIndent(o, u.Cols, u.Indent, u.Indent+indent))
}

// Hint calls hint on the default UI
func Hint(str string, noPadding bool) { defUI.Hint(str, noPadding) }

// Hint handles the ui output for hint/onboarding messages, when enabled
func (u *UI) Hint(str string, noPadding bool) {
	if !u.EnableHints {
		return
	}
	if !noPadding {
		fmt.Println()
	}

	label := bold("Protip: ")
	rc := ansiwrap.RuneCount(label)
	fmt.Fprintln(readline.Stdout, ansiwrap.WrapIndent(label+str, u.Cols, u.Indent, u.Indent+rc))
}

// Child calls Child on the default UI
func Child(indent int) *UI { return defUI.Child(indent) }

// Child returns a new UI, with settings from the receiver UI, and Indent
// increased by the provided value.
func (u *UI) Child(indent int) *UI {
	return &UI{
		Indent: u.Indent + indent,
		Cols:   u.Cols,

		EnableProgress: u.EnableProgress,
		EnableHints:    u.EnableHints,
	}
}

// Write implements the io.Writer interface
// The provided bytes are split on newlines, and written with the UI's
// configured indent.
func (u *UI) Write(p []byte) (n int, err error) {
	parts := bytes.Split(p, []byte{'\n'})

	indent := bytes.Repeat([]byte{' '}, u.Indent)
	for i, part := range parts {
		if len(part) > 0 {
			part = append(indent, part...)
		}
		os.Stdout.Write(part)
		if i < len(parts)-1 {
			fmt.Println()
		}
	}

	return len(p), nil
}

func screenWidth() int {
	w := readline.GetScreenWidth()
	if w <= 0 {
		w = defaultCols
	}

	return w - rightPad
}
