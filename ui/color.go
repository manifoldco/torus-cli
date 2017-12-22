package ui

import (
	"bytes"
	"os"

	"github.com/chzyer/readline"
	"github.com/juju/ansiterm"
)

// Progress calls Progress on the default UI
func Bold(s string) string { return defUI.Bold(s) }

func (u *UI) Bold(s string) string {
	if !u.EnableColors {
		return s
	}

	ctx := ansiterm.Context{
		Styles: []ansiterm.Style{ansiterm.Bold},
	}

	return createStyledString(ctx, s)
}

func Faint(s string) string { return defUI.Faint(s) }

func (u *UI) Faint(s string) string {
	if !u.EnableColors {
		return s
	}

	return Color(DarkGray, s)
}

func Color(c UIColor, s string) string { return defUI.Color(c, s) }

func (u *UI) Color(c UIColor, s string) string {
	if !u.EnableColors {
		return s
	}

	ctx := ansiterm.Context{
		Foreground: ansiterm.Color(c),
	}

	return createStyledString(ctx, s)
}

func BoldColor(c UIColor, s string) string { return defUI.BoldColor(c, s) }

func (u *UI) BoldColor(c UIColor, s string) string {
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
	if !readline.IsTerminal(int(os.Stdout.Fd())) {
		return s
	}

	buf := bytes.Buffer{}
	w := ansiterm.NewWriter(&buf)
	w.SetColorCapable(true)
	ctx.Fprint(w, s)
	return buf.String()
}
