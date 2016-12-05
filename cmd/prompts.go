package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/prefs"
	"github.com/manifoldco/torus-cli/primitive"
	"github.com/manifoldco/torus-cli/promptui"
)

const slugPattern = "^[a-z][a-z0-9\\-\\_]{0,63}$"
const namePattern = "^[a-zA-Z\\s,\\.'\\-pL]{1,64}$"
const inviteCodePattern = "^[0-9a-ht-zjkmnpqr]{10}$"
const verifyCodePattern = "^[0-9a-ht-zjkmnpqr]{9}$"

func validateSlug(slugType string) promptui.ValidateFunc {
	msg := slugType + " names can only use a-z, 0-9, hyphens and underscores"
	err := promptui.NewValidationError(msg)
	return func(input string) error {
		if govalidator.StringMatches(input, slugPattern) {
			return nil
		}
		return err
	}
}

func validateInviteCode(input string) error {
	if govalidator.StringMatches(input, inviteCodePattern) {
		return nil
	}
	return promptui.NewValidationError("Please enter a valid invite code")
}

// AskPerform prompts the user if they want to do a specified action
func AskPerform(label string) error {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return err
	}
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
		Preamble:  nil,
		IsVimMode: preferences.Core.Vim,
	}
	_, err = prompt.Run()
	return err
}

// ConfirmDialogue prompts the user to confirm their action
func ConfirmDialogue(ctx *cli.Context, labelOverride, warningOverride *string, allowSkip bool) error {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return err
	}
	if allowSkip && (ctx.Bool("yes") || preferences.Core.AutoConfirm) {
		return nil
	}

	label := "Do you wish to continue"
	if labelOverride != nil {
		label = *labelOverride
	}

	warning := "The action you are about to perform cannot be undone."
	if warningOverride != nil {
		warning = *warningOverride
	}

	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
		Preamble:  &warning,
		IsVimMode: preferences.Core.Vim,
	}

	_, err = prompt.Run()
	return err
}

// NamePrompt prompts the user to input a person's name
func NamePrompt(override *string, defaultValue string, autoAccept bool) (string, error) {
	var prompt promptui.Prompt
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return "", err
	}

	label := "Name"
	if override != nil {
		label = *override
	}

	if autoAccept {
		err := validateSlug(strings.ToLower(label))(defaultValue)
		if err != nil {
			fmt.Println(promptui.FailedValue(label, defaultValue))
		} else {
			fmt.Println(promptui.SuccessfulValue(label, defaultValue))
		}
		return defaultValue, err
	}

	prompt = promptui.Prompt{
		Label:     label,
		Default:   defaultValue,
		Validate:  validateSlug(strings.ToLower(label)),
		IsVimMode: preferences.Core.Vim,
	}
	return prompt.Run()
}

// VerificationPrompt prompts the user to input an email verify code
func VerificationPrompt() (string, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return "", err
	}
	prompt := promptui.Prompt{
		Label: "Verification code",
		Validate: func(input string) error {
			input = strings.ToLower(input)
			if govalidator.StringMatches(input, verifyCodePattern) {
				return nil
			}
			return promptui.NewValidationError("Please enter a valid code")
		},
		IsVimMode: preferences.Core.Vim,
	}

	return prompt.Run()
}

// SelectProjectPrompt prompts the user to select an org from a list, or enter a new name
func SelectProjectPrompt(projects []api.ProjectResult) (int, string, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return 0, "", err
	}

	names := make([]string, len(projects))
	for i, p := range projects {
		names[i] = p.Body.Name
	}

	// Get the user's org selection
	prompt := promptui.SelectWithAdd{
		Label:     "Select project",
		Items:     names,
		AddLabel:  "Create a new project",
		Validate:  validateSlug("project"),
		IsVimMode: preferences.Core.Vim,
	}

	return prompt.Run()
}

// SelectOrgPrompt prompts the user to select an org from a list, or enter a new name
func SelectOrgPrompt(orgs []api.OrgResult) (int, string, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return 0, "", err
	}

	names := make([]string, len(orgs))
	for i, o := range orgs {
		names[i] = o.Body.Name
	}

	// Get the user's org selection
	prompt := promptui.SelectWithAdd{
		Label:     "Select organization",
		Items:     names,
		AddLabel:  "Create a new organization",
		Validate:  validateSlug("org"),
		IsVimMode: preferences.Core.Vim,
	}

	return prompt.Run()
}

