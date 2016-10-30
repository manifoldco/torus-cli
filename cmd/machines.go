package cmd

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/base32"
	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

const (
	machineRandomIDLength = 5 // 8 characters in base32
	machineCreateFailed   = "Could not create machine, please try again."
)

func init() {
	machines := cli.Command{
		Name:     "machines",
		Usage:    "Manage machine for an organization",
		Category: "ORGANIZATIONS",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create a machine for an organization",
				Flags: []cli.Flag{
					orgFlag("the org the machine will belong to", false),
					teamFlag("the team the machine will belong to", false),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, createMachine,
				),
			},
		},
	}
	Cmds = append(Cmds, machines)
}

func createMachine(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, orgName, newOrg, err := SelectCreateOrg(c, client, ctx.String("org"))
	if err != nil {
		return handleSelectError(err, "Org selection failed.")
	}

	var orgID *identity.ID
	if !newOrg {
		if org == nil {
			return errs.NewExitError("Org not found.")
		}
		orgID = org.ID
	}

	team, teamName, newTeam, err := SelectCreateTeam(
		c, client, orgID, primitive.MachineTeam, ctx.String("team"))
	if err != nil {
		return handleSelectError(err, "Team selection failed.")
	}

	var teamID *identity.ID
	if !newTeam {
		if org == nil {
			return errs.NewExitError("Team not found.")
		}
		teamID = team.ID
	}

	args := ctx.Args()
	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	name, err = promptForMachineName(name, teamName)
	fmt.Println()
	if err != nil {
		return errs.NewErrorExitError(machineCreateFailed, err)
	}

	if newOrg {
		org, err := client.Orgs.Create(c, orgName)
		if err != nil {
			return errs.NewErrorExitError("Could not create org", err)
		}

		orgID = org.ID
		err = generateKeypairsForOrg(c, ctx, client, org.ID, false)
		if err != nil {
			return err
		}

		fmt.Printf("Org %s created.\n\n", orgName)
	}

	if newTeam {
		team, err := client.Teams.Create(c, orgID, teamName, primitive.MachineTeam)
		if err != nil {
			return errs.NewErrorExitError("Could not create machine team", err)
		}

		teamID = team.ID
		fmt.Printf("Machine team %s created for org %s.\n\n", teamName, orgName)
	}

	machine, tokenSecret, err := createMachineByName(c, client, orgID, teamID, name)
	if err != nil {
		return err
	}

	fmt.Print("\nYou will only be shown the secret once, please keep it safe.\n\n")

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)

	tokenID := machine.Tokens[0].Token.ID
	fmt.Fprintf(w, "Machine ID:\t%s\n", machine.Machine.ID)
	fmt.Fprintf(w, "Machine Token ID:\t%s\n", tokenID)
	fmt.Fprintf(w, "Machine Token Secret:\t%s\n", tokenSecret)

	w.Flush()
	return err
}

func createMachineByName(c context.Context, client *api.Client,
	orgID, teamID *identity.ID, name string) (*apitypes.MachineSegment, *base64.Value, error) {

	machine, tokenSecret, err := client.Machines.Create(
		c, orgID, teamID, name, &progress)
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return nil, nil, errs.NewExitError("Machine already exists")
		}

		return nil, nil, errs.NewErrorExitError(
			"Could not create machine, please try again.", err)
	}

	return machine, tokenSecret, nil
}

func promptForMachineName(providedName, teamName string) (string, error) {
	defaultName, err := deriveMachineName(teamName)
	if err != nil {
		return "", errs.NewErrorExitError("Could not derive machine name", err)
	}

	name := ""
	if providedName == "" {
		name = defaultName
	} else {
		name = providedName
	}

	label := "Enter machine name"
	autoAccept := providedName != ""
	return NamePrompt(&label, name, autoAccept)
}

func deriveMachineName(teamName string) (string, error) {
	value := make([]byte, machineRandomIDLength)
	_, err := rand.Read(value)
	if err != nil {
		return "", err
	}

	name := teamName + "-" + base32.EncodeToString(value)
	return name, nil
}
