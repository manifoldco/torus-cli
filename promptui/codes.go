package promptui

import (
	"fmt"
	"strconv"
	"strings"
)

const esc = "\033["

type attribute int

const (
	reset attribute = iota
	fgBold
	fgFaint
	fgItalic
	fgUnderline
)

const (
	fgBlack attribute = iota + 30
	fgRed
	fgGreen
	fgYellow
	fgBlue
	fgMagenta
	fgCyan
	fgWhite
)

var (
	resetCode  = fmt.Sprintf("%s%dm", esc, reset)
	hideCursor = esc + "?25l"
	showCursor = esc + "?25h"
	clearLine  = esc + "2K"
)

func upLine(n uint) string {
	return movementCode(n, 'A')
}

func downLine(n uint) string {
	return movementCode(n, 'B')
}

func movementCode(n uint, code rune) string {
	return esc + strconv.FormatUint(uint64(n), 10) + string(code)
}

func styler(attrs ...attribute) func(string) string {
	attrstrs := make([]string, len(attrs))
	for i, v := range attrs {
		attrstrs[i] = strconv.Itoa(int(v))
	}

	seq := strings.Join(attrstrs, ";")

	return func(s string) string {
		end := ""
		if !strings.HasSuffix(s, resetCode) {
			end = resetCode
		}
		return fmt.Sprintf("%s%sm%s%s", esc, seq, s, end)
	}
}
