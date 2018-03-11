package cmd

import (
	"context"
	"errors"
	"strings"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/pathexp"
	"github.com/manifoldco/torus-cli/primitive"
	"github.com/manifoldco/torus-cli/prompts"
)

// ByTeamPrecedence will implement the sort Interface
type ByTeamPrecedence []envelope.Team

func (tp ByTeamPrecedence) Len() int {
	return len(tp)
}

func (tp ByTeamPrecedence) Swap(i, j int) {
	tp[i], tp[j] = tp[j], tp[i]
}

func (tp ByTeamPrecedence) Less(i, j int) bool {

	// 1) 	If both teams are system teams, must compare owner < admin < member
	//		where "<" means has higher precedence (ie. owner (0) < admin (1) where
	//		owner has a higher precedence).
	//
	// 2) 	Any system team automatically has higher precedence than a non-system team.
	//
	// 3) 	If both teams are non-system user teams, then list in alphabetical order.
	//
	// 4)	If a team is a machine team, it automatically has the lowest precedence.
	//
	// 5) 	If both teams are machine teams, then list in alphabetical order.
	//
	if tp[i].Body.TeamType == "system" && tp[j].Body.TeamType == "system" {
		return primitive.SystemTeams[tp[i].Body.Name] < primitive.SystemTeams[tp[j].Body.Name]
	} else if tp[i].Body.TeamType == "system" {
		return true
	} else if tp[j].Body.TeamType == "system" {
		return false
	} else if tp[i].Body.TeamType == "user" && tp[j].Body.TeamType == "user" {
		return tp[i].Body.Name[0] < tp[j].Body.Name[0]
	} else if tp[i].Body.TeamType == "user" {
		return true
	} else if tp[j].Body.TeamType == "user" {
		return false
	}
	return tp[i].Body.Name[0] < tp[j].Body.Name[0]
}

func selectOrg(ctx context.Context, client *api.Client, provided string, allowCreate bool) (*envelope.Org, string, bool, error) {
	orgs, err := client.Orgs.List(ctx)
	if err != nil {
		return nil, "", false, err
	}

	var idx int
	var oName string
	if allowCreate {
		idx, oName, err = prompts.SelectCreateOrg(toOrgNames(orgs), provided)
	} else {
		idx, oName, err = prompts.SelectOrg(toOrgNames(orgs), provided)
	}

	if err != nil {
		return nil, "", false, err
	}

	if idx == prompts.SelectedAdd {
		return nil, oName, true, nil
	}

	return &orgs[idx], oName, false, nil
}

func selectProject(ctx context.Context, client *api.Client, org *envelope.Org, provided string, allowCreate bool) (*envelope.Project, string, bool, error) {
	var projects []envelope.Project
	var err error
	if org != nil {
		projects, err = client.Projects.List(ctx, org.ID)
		if err != nil {
			return nil, "", false, err
		}
	}

	var idx int
	var pName string
	if allowCreate {
		idx, pName, err = prompts.SelectCreateProject(toProjectNames(projects), provided)
	} else {
		idx, pName, err = prompts.SelectProject(toProjectNames(projects), provided)
	}

	if err != nil {
		return nil, "", false, err
	}

	if idx == prompts.SelectedAdd {
		return nil, pName, true, nil
	}

	return &projects[idx], pName, false, nil
}

func selectRole(ctx context.Context, client *api.Client, org *envelope.Org, provided string, allowCreate bool) (*envelope.Team, string, bool, error) {
	teams, err := client.Teams.List(ctx, org.ID, "", primitive.MachineTeamType)
	if err != nil {
		return nil, "", false, err
	}

	var idx int
	var name string
	if allowCreate {
		idx, name, err = prompts.SelectCreateRole(toTeamNames(teams), provided)
	} else {
		idx, name, err = prompts.SelectRole(toTeamNames(teams), provided)
	}

	if err != nil {
		return nil, "", false, err
	}

	if idx == prompts.SelectedAdd {
		return nil, name, true, nil
	}

	return &teams[idx], name, false, nil
}

