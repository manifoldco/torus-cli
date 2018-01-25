package ui

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/chzyer/readline"
	"github.com/manifoldco/ansiwrap"

	"github.com/manifoldco/torus-cli/prefs"
)

const (
	defaultCols = 80
	rightPad    = 2
)

type UIColor int

const (
	_ UIColor = iota
	Default
	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	Gray
	DarkGray
	BrightRed
	BrightGreen
	BrightYellow
	BrightBlue
	BrightMagenta
	BrightCyan
	White
)

var defUI *UI

// Init initializes a default global UI, accessible via the package functions.
func Init(preferences *prefs.Preferences) {
	enableColours := preferences.Core.EnableColors

	if !readline.IsTerminal(int(os.Stdout.Fd())) {
		enableColours = false
	}

	defUI = &UI{
		Indent: 0,
		Cols:   screenWidth(),

		EnableProgress: preferences.Core.EnableProgress,
		EnableHints:    preferences.Core.EnableHints,
		EnableColors:   enableColours,
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

	// EnableColors is whether formatted text should be colored
	EnableColors bool
}

// NewSpinner creates a new ui.Spinner struct (spinner.go)
func (u *UI) NewSpinner(text string) *Spinner {
	return NewSpinner(text)
}

// StartSpinner checks to ensure progress is enabled and the output is a
// terminal. After that, it starts the ui.Spinner spinning.
func (u *UI) StartSpinner(s *Spinner) {
  if !u.EnableProgress || !readline.IsTerminal(int(os.Stdout.Fd())) {
    return
  }

	s.Start()
}

// StartSpinner checks to ensure progress is enabled and the output is a
// terminal. After that, it stops the ui.Spinner spinning.
func (u *UI) StopSpinner(s *Spinner) {
	if !u.EnableProgress || !readline.IsTerminal(int(os.Stdout.Fd())) {
    return
  }

  s.Stop()
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
func Hint(str string, noPadding bool, label *string) { defUI.Hint(str, noPadding, label) }

// Hint handles the ui output for hint/onboarding messages, when enabled
func (u *UI) Hint(str string, noPadding bool, label *string) {
	if !u.EnableHints {
		return
	}

	if !readline.IsTerminal(int(os.Stdout.Fd())) {
		return
	}

	if !noPadding {
		fmt.Println()
	}

	hintLabel := u.Bold("Protip: ")
	if label != nil {
		hintLabel = u.Bold(*label)
	}
	rc := ansiwrap.RuneCount(hintLabel)
	fmt.Fprintln(readline.Stdout, ansiwrap.WrapIndent(hintLabel+str, u.Cols, u.Indent, u.Indent+rc))
}

// Info calls Info on the default UI
func Info(format string, args ...interface{}) { defUI.Info(format, args...) }

// Info handles outputting secondary information to the user such as messages
// about progress but are the actual result of an operation. For example,
// printing out that we're attempting to log a user in using the specific
// environment variables.
//
// Only printed if stdout is attached to a terminal.
func (u *UI) Info(format string, args ...interface{}) {
	if !readline.IsTerminal(int(os.Stdout.Fd())) {
		return
	}

	u.Line(format, args...)
}

// Warn calls Warn on the default UI
func Warn(format string, args ...interface{}) { defUI.Warn(format, args...) }

// Warn handles outputting warning information to the user such as
// messages about needing to be logged in.
//
// The warning is printed out to stderr if stdout is not attached to a
// terminal.
func (u *UI) Warn(format string, args ...interface{}) {
	var w io.Writer = readline.Stdout
	if !readline.IsTerminal(int(os.Stdout.Fd())) {
		w = readline.Stderr
	}

	o := fmt.Sprintf(format, args...)
	fmt.Fprintln(w, ansiwrap.WrapIndent(o, u.Cols, u.Indent, u.Indent))
}

// Error calls Error on the default UI
func Error(format string, args ...interface{}) { defUI.Error(format, args...) }

// Error handles outputting error information to the user such as the fact they
// couldn't log in due to an error.
//
// The error is printed out to stderr if stdout is not attached to a termainl
func (u *UI) Error(format string, args ...interface{}) {
	var w io.Writer = readline.Stdout
	if !readline.IsTerminal(int(os.Stdout.Fd())) {
		w = readline.Stderr
	}

	o := fmt.Sprintf(format, args...)
	fmt.Fprintln(w, ansiwrap.WrapIndent(o, u.Cols, u.Indent, u.Indent))
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
