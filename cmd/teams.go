package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
)

func init() {
	teams := cli.Command{
		Name:     "teams",
		Usage:    "View and manipulate teams within an organization",
		Category: "ORGANIZATIONS",
		Subcommands: []cli.Command{
			{
				Name:  "list",
				Usage: "List teams in an organization",
				Flags: []cli.Flag{
					StdOrgFlag,
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
					SetUserEnv, checkRequiredFlags, teamsListCmd,
				),
			},
		},
	}
	Cmds = append(Cmds, teams)
}

func teamsListCmd(ctx *cli.Context) error {
	orgName := ctx.String("org")

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	var getMemberships, display sync.WaitGroup
	getMemberships.Add(2)
	display.Add(2)

	var teams []api.TeamResult
	var org *api.OrgResult
	var self *api.UserResult
	var oErr, sErr error

	memberOf := make(map[identity.ID]bool)

	c := context.Background()

	go func() {
		self, sErr = client.Users.Self(c)
		getMemberships.Done()
	}()

	go func() {
		org, oErr = client.Orgs.GetByName(c, orgName)
		getMemberships.Done()

		if org == nil {
			oErr = cli.NewExitError("Org not found", -1)
		}

		if oErr == nil {
			teams, oErr = client.Teams.GetByOrg(c, org.ID)
		}

		display.Done()
	}()

	go func() {
		getMemberships.Wait()
		var memberships []api.MembershipResult
		if oErr == nil && sErr == nil {
			memberships, sErr = client.Memberships.List(c, org.ID, self.ID)
		}

		for _, m := range memberships {
			memberOf[*m.Body.TeamID] = true
		}
		display.Done()
	}()

	display.Wait()
	if oErr != nil || sErr != nil {
		return cli.NewMultiError(
			oErr,
			sErr,
			cli.NewExitError("Error fetching teams list", -1),
		)
	}

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	for _, t := range teams {
		isMember := ""
		teamType := ""
		if t.Body.TeamType == primitive.SystemTeam {
			teamType = "[system]"
		}

		if _, ok := memberOf[*t.ID]; ok {
			isMember = "*"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", isMember, t.Body.Name, teamType)
	}

	w.Flush()
	fmt.Println("\n  (*) member")
	return nil
}
