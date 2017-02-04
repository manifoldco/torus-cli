package promptui

import (
	"fmt"
	"strconv"
)

const esc = "\033["

type attribute int

// Forground weight/decoration attributes.
const (
	reset attribute = iota

	FGBold
	FGFaint
	FGItalic
	FGUnderline
)

// Forground color attributes
const (
	FGBlack attribute = iota + 30
	FGRed
	FGGreen
	FGYellow
	FGBlue
	FGMagenta
	FGCyan
	FGWhite
)

// ResetCode is the character code used to reset the terminal formatting
var ResetCode = fmt.Sprintf("%s%dm", esc, reset)

var (
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

// Styler returns a func that applies the attributes given in the Styler call
// to the provided string.
func Styler(attrs ...attribute) func(string) string {
	return styler(attrs...)
}