// SelectTeamPrompt prompts the user to select a team from a list or enter a
// new name, an optional label can be provided.
func SelectTeamPrompt(teams []api.TeamResult, label, addLabel string) (int, string, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return 0, "", err
	}

	names := make([]string, len(teams))
	for i, t := range teams {
		names[i] = t.Body.Name
	}

	if label == "" {
		label = "Select Team"
	}

	if addLabel == "" {
		label = "Create a new team"
	}

	prompt := promptui.SelectWithAdd{
		Label:     label,
		Items:     names,
		AddLabel:  addLabel,
		Validate:  validateSlug("team"),
		IsVimMode: preferences.Core.Vim,
	}

	return prompt.Run()
}

func handleSelectError(err error, generic string) error {
	if err == promptui.ErrEOF || err == promptui.ErrInterrupt {
		return err
	}

	return errs.NewErrorExitError(generic, err)
}

// SelectCreateProject prompts the user to select a project from at list of projects
// populated via api request.
//
// The user may select to create a new project, or they may preselect an project
// via a non-empty name parameter.
//
// It returns the object of the selected project (if created a new project was not chosed),
// the name of the selected project, and a boolean indicating if a new project should
// be created.
func SelectCreateProject(c context.Context, client *api.Client, orgID *identity.ID, name string) (*api.ProjectResult, string, bool, error) {
	var projects []api.ProjectResult
	var err error
	if orgID != nil {
		// Get the list of projects the user has access to in the specified org
		projects, err = listProjects(&c, client, orgID, nil)
		if err != nil {
			return nil, "", false, err
		}
	}

	var idx int
	if name == "" {
		idx, name, err = SelectProjectPrompt(projects)
		if err != nil {
			return nil, "", false, err
		}
	} else {
		found := false
		for i, p := range projects {
			if p.Body.Name == name {
				found = true
				idx = i
				break
			}
		}
		if !found {
			fmt.Println(promptui.FailedValue("Project name", name))
			return nil, "", false, errs.NewExitError("Project not found.")
		}
		fmt.Println(promptui.SuccessfulValue("Project name", name))
	}

	if idx == promptui.SelectedAdd {
		return nil, name, true, nil
	}

	return &projects[idx], name, false, nil
}

// SelectCreateOrg prompts the user to select an org from at list of orgs
// populated via api request.
//
// The user may select to create a new org, or they may preselect an org
// via a non-empty name parameter.
//
// It returns the object of the selected org (if created a new org was not chosed),
// the name of the selected org, and a boolean indicating if a new org should
// be created.
func SelectCreateOrg(c context.Context, client *api.Client, name string) (*api.OrgResult, string, bool, error) {
	// Get the list of orgs the user has access to
	orgs, err := client.Orgs.List(c)
	if err != nil {
		return nil, "", false, err
	}

	var idx int

	if name == "" {
		idx, name, err = SelectOrgPrompt(orgs)
		if err != nil {
			return nil, "", false, err
		}
	} else {
		found := false
		for i, o := range orgs {
			if o.Body.Name == name {
				found = true
				idx = i
				break
			}
		}
		if !found {
			fmt.Println(promptui.FailedValue("Org name", name))
			return nil, "", false, errs.NewExitError("Org not found")
		}
		fmt.Println(promptui.SuccessfulValue("Org name", name))
	}

	if idx == promptui.SelectedAdd {
		return nil, name, true, nil
	}

	return &orgs[idx], name, false, nil
}

// SelectCreateRole prompts the user to select a machine team from a list of
// teams for the given org.
//
// The user may select to create a new team, or they may may preselect a team via
// a non-empty name parameter.
//
// It returns the object of the selected team, the name of the selected team,
// and a boolean indicating if a new team should be created.
func SelectCreateRole(c context.Context, client *api.Client, orgID *identity.ID, name string) (*api.TeamResult, string, bool, error) {

	teams, err := client.Teams.List(c, orgID, "", primitive.MachineTeam)
	if err != nil {
		return nil, "", false, err
	}

	label := "Select Machine Role"
	addLabel := "Create a new role"
	var idx int
	if name == "" {
		idx, name, err = SelectTeamPrompt(teams, label, addLabel)
		if err != nil {
			return nil, "", false, err
		}
	} else {
		found := false
		for i, t := range teams {
			if t.Body.Name == name {
				found = true
				idx = i
				break
			}
		}

		if !found {
			fmt.Println(promptui.FailedValue("Machine Role", name))
			return nil, "", false, errs.NewExitError("Role not found")
		}
		fmt.Println(promptui.SuccessfulValue("Machine Role", name))
	}

	if idx == promptui.SelectedAdd {
		return nil, name, true, nil
	}

	return &teams[idx], name, false, nil
}