func selectTeam(ctx context.Context, client *api.Client, org *envelope.Org, provided string, allowCreate bool) (*envelope.Team, string, bool, error) {
	teams, err := client.Teams.List(ctx, org.ID, "", primitive.AnyTeamType)
	if err != nil {
		return nil, "", false, err
	}

	// XXX: This is because there isn't a good way to just get "user" teams
	// right now..
	filtered := []envelope.Team{}
	for _, team := range teams {
		if !isMachineTeam(team.Body) {
			filtered = append(filtered, team)
		}
	}

	var idx int
	var name string
	if allowCreate {
		idx, name, err = prompts.SelectCreateTeam(toTeamNames(teams), provided)
	} else {
		idx, name, err = prompts.SelectTeam(toTeamNames(teams), provided)
	}

	if err != nil {
		return nil, "", false, err
	}

	if idx == prompts.SelectedAdd {
		return nil, name, true, nil
	}

	return &teams[idx], name, false, nil
}

func toOrgNames(orgs []envelope.Org) []string {
	out := make([]string, len(orgs), len(orgs))
	for i, o := range orgs {
		out[i] = o.Body.Name
	}

	return out
}

func toProjectNames(projects []envelope.Project) []string {
	out := make([]string, len(projects), len(projects))
	for i, p := range projects {
		out[i] = p.Body.Name
	}

	return out
}

func toTeamNames(teams []envelope.Team) []string {
	out := make([]string, len(teams), len(teams))
	for i, t := range teams {
		out[i] = t.Body.Name
	}

	return out
}

func argCheck(ctx *cli.Context, possible, required int) error {
	given := len(ctx.Args())
	if given > possible {
		return errs.NewUsageExitError("Too many arguments provided", ctx)
	}

	if given < required {
		return errs.NewUsageExitError("Too few arguments provided", ctx)
	}

	return nil
}

func displayPathExp(pe *pathexp.PathExp) string {
	if pe.Identities.String() == "*" && pe.Instances.String() == "*" {
		return strings.Join([]string{"", pe.Org.String(), pe.Project.String(),
			pe.Envs.String(),
			pe.Services.String(),
		}, "/")
	}

	return pe.String()
}

func displayResourcePath(path string) (string, error) {
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		return "", errors.New("invalid resource path provided")
	}

	pe, err := parsePathExp(path[:idx])
	if err != nil {
		return "", err
	}

	return displayPathExp(pe) + "/" + path[idx+1:], nil
}

func parsePathExp(path string) (*pathexp.PathExp, error) {
	parts := strings.Split(path, "/")
	if parts[0] != "" {
		return nil, errors.New("path expressions must start with '/'")
	}

	parts = parts[1:] // meove leading empty section

	switch len(parts) {
	case 5:
		if parts[4] == "**" {
			return pathexp.Parse(path)
		}

		parts = append(parts, "*", "*")
		path = "/" + strings.Join(parts, "/")

		return pathexp.Parse(path)
	default:
		return pathexp.ParsePartial(path)
	}
}

func deriveIdentity(session *api.Session) string {
	if session.Type() == apitypes.MachineSession {
		return "machine-" + session.Username()
	}

	return session.Username()
}

func filterTeamsByNames(names []string, teams []envelope.Team) ([]envelope.Team, error) {
	teamNamesToTeam := toTeamNameMap(teams)
	filteredTeams := make([]envelope.Team, len(names))

	for i, name := range names {
		if team, ok := teamNamesToTeam[name]; ok {
			filteredTeams[i] = team
		} else {
			return []envelope.Team{}, errs.NewExitError("No such team " + name)
		}
	}

	return filteredTeams, nil
}

func toTeamNameMap(teams []envelope.Team) map[string]envelope.Team {
	teamNamesToTeam := make(map[string]envelope.Team)
	for _, t := range teams {
		teamNamesToTeam[t.Body.Name] = t
	}
	return teamNamesToTeam
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
