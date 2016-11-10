package cmd

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

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
		Name:      "machines",
		Usage:     "View and create machines within an organization",
		ArgsUsage: "<machine>",
		Category:  "ORGANIZATIONS",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create a machine for an organization",
				Flags: []cli.Flag{
					orgFlag("Org the machine will belong to", false),
					roleFlag("Role the machine will belong to", false),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					checkRequiredFlags, createMachine,
				),
			},
			{
				Name:      "view",
				Usage:     "Show the details of a machine",
				ArgsUsage: "<id|name>",
				Flags: []cli.Flag{
					orgFlag("Org the machine will belongs to", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					checkRequiredFlags, viewMachineCmd,
				),
			},
			{
				Name:      "destroy",
				Usage:     "Destroy a machine in the specified organization",
				ArgsUsage: "<id|name>",
				Flags: []cli.Flag{
					orgFlag("Org the machine will belongs to", true),
					stdAutoAcceptFlag,
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					checkRequiredFlags, destroyMachineCmd,
				),
			},
			{
				Name:  "list",
				Usage: "List machines for an organization",
				Flags: []cli.Flag{
					orgFlag("Org the machine belongs to", true),
					roleFlag("List machines of this role", false),
					destroyedFlag(),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					checkRequiredFlags, listMachinesCmd,
				),
			},
			{
				Name:      "roles",
				Usage:     "Lists and create machine roles for an organization",
				ArgsUsage: "<machine-role>",
				Subcommands: []cli.Command{
					{
						Name:  "list",
						Usage: "List all machine roles for an organization",
						Flags: []cli.Flag{
							orgFlag("Org the machine roles belongs to", true),
						},
						Action: chain(
							ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
							checkRequiredFlags, listMachineRoles,
						),
					},
					{
						Name:      "create",
						Usage:     "Create a machine role for an organization",
						ArgsUsage: "<name>",
						Flags: []cli.Flag{
							orgFlag("Org the machine role will belong to", true),
						},
						Action: chain(
							ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
							checkRequiredFlags, createMachineRole,
						),
					},
				},
			},
		},
	}
	Cmds = append(Cmds, machines)
}

