package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/promptui"
)

const slugPattern = "^[a-z][a-z0-9\\-\\_]{0,63}$"
const namePattern = "^[a-zA-Z\\s,\\.'\\-pL]{1,64}$"
const inviteCodePattern = "^[0-9a-ht-zjkmnpqr]{10}$"

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

// NamePrompt prompts the user to input a person's name
func NamePrompt(override *string, defaultValue string) (string, error) {
	label := "Name"
	if override != nil {
		label = *override
	}
	prompt := promptui.Prompt{
		Label:    label,
		Default:  defaultValue,
		Validate: validateSlug(strings.ToLower(label)),
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
	return cli.NewExitError(generic, -1)
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
func SelectCreateProject(client *api.Client, c context.Context, orgID *identity.ID, name string) (*api.ProjectResult, string, bool, error) {
	var projects []api.ProjectResult
	var err error
	if orgID != nil {
		// Get the list of projects the user has access to in the specified org
		projects, err = client.Projects.List(c, orgID, nil)
		if err != nil {
			return nil, "", false, cli.NewExitError("Error fetching projects list", -1)
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
			return nil, "", false, cli.NewExitError("Project not found", -1)
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
func SelectCreateOrg(client *api.Client, c context.Context, name string) (*api.OrgResult, string, bool, error) {
	// Get the list of orgs the user has access to
	orgs, err := client.Orgs.List(c)
	if err != nil {
		return nil, "", false, cli.NewExitError("Error fetching orgs list", -1)
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
			return nil, "", false, cli.NewExitError("Org not found", -1)
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
		Label:   "Invite Code",
		Default: defaultValue,
		Validate: func(input string) error {
			if govalidator.StringMatches(input, inviteCodePattern) {
				return nil
			}
			return promptui.NewValidationError("Please enter a valid invite code")
		},
	}

	return prompt.Run()
}
