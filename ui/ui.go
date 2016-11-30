package ui

import (
	"fmt"

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
func Hint(str string) {
	if !enableHints {
		return
	}
	fmt.Println(str)
}