func destroyMachineCmd(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 1 {
		return errs.NewUsageExitError("Too many arguments supplied.", ctx)
	}
	if len(args) < 1 {
		return errs.NewUsageExitError("Name or ID is required", ctx)
	}
	if ctx.String("org") == "" {
		return errs.NewUsageExitError("Missing flags: --org", ctx)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Look up the target org
	org, err := getOrg(c, client, ctx.String("org"))
	if err != nil {
		return errs.NewErrorExitError("Machine destroy failed", err)
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}

	machineID, err := identity.DecodeFromString(args[0])
	if err != nil {
		name := args[0]
		machines, lErr := client.Machines.List(c, org.ID, nil, &name, nil)
		if lErr != nil {
			return errs.NewErrorExitError("Failed to retrieve machine", err)
		}
		if len(machines) < 1 {
			return errs.NewExitError("Machine not found")
		}
		machineID = *machines[0].Machine.ID
	}

	preamble := "You are about to destroy a machine. This cannot be undone."
	abortErr := ConfirmDialogue(ctx, nil, &preamble, true)
	if abortErr != nil {
		return abortErr
	}

	err = client.Machines.Destroy(c, &machineID)
	if err != nil {
		return errs.NewErrorExitError("Failed to destroy machine", err)
	}

	fmt.Println("Machine destroyed.")
	return nil
}

func viewMachineCmd(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 1 {
		return errs.NewUsageExitError("Too many arguments supplied.", ctx)
	}
	if len(args) < 1 {
		return errs.NewUsageExitError("Name or ID is required", ctx)
	}
	if ctx.String("org") == "" {
		return errs.NewUsageExitError("Missing flags: --org", ctx)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Look up the target org
	org, err := getOrg(c, client, ctx.String("org"))
	if err != nil {
		return errs.NewErrorExitError("Machine view failed", err)
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}

	machineID, err := identity.DecodeFromString(args[0])
	if err != nil {
		name := args[0]
		machines, lErr := client.Machines.List(c, org.ID, nil, &name, nil)
		if lErr != nil {
			return errs.NewErrorExitError("Failed to retrieve machine", lErr)
		}
		if len(machines) < 1 {
			return errs.NewExitError("Machine not found")
		}
		machineID = *machines[0].Machine.ID
	}

	machineSegment, err := client.Machines.Get(c, &machineID)
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve machine", err)
	}
	if machineSegment == nil {
		return errs.NewExitError("Machine not found.")
	}

	orgTrees, err := client.Orgs.GetTree(c, *org.ID)
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve machine", err)
	}
	if len(orgTrees) < 1 {
		return errs.NewExitError("Machine metadata not found.")
	}
	orgTree := orgTrees[0]

	profileMap := make(map[identity.ID]apitypes.Profile, len(orgTree.Profiles))
	for _, p := range orgTree.Profiles {
		profileMap[*p.ID] = *p
	}

	teamMap := make(map[identity.ID]apitypes.Team, len(orgTree.Teams))
	for _, t := range orgTree.Teams {
		teamMap[*t.Team.ID] = *t.Team
	}

	machine := machineSegment.Machine
	machineBody := machine.Body

	// Created profile
	creator := profileMap[*machineBody.CreatedBy]
	createdBy := creator.Body.Username + " (" + creator.Body.Name + ")"
	createdOn := machineBody.Created.Format(time.RFC3339)

	// Destroyed profile
	destroyedOn := "-"
	destroyedBy := "-"
	if machineBody.State == primitive.MachineDestroyedState {
		destroyer := profileMap[*machineBody.DestroyedBy]
		destroyedOn = machineBody.Destroyed.Format(time.RFC3339)
		destroyedBy = destroyer.Body.Username + " (" + destroyer.Body.Name + ")"
	}

	// Membership info
	var teamNames []string
	for _, m := range machineSegment.Memberships {
		team := teamMap[*m.Body.TeamID]
		if team.Body.TeamType == primitive.MachineTeam {
			teamNames = append(teamNames, team.Body.Name)
		}
	}
	roleOutput := strings.Join(teamNames, ", ")
	if roleOutput == "" {
		roleOutput = "-"
	}

	fmt.Println("")
	w1 := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintf(w1, "ID:\t%s\n", machine.ID)
	fmt.Fprintf(w1, "Name:\t%s\n", machineBody.Name)
	fmt.Fprintf(w1, "Role:\t%s\n", roleOutput)
	fmt.Fprintf(w1, "State:\t%s\n", machineBody.State)
	fmt.Fprintf(w1, "Created By:\t%s\n", createdBy)
	fmt.Fprintf(w1, "Created On:\t%s\n", createdOn)
	fmt.Fprintf(w1, "Destroyed By:\t%s\n", destroyedBy)
	fmt.Fprintf(w1, "Destroyed On:\t%s\n", destroyedOn)
	w1.Flush()
	fmt.Println("")

	w2 := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintf(w2, "TOKEN ID\tSTATE\tCREATED BY\tCREATED ON\n")
	fmt.Fprintln(w2, " \t \t \t ")
	for _, token := range machineSegment.Tokens {
		tokenID := token.Token.ID
		state := token.Token.Body.State
		creator := profileMap[*token.Token.Body.CreatedBy]
		createdBy := creator.Body.Username + " (" + creator.Body.Name + ")"
		createdOn := token.Token.Body.Created.Format(time.RFC3339)
		fmt.Fprintf(w2, "%s\t%s\t%s\t%s\n", tokenID, state, createdBy, createdOn)
	}

	w2.Flush()
	fmt.Println("")

	return nil
}

func listMachinesCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	args := ctx.Args()
	if len(args) > 0 {
		return errs.NewUsageExitError("Too many arguments supplied.", ctx)
	}

	// Look up the target org
	org, err := client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve org", err)
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}
	orgID := org.ID

	state := primitive.MachineActiveState
	if ctx.Bool("destroyed") {
		state = primitive.MachineDestroyedState
	}

	if ctx.String("role") != "" && ctx.Bool("destroyed") {
		return errs.NewExitError(
			"Cannot specify --destroyed and --role at the same time")
	}

	teams, err := client.Teams.List(c, org.ID, ctx.String("role"), primitive.MachineTeam)
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve metadata", err)
	}
	if len(teams) < 1 {
		return errs.NewExitError("Machine roles not found")
	}

	var teamID *identity.ID
	if ctx.String("role") != "" {
		teamID = teams[0].ID
	}

	machines, err := client.Machines.List(c, orgID, &state, nil, teamID)
	if err != nil {
		return err
	}

	teamMap := make(map[identity.ID]primitive.Team, len(teams))
	for _, t := range teams {
		if t.Body.TeamType == primitive.MachineTeam {
			teamMap[*t.ID] = *t.Body
		}
	}

	fmt.Println("")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSTATE\tROLE\tCREATION DATE")
	fmt.Fprintln(w, " \t \t \t \t ")
	for _, machine := range machines {
		mID := machine.Machine.ID.String()
		m := machine.Machine.Body
		teamName := "-"
		for _, m := range machine.Memberships {
			team, ok := teamMap[*m.Body.TeamID]
			if ok {
				teamName = team.Name
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", mID, m.Name, m.State, teamName, m.Created.Format(time.RFC3339))
	}
	w.Flush()
	fmt.Println("")

	return nil
}

func listMachineRoles(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	args := ctx.Args()
	if len(args) > 0 {
		return errs.NewUsageExitError("Too many arguments supplied", ctx)
	}

	org, err := client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve org", err)
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}

	teams, err := client.Teams.List(c, org.ID, "", "")
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve roles", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 20, 0, 1, ' ', 0)
	for _, t := range teams {
		if !isMachineTeam(t.Body) {
			continue
		}

		displayTeamType := ""
		if t.Body.TeamType == primitive.SystemTeam && t.Body.Name == primitive.MachineTeamName {
			displayTeamType = "[system]"
		}

		fmt.Fprintf(w, "%s\t%s\n", t.Body.Name, displayTeamType)
	}

	w.Flush()
	fmt.Println("\nAll machines belong to the \"machine\" role.")
	return nil
}

func createMachineRole(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	args := ctx.Args()
	teamName := ""
	if len(args) > 0 {
		teamName = args[0]
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, oName, newOrg, err := SelectCreateOrg(c, client, ctx.String("org"))
	if err != nil {
		return handleSelectError(err, "Org selection failed")
	}
	if org == nil && !newOrg {
		fmt.Println("")
		return errs.NewExitError("Org not found.")
	}
	if newOrg && oName == "" {
		fmt.Println("")
		return errs.NewExitError("Invalid org name.")
	}

	var orgID *identity.ID
	if org != nil {
		orgID = org.ID
	}

	label := "Role name"
	autoAccept := teamName != ""
	teamName, err = NamePrompt(&label, teamName, autoAccept)
	if err != nil {
		return handleSelectError(err, "Role creation failed.")
	}

	if org == nil && newOrg {
		org, err = createOrgByName(c, ctx, client, oName)
		if err != nil {
			fmt.Println("")
			return err
		}

		orgID = org.ID
	}

	fmt.Println("")
	_, err = client.Teams.Create(c, orgID, teamName, primitive.MachineTeam)
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return errs.NewExitError("Role already exists")
		}

		return errs.NewErrorExitError("Role creation failed.", err)
	}

	fmt.Printf("Role %s created.\n", teamName)
	return nil
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

	team, teamName, newTeam, err := SelectCreateRole(c, client, orgID, ctx.String("role"))
	if err != nil {
		return handleSelectError(err, "Role selection failed.")
	}

	var teamID *identity.ID
	if !newTeam {
		if org == nil {
			return errs.NewExitError("Role not found.")
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
			return errs.NewErrorExitError("Could not create machine role", err)
		}

		teamID = team.ID
		fmt.Printf("Machine role %s created for org %s.\n\n", teamName, orgName)
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

	var name string
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
