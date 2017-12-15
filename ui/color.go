package ui

import (
  "bytes"

	"github.com/juju/ansiterm"
)

var (
  ServiceColor      = ansiterm.Magenta
  EnvironmentColor  = ansiterm.BrightRed
  CredPathColor     = ansiterm.DarkGray
)

var CredStyle       = ansiterm.Bold

func Bold(s string) string {
  ctx := ansiterm.Context {
    Styles: []ansiterm.Style{ansiterm.Bold},
  }

  return createStyledString(ctx, s)
}

func Color(c ansiterm.Color, s string) string {
  ctx := ansiterm.Context {
    Foreground: c,
  }

  return createStyledString(ctx, s)
}

func Environment(s string) string {
  return Color(EnvironmentColor, s)
}

func Service(s string) string {
  return Color(ServiceColor, s)
}

func Cred(s string) string {
  return Bold(s)
}

func CredPath(s string) string {
  return Color(CredPathColor, s)
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
