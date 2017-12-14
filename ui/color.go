package ui

import (
  "bytes"

	"github.com/juju/ansiterm"
)

var (
  ServiceColor = ansiterm.Magenta
  EnvironmentColor = ansiterm.Blue
  SecretColor = ansiterm.Green
)

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

func Secret(s string) string {
  return Color(SecretColor, s)
}

func SetEnvironmentColor(c ansiterm.Color) {
  EnvironmentColor = c;
}

func SetServiceColor(c ansiterm.Color) {
  ServiceColor = c;
}

func SetSecretColor(c ansiterm.Color) {
  SecretColor = c;
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
