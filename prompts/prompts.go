package prompts

import (
	"fmt"

	"github.com/manifoldco/promptui"

	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/prefs"
	"github.com/manifoldco/torus-cli/ui"
	"github.com/manifoldco/torus-cli/validate"
)

// StringPrompt represents a function that prompts the user and returns a
// string. It supports accepting a string as a default value and a boolean
// representing whether it should automatically validate the data.
//
// When the client is not attached to a terminal the prompt will validate the
// defaultValue returning an error of the default value. If no value is
// provided then an error is returned.
type StringPrompt func(string, bool) (string, error)

// Email asks the user to provide an email address
var Email StringPrompt

// InviteCode ask the user to provide an org invite code
var InviteCode StringPrompt

// VerificationCode asks the user to provide a verification code
var VerificationCode StringPrompt

// FullName asks the user to provide a full name
var FullName StringPrompt

// Username asks the user to provide a username
var Username StringPrompt

// OrgName asks the user to provide a name for an org
var OrgName StringPrompt

// TeamName asks the user to provide a name for a team
var TeamName StringPrompt

// RoleName asks the user to provide a name for a role (team)
var RoleName StringPrompt

// ProjectName asks the user to provide a name for a project
var ProjectName StringPrompt

// EnvName asks the user to provide a name for an environment
var EnvName StringPrompt

// ServiceName asks the user to provide a name for a service
var ServiceName StringPrompt

// MachineName asks the user to provide a name for a machine
var MachineName StringPrompt

// DefaultWarning represents the default warning to be used with a
// ConfirmDialog
var DefaultWarning = "The actions you are about to perform cannot be undone."

const pwMask = '●'

// Password prompts the user to provide a password. The value provided by the
// user is masked.
//
// If confirm is true the user will be prompted to supply the password again.
func Password(confirm bool, override *string) (string, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return "", err
	}

	label := "Password"
	if override != nil {
		label = *override
	}

	if !ui.Attached() {
		return "", errs.ErrTerminalRequired
	}

	prompt := &promptui.Prompt{
		Label:     label,
		Mask:      pwMask,
		Validate:  promptui.ValidateFunc(validate.Password),
		IsVimMode: preferences.Core.Vim,
	}

	password, err := prompt.Run()
	if err != nil {
		return "", err
	}

	if !confirm {
		return password, err
	}

	prompt = &promptui.Prompt{
		Label:     "Confirm " + label,
		Mask:      pwMask,
		Validate:  promptui.ValidateFunc(validate.ConfirmPassword(password)),
		IsVimMode: preferences.Core.Vim,
	}

	_, err = prompt.Run()
	if err != nil {
		return "", convertErr(err)
	}

	return password, nil
}

func stringTmpl() *promptui.PromptTemplates {
	questionTmpl := `%s {{ . | bold }}: `
	return &promptui.PromptTemplates{
		Prompt:          fmt.Sprintf(questionTmpl, ui.BoldString(promptui.IconInitial)),
		Valid:           fmt.Sprintf(questionTmpl, ui.BoldString(promptui.IconGood)),
		Invalid:         fmt.Sprintf(questionTmpl, ui.BoldString(promptui.IconBad)),
		Success:         "", // promptui doesnt use this right now
		ValidationError: validationErrorTmpl,
	}
}

func stringPrompt(label string, validator validate.Func) StringPrompt {
	return func(providedValue string, autoAccept bool) (string, error) {
		preferences, err := prefs.NewPreferences()
		if err != nil {
			return "", err
		}

		tmpl := stringTmpl()

		if !ui.Attached() {
			err = validator(providedValue)
			if err != nil {
				return "", err
			}

			return providedValue, nil
		}

		if autoAccept && providedValue != "" {
			err = validator(providedValue)
			if err != nil {
				fmt.Println(direct(tmpl.Invalid, label, providedValue))
			} else {
				fmt.Println(direct(tmpl.Valid, label, providedValue))
			}

			return providedValue, err
		}

		prompt := &promptui.Prompt{
			Label:     label,
			Default:   providedValue,
			Validate:  promptui.ValidateFunc(validator),
			IsVimMode: preferences.Core.Vim,
			Templates: tmpl,
		}

		result, err := prompt.Run()
		if err != nil {
			return "", convertErr(err)
		}

		return result, nil
	}
}

func init() {
	Email = stringPrompt("Email", validate.Email)
	InviteCode = stringPrompt("Invite Code", validate.InviteCode)
	VerificationCode = stringPrompt("Verification Code", validate.VerificationCode)
	FullName = stringPrompt("Full Name", validate.Name)
	Username = stringPrompt("Username", validate.SlugValidator("Usernames"))
	OrgName = stringPrompt("Org Name", validate.SlugValidator("Org names"))
	TeamName = stringPrompt("Team Name", validate.SlugValidator("Team names"))
	RoleName = stringPrompt("Role Name", validate.SlugValidator("Role names"))
	ProjectName = stringPrompt("Project Name", validate.SlugValidator("Project names"))
	EnvName = stringPrompt("Environment Name", validate.SlugValidator("Environment names"))
	ServiceName = stringPrompt("Service Name", validate.SlugValidator("Service names"))
	MachineName = stringPrompt("Machine Name", validate.SlugValidator("Machine names"))
}
