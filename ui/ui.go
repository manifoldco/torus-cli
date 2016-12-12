package ui

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"github.com/kr/text"

	"github.com/manifoldco/torus-cli/prefs"
)

// enableProgress is whether progress events should be displayed
var enableProgress = false

// enableHints is whether hints should be displayed
var enableHints = false

// Init prepares the ui preferences
func Init(preferences *prefs.Preferences) {
	enableProgress = preferences.Core.EnableProgress
	enableHints = preferences.Core.EnableHints
}

// Progress handles the ui output for progress events, when enabled
func Progress(str string) {
	if !enableProgress {
		return
	}
	fmt.Println(str)
}

// Hint handles the ui output for hint/onboarding messages, when enabled
func Hint(str string, noPadding bool) {
	if !enableHints {
		return
	}
	if !noPadding {
		fmt.Println("")
	}
	printWrapLabeled("Protip:", str)
}

func printWrapLabeled(label, message string) {
	cols := readline.GetScreenWidth() - 2
	fmt.Printf("%s  ", label)
	longest := len(label)
	wrapped := text.Wrap(message, cols-(2+longest))
	fmt.Println(indentOthers(wrapped, 2+longest))
}

func indentOthers(str string, indent int) string {
	nl := strings.IndexRune(str, '\n')
	if nl == -1 {
		nl = len(str)
	}
	return str[:nl] + text.Indent(str[nl:], fmt.Sprintf("%*s", indent, ""))
}
