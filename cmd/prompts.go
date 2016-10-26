package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/identity"
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

// NamePrompt prompts the user to input a person's name
func NamePrompt(override *string, defaultValue string, autoAccept bool) (string, error) {
	var prompt promptui.Prompt

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
		Label:    label,
		Default:  defaultValue,
		Validate: validateSlug(strings.ToLower(label)),
	}
	return prompt.Run()
}

// VerificationPrompt prompts the user to input an email verify code
func VerificationPrompt() (string, error) {
	prompt := promptui.Prompt{
		Label: "Verification code",
		Validate: func(input string) error {
			input = strings.ToLower(input)
			if govalidator.StringMatches(input, verifyCodePattern) {
				return nil
			}
			return promptui.NewValidationError("Please enter a valid code")
		},
	}

	return prompt.Run()
}

// SelectProjectPrompt prompts the user to select an org from a list, or enter a new name
func SelectProjectPrompt(projects []api.ProjectResult) (int, string, error) {
	names := make([]string, len(projects))
	for i, p := range projects {
		names[i] = p.Body.Name
	}

	// Get the user's org selection
	prompt := promptui.SelectWithAdd{
		Label:    "Select project",
		Items:    names,
		AddLabel: "Create a new project",
		Validate: validateSlug("project"),
	}

	return prompt.Run()
}

// SelectOrgPrompt prompts the user to select an org from a list, or enter a new name
func SelectOrgPrompt(orgs []api.OrgResult) (int, string, error) {
	names := make([]string, len(orgs))
	for i, o := range orgs {
		names[i] = o.Body.Name
	}

	// Get the user's org selection
	prompt := promptui.SelectWithAdd{
		Label:    "Select organization",
		Items:    names,
		AddLabel: "Create a new organization",
		Validate: validateSlug("org"),
	}

	return prompt.Run()
}

func handleSelectError(err error, generic string) error {
	if err == promptui.ErrEOF || err == promptui.ErrInterrupt {
		return err
	}
	return errs.NewExitError(generic)
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
			return nil, "", false, errs.NewExitError("Error fetching projects list.")
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
		return nil, "", false, errs.NewExitError("Error fetching orgs list")
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

// PasswordPrompt prompts the user to input a password value
func PasswordPrompt(shouldConfirm bool) (string, error) {
	prompt := promptui.Prompt{
		Label: "Password",
		Mask:  '●',
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
	}

	password, err := prompt.Run()
	if err != nil {
		return "", err
	}
	if !shouldConfirm {
		return password, err
	}

	prompt = promptui.Prompt{
		Label: "Confirm Password",
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
	}

	_, err = prompt.Run()
	if err != nil {
		return "", err
	}

	return password, nil
}

// EmailPrompt prompts the user to input an email
func EmailPrompt(defaultValue string) (string, error) {
	prompt := promptui.Prompt{
		Label: "Email",
		Validate: func(input string) error {
			if govalidator.IsEmail(input) {
				return nil
			}
			return promptui.NewValidationError("Please enter a valid email address")
		},
	}
	if defaultValue != "" {
		prompt.Default = defaultValue
	}

	return prompt.Run()
}

// UsernamePrompt prompts the user to input a person's name
func UsernamePrompt() (string, error) {
	prompt := promptui.Prompt{
		Label: "Username",
		Validate: func(input string) error {
			if govalidator.StringMatches(input, slugPattern) {
				return nil
			}
			return promptui.NewValidationError("Please enter a valid username")
		},
	}

	return prompt.Run()
}

// FullNamePrompt prompts the user to input a person's name
func FullNamePrompt() (string, error) {
	prompt := promptui.Prompt{
		Label: "Name",
		Validate: func(input string) error {
			if govalidator.StringMatches(input, namePattern) {
				return nil
			}
			return promptui.NewValidationError("Please enter a valid name")
		},
	}

	return prompt.Run()
}

// InviteCodePrompt prompts the user to input an invite code
func InviteCodePrompt(defaultValue string) (string, error) {
	prompt := promptui.Prompt{
		Label:    "Invite Code",
		Default:  defaultValue,
		Validate: validateInviteCode,
	}

	return prompt.Run()
}

// SelectAcceptAction prompts the user to select an org from a list, or enter a new name
func SelectAcceptAction() (int, string, error) {
	names := []string{
		"Login",
		"Signup",
	}

	// Get the user's org selection
	prompt := promptui.Select{
		Label: "Do you want to login or create an account?",
		Items: names,
	}

	return prompt.Run()
}