// PasswordMask is the character used to mask password inputs
const PasswordMask = '●'

// PasswordPrompt prompts the user to input a password value
func PasswordPrompt(shouldConfirm bool, labelOverride *string) (string, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return "", err
	}

	label := "Password"
	if labelOverride != nil {
		label = *labelOverride
	}

	prompt := promptui.Prompt{
		Label: label,
		Mask:  PasswordMask,
		Validate: func(input string) error {
			length := len(input)
			if length >= 8 {
				return nil
			}
			if length > 0 {
				return promptui.NewValidationError("Passwords must be at least 8 characters")
			}

			return promptui.NewValidationError("Please enter your password")
		},
		IsVimMode: preferences.Core.Vim,
	}

	password, err := prompt.Run()
	if err != nil {
		return "", err
	}
	if !shouldConfirm {
		return password, err
	}

	prompt = promptui.Prompt{
		Label: "Confirm " + label,
		Mask:  '●',
		Validate: func(input string) error {
			if len(input) > 0 {
				if input != password {
					return promptui.NewValidationError("Passwords do not match")
				}
				return nil
			}

			return promptui.NewValidationError("Please confirm your password")
		},
		IsVimMode: preferences.Core.Vim,
	}

	_, err = prompt.Run()
	if err != nil {
		return "", err
	}

	return password, nil
}

// EmailPrompt prompts the user to input an email
func EmailPrompt(defaultValue string) (string, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return "", err
	}

	prompt := promptui.Prompt{
		Label: "Email",
		Validate: func(input string) error {
			if govalidator.IsEmail(input) {
				return nil
			}
			return promptui.NewValidationError("Please enter a valid email address")
		},
		IsVimMode: preferences.Core.Vim,
	}
	if defaultValue != "" {
		prompt.Default = defaultValue
	}

	return prompt.Run()
}

// UsernamePrompt prompts the user to input a person's name
func UsernamePrompt(un string) (string, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return "", err
	}

	prompt := promptui.Prompt{
		Label: "Username",
		Validate: func(input string) error {
			if govalidator.StringMatches(input, slugPattern) {
				return nil
			}
			return promptui.NewValidationError("Please enter a valid username")
		},
		IsVimMode: preferences.Core.Vim,
	}
	if un != "" {
		prompt.Default = un
	}

	return prompt.Run()
}

// FullNamePrompt prompts the user to input a person's name
func FullNamePrompt(name string) (string, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return "", err
	}

	prompt := promptui.Prompt{
		Label: "Name",
		Validate: func(input string) error {
			if govalidator.StringMatches(input, namePattern) {
				return nil
			}
			return promptui.NewValidationError("Please enter a valid name")
		},
		IsVimMode: preferences.Core.Vim,
	}
	if name != "" {
		prompt.Default = name
	}

	return prompt.Run()
}

// InviteCodePrompt prompts the user to input an invite code
func InviteCodePrompt(defaultValue string) (string, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return "", err
	}

	prompt := promptui.Prompt{
		Label:     "Invite Code",
		Default:   defaultValue,
		Validate:  validateInviteCode,
		IsVimMode: preferences.Core.Vim,
	}

	return prompt.Run()
}

// SelectAcceptAction prompts the user to select an org from a list, or enter a new name
func SelectAcceptAction() (int, string, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return 0, "", err
	}

	names := []string{
		"Login",
		"Signup",
	}

	prompt := promptui.Select{
		Label:     "Do you want to login or create an account?",
		Items:     names,
		IsVimMode: preferences.Core.Vim,
	}

	return prompt.Run()
}

// SelectProfileAction prompts the user to select an option from a list
func SelectProfileAction() (int, string, error) {
	preferences, err := prefs.NewPreferences()
	if err != nil {
		return 0, "", err
	}

	names := []string{
		"Name or email",
		"Change password",
	}

	// Get the user's org selection
	prompt := promptui.Select{
		Label:     "What would you like to update?",
		Items:     names,
		IsVimMode: preferences.Core.Vim,
	}

	return prompt.Run()
}
