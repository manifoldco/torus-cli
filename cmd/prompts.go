package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/promptui"
)

const slugPattern = "^[a-z][a-z0-9\\-_]{0,63}$"

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

// SelectCreateOrgAndProject prompts for org and project and creates them if necessary
func SelectCreateOrgAndProject(client *api.Client, c context.Context, ctx *cli.Context, oName, pName string) (*api.OrgResult, *api.ProjectResult, bool, error) {
	var org *api.OrgResult
	var project *api.ProjectResult
	var newResource bool
	var pIdx, oIdx int
	var pFound, oFound bool

	// Get the list of orgs the user has access to
	orgs, err := client.Orgs.List(c)
	if err != nil {
		return nil, nil, newResource, cli.NewExitError("Error fetching orgs list", -1)
	}

	if oName == "" {
		oIdx, oName, err = SelectOrgPrompt(orgs)
		if err != nil {
			return nil, nil, newResource, err
		}
	} else {
		for i, o := range orgs {
			if o.Body.Name == oName {
				oFound = true
				oIdx = i
				break
			}
		}
		if !oFound {
			fmt.Println(promptui.FailedValue("Org name", oName))
			return nil, nil, newResource, cli.NewExitError("Org not found", -1)
		}
		fmt.Println(promptui.SuccessfulValue("Org name", oName))
	}

	var projects []api.ProjectResult

	// Load existing projects for the selected org
	if oIdx != promptui.SelectedAdd {
		org = &orgs[oIdx]
		projects, err = client.Projects.List(c, org.ID, nil)
		if err != nil {
			return nil, nil, newResource, cli.NewExitError("Error fetching projects list", -1)
		}
	}

	if pName == "" {
		pIdx, pName, err = SelectProjectPrompt(projects)
		if err != nil {
			return nil, nil, newResource, err
		}
	} else {
		for i, p := range projects {
			if p.Body.Name == pName {
				pIdx = i
				pFound = true
				break
			}
		}
		if !pFound {
			fmt.Println(promptui.FailedValue("Project name", pName))
			return nil, nil, newResource, cli.NewExitError("Project not found", -1)
		}
		fmt.Println(promptui.SuccessfulValue("Project name", pName))
	}

	if oIdx == promptui.SelectedAdd || pIdx == promptui.SelectedAdd {
		fmt.Println("")
	}

	// Create org if required
	if oIdx == promptui.SelectedAdd {
		org, err = client.Orgs.Create(c, oName)
		if err != nil {
			return nil, nil, newResource, cli.NewExitError("Could not create org: "+err.Error(), -1)
		}

		var progress api.ProgressFunc = func(evt *api.Event, err error) {
			if evt != nil {
				fmt.Println(evt.Message)
			}
		}

		err = client.Keypairs.Generate(c, org.ID, &progress)
		if err != nil {
			msg := fmt.Sprintf("Could not generate keypairs for org. Run '%s keypairs generate' to fix.", ctx.App.Name)
			return nil, nil, newResource, cli.NewExitError(msg, -1)
		}

		newResource = true
		fmt.Printf("Org %s created.\n", oName)
	}

	// Create project if required
	if pIdx == promptui.SelectedAdd {
		project, err = client.Projects.Create(c, org.ID, pName)
		if err != nil {
			return nil, nil, newResource, cli.NewExitError("Could not create project: "+err.Error(), -1)
		}

		newResource = true
		fmt.Printf("Project %s created.\n", pName)
	} else {
		project = &projects[pIdx]
	}

	return org, project, newResource, nil
}
