package ui

import (
	"bytes"

	"github.com/juju/ansiterm"
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
