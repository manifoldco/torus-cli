package ui

import (
	"bytes"

	"github.com/juju/ansiterm"
)

// Color represents 1 of 16 ANSI color sequences
type Color int

const (
	_ Color = iota
	// Default sets the temrinal text to its default color
	Default
	// Black text
	Black
	// Red text
	Red
	//Green text
	Green
	// Yellow text
	Yellow
	// Blue text
	Blue
	// Magenta text
	Magenta
	// Cyan text
	Cyan
	// Gray text
	Gray
	// DarkGray text
	DarkGray
	// BrightRed text
	BrightRed
	// BrightGreen text
	BrightGreen
	// BrightYellow text
	BrightYellow
	// BrightBlue text
	BrightBlue
	// BrightMagenta text
	BrightMagenta
	// BrightCyan text
	BrightCyan
	// White text
	White
)

// BoldString returns a bolded copy of the string (ANSI escape sequenced)
func BoldString(s string) string { return defUI.BoldString(s) }

// BoldString checks the EnableColors pref and returns a bolded copy
// of the string (ANSI escape sequenced)
func (u *UI) BoldString(s string) string {
	if !u.EnableColors {
		return s
	}

	ctx := ansiterm.Context{
		Styles: []ansiterm.Style{ansiterm.Bold},
	}

	return createStyledString(ctx, s)
}

// FaintString returns a faint (ANSI DarkGray) copy of the string (ANSI escape sequenced)
func FaintString(s string) string { return defUI.FaintString(s) }

// FaintString checks the EnableColors pref and returns a faint (ANSI DarkGrey)
// copy of the string (ANSI escape sequenced)
func (u *UI) FaintString(s string) string {
	if !u.EnableColors {
		return s
	}

	return ColorString(DarkGray, s)
}

// ColorString returns the original string, converted to the provided ANSI color
func ColorString(c Color, s string) string { return defUI.ColorString(c, s) }

// ColorString checks the EnableColors pref and returns a colored
// copy of the string (ANSI escape sequenced)
func (u *UI) ColorString(c Color, s string) string {
	if !u.EnableColors {
		return s
	}

	ctx := ansiterm.Context{
		Foreground: ansiterm.Color(c),
	}

	return createStyledString(ctx, s)
}

// BoldColorString returns a bolded, colored copy of the string (ANSI escape sequenced)
func BoldColorString(c Color, s string) string { return defUI.BoldColorString(c, s) }

// BoldColorString checks the EnableColors pref and returns a bolded, colored
// copy of the string (ANSI escape sequenced)
func (u *UI) BoldColorString(c Color, s string) string {
	if !u.EnableColors {
		return s
	}

	ctx := ansiterm.Context{
		Foreground: ansiterm.Color(c),
		Styles:     []ansiterm.Style{ansiterm.Bold},
	}

	return createStyledString(ctx, s)
}

func createStyledString(ctx ansiterm.Context, s string) string {
	if !Attached() {
		return s
	}

	buf := bytes.Buffer{}
	w := ansiterm.NewWriter(&buf)
	w.SetColorCapable(true)
	ctx.Fprint(w, s)
	return buf.String()
}
