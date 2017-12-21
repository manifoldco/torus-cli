package ui

import (
  "bytes"

	"github.com/juju/ansiterm"
)

func Bold(s string) string {
  ctx := ansiterm.Context {
    Styles: []ansiterm.Style{ansiterm.Bold},
  }

  return createStyledString(ctx, s)
}

func Faint(s string) string {
  return Color(ansiterm.DarkGray, s)
}

func Color(c ansiterm.Color, s string) string {
  ctx := ansiterm.Context {
    Foreground: c,
  }

  return createStyledString(ctx, s)
}

func BoldColor(c ansiterm.Color, s string) string {
  ctx := ansiterm.Context {
    Foreground: c,
    Styles: []ansiterm.Style{ansiterm.Bold},
  }

  return createStyledString(ctx, s)
}

func createStyledString(ctx ansiterm.Context, s string) string {
  buf := bytes.Buffer{}
  w := ansiterm.NewWriter(&buf)
  w.SetColorCapable(true)
  ctx.Fprint(w, s)
  return buf.String()
}
