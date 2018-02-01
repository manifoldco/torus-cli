package prompts

import (
	"fmt"

	"github.com/manifoldco/promptui"

	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/prefs"
	"github.com/manifoldco/torus-cli/ui"
	"github.com/manifoldco/torus-cli/validate"
)

// SelectPrompt represents a function that prompts the user with a list of
// choices they can select. If the choice exists then the index and displayed
// name of the item are returned and the prompt is not displayed.
type SelectPrompt func([]string, string) (int, string, error)

// SelectCreatePrompt represents a function that prompts the user witha  list
// of choices they can select or the ability to create a new item. If they
// choose to create a new item they are prompted with a string prompt. The
// index returned will be SelectedAdd.
//
// A user can provide a default choice, if it exists then the index and
// displayed name of the item are returned. The success state of the prompt is
// displayed.
type SelectCreatePrompt func([]string, string) (int, string, error)

// SelectOrg prompts the user to select or create an org depending on arguments
// provided by the callee.

// SelectOrg prompts the user to select an org.
var SelectOrg SelectPrompt

// SelectCreateOrg prompts the user to select or create an org
var SelectCreateOrg SelectCreatePrompt

// SelectRole prompts the user to select a role.
var SelectRole SelectPrompt

// SelectCreateRole prompts the user to select or create a new role.
var SelectCreateRole SelectCreatePrompt

// SelectTeam prompts the user to select a team.
var SelectTeam SelectPrompt

// SelectCreateTeam prompts the user to select or create a new team.
var SelectCreateTeam SelectCreatePrompt

// SelectProject prompts the user to select a project.
var SelectProject SelectPrompt

// SelectCreateProject prompts the user to select or create a new project.
var SelectCreateProject SelectCreatePrompt

// SelectAcceptAction prompts the user to select whether they want to log into
// an existing account or to create a new one when they're accepting an org
// invitation.
func SelectAcceptAction() (int, string, error) {
	label := "Do you want to login or create an account?"
	v := []string{"Login", "Signup"}
	p := selectPrompt(label, "Action", validate.OneOf(v))

	return p(v, "")
}

func selectTmpl(name string) *promptui.SelectTemplates {
	return &promptui.SelectTemplates{
		Label:    fmt.Sprintf(`%s {{ . | bold }}: `, promptui.IconInitial),
		Active:   fmt.Sprintf(`%s {{ . | underline }}`, promptui.IconSelect),
		Inactive: `  {{ . }}`,
		Selected: fmt.Sprintf(`{{ "%s" | green }} {{ "%s" | bold }}: {{ . | faint }}`, promptui.IconGood, name),
	}
}

func selectCreatePrompt(label, addLabel, name string, validator validate.Func) SelectCreatePrompt {
	return func(values []string, providedValue string) (int, string, error) {
		preferences, err := prefs.NewPreferences()
		if err != nil {
			return 0, "", err
		}

		tmpl := selectTmpl(name)
		createTmpl := stringTmpl()
		if providedValue != "" || !ui.Attached() {
			err = validator(providedValue)
			if err != nil {
				fmt.Println(direct(createTmpl.Invalid, name, providedValue))
				return 0, "", err
			}

			idx := find(values, providedValue)
			if idx == -1 {
				fmt.Println(direct(createTmpl.Invalid, name, providedValue))
				return 0, "", errs.NewNotFound(name)
			}

			fmt.Println(direct(createTmpl.Valid, name, providedValue))
			return idx, providedValue, nil
		}

		prompt := selectWithAdd{
			Label:           label,
			AddLabel:        addLabel,
			Items:           values,
			Validate:        promptui.ValidateFunc(validator),
			IsVimMode:       preferences.Core.Vim,
			SelectTemplates: tmpl,
			PromptTemplates: createTmpl,
		}

		idx, value, err := prompt.Run()
		if err != nil {
			return 0, "", convertErr(err)
		}

		if idx == promptui.SelectedAdd {
			return SelectedAdd, value, nil
		}

		return idx, value, nil
	}

}

func selectPrompt(label, name string, validator validate.Func) SelectPrompt {
	return func(values []string, providedValue string) (int, string, error) {
		preferences, err := prefs.NewPreferences()
		if err != nil {
			return 0, "", err
		}

		if providedValue != "" || !ui.Attached() {
			err = validator(providedValue)
			if err != nil {
				return 0, "", err
			}

			idx := find(values, providedValue)
			if idx == -1 {
				return 0, "", errs.NewNotFound(name)
			}

			return idx, providedValue, nil
		}

		tmpl := selectTmpl(name)
		prompt := promptui.Select{
			Label:     label,
			Items:     values,
			Templates: tmpl,
			IsVimMode: preferences.Core.Vim,
		}

		idx, value, err := prompt.Run()
		if err != nil {
			return 0, "", convertErr(err)
		}

		return idx, value, nil
	}
}

func find(values []string, name string) int {
	idx := -1
	for i, v := range values {
		if v == name {
			idx = i
			break
		}
	}

	return idx
}

func init() {
	SelectOrg = selectPrompt("Select an org", "Org", validate.OrgName)
	SelectCreateOrg = selectCreatePrompt("Select an org", "Create a new org", "Org", validate.OrgName)
	SelectRole = selectPrompt("Select a Role", "Machine Role", validate.RoleName)
	SelectCreateRole = selectCreatePrompt("Select a role", "Create a new role", "Machine Role", validate.RoleName)
	SelectTeam = selectPrompt("Select a team", "Team", validate.TeamName)
	SelectCreateTeam = selectCreatePrompt("Select a team", "Create a new team", "Team", validate.TeamName)
	SelectProject = selectPrompt("Select a project", "Project", validate.ProjectName)
	SelectCreateProject = selectCreatePrompt("Select a project", "Create a new project", "Project", validate.ProjectName)
}
