package prompts

import (
	"fmt"

	"github.com/manifoldco/promptui"

	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/prefs"
	"github.com/manifoldco/torus-cli/ui"
	"github.com/manifoldco/torus-cli/validate"
)

// Confirm prompts the user to perform an action, accepting optional
// labels and a preamble warning.
//
// If no warning is provided, a warning will not be displayed. If no label is
// provided, a default label will be used.
//
// A confirm prompt should be defaultYes if the user has asked for this
// explicit action (e.g. we want to destroy a machine). If this prompt is due
// to a possible side effect (e.g. changing a password while updating a profie)
// then defaultYes should be false.
//
// Similar to defaultYes, defaultValue should be true in situations where the
// user is explicitly asking to do something. This way, when we're not attached
// to a terminal we won't ask the user to confirm they want to perform the
// action, we'll just do it. In these situations where scripting is expected,
// defaultValue can be set by the value of a `--yes` flag.
func Confirm(labelOverride, warning *string, defaultValue bool, defaultYes bool) (bool, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return false, nil
	}

	if preferences.Core.AutoConfirm {
		return true, nil
	}

	if defaultValue && !ui.Attached() {
		return true, nil
	}

	if !ui.Attached() {
		return false, errs.ErrTerminalRequired
	}

	label := "Do you wish to continue"
	if labelOverride != nil {
		label = *labelOverride
	}

	if warning != nil {
		ui.Warn(*warning)
	}

	options := "(Y/n)"
	if defaultYes {
		options = "(y/N)"
	}

	questionTmpl := `%s {{ . | bold }} {{ "` + options + `" | faint }}? `
	confirmer := validate.Confirmer(defaultYes)
	tmpl := &promptui.PromptTemplates{
		Confirm:         fmt.Sprintf(questionTmpl, ui.BoldString(promptui.IconInitial)),
		Prompt:          fmt.Sprintf(questionTmpl, ui.BoldString(promptui.IconGood)),
		Valid:           fmt.Sprintf(questionTmpl, ui.BoldString(promptui.IconGood)),
		Success:         "", // promptui doesnt use this right now
		Invalid:         fmt.Sprintf(questionTmpl, ui.BoldString(promptui.IconBad)),
		ValidationError: validationErrorTmpl,
	}

	// Use a template to clean up the actual prompts. error state and finish
	// state
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
		IsVimMode: preferences.Core.Vim,
		Validate:  promptui.ValidateFunc(confirmer),
		Templates: tmpl,
	}

	_, err = prompt.Run()
	if err == promptui.ErrAbort || err == promptui.ErrEOF || err == promptui.ErrInterrupt {
		return false, nil
	}

	if err == nil {
		return true, nil
	}

	return false, err
}
