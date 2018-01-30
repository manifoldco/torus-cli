package cmd

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/juju/ansiterm"
	"github.com/manifoldco/go-base32"
	"github.com/manifoldco/go-base64"
	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/gatekeeper/bootstrap"
	"github.com/manifoldco/torus-cli/hints"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
	"github.com/manifoldco/torus-cli/prompts"
	"github.com/manifoldco/torus-cli/ui"
)

const (
	machineRandomIDLength = 5 // 8 characters in base32
	machineCreateFailed   = "Could not create machine, please try again."

	// GlobalRoot is the global root of the Torus config
	GlobalRoot = "/etc/torus"

	// EnvironmentFile is the environment file that stores machine information
	EnvironmentFile = "token.environment"
)

// urlFlag creates a new --bootstrap cli.Flag
func urlFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("url, u", "URL", usage, "", "TORUS_BOOTSTRAP_URL", required)
}

// authProviderFlag creates a new --auth cli.Flag
func authProviderFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("auth, a", "AUTHPROVIDER", usage, "", "TORUS_AUTH_PROVIDER", required)
}

func caFlag(usage string, required bool) cli.Flag {
	return newPlaceholder("ca", "CA_BUNDLE", usage, "", "TORUS_BOOTSTRAP_CA", required)
}

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
				Name:  "list",
				Usage: "List machines for an organization",
				Flags: []cli.Flag{
					orgFlag("Org the machine belongs to", false),
					roleFlag("List machines of this role", false),
					destroyedFlag(),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					checkRequiredFlags, listMachinesCmd,
				),
			},
			{
				Name:      "view",
				Usage:     "Show the details of a machine",
				ArgsUsage: "<id|name>",
				Flags: []cli.Flag{
					orgFlag("Org the machine will belongs to", false),
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
					orgFlag("Org the machine will belongs to", false),
					stdAutoAcceptFlag,
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					checkRequiredFlags, destroyMachineCmd,
				),
			},
			{
				Name:      "roles",
				Usage:     "Lists and create machine roles for an organization",
				ArgsUsage: "<machine-role>",
				Subcommands: []cli.Command{
					{
						Name:      "create",
						Usage:     "Create a machine role for an organization",
						ArgsUsage: "<name>",
						Flags: []cli.Flag{
							orgFlag("Org the machine role will belong to", false),
						},
						Action: chain(
							ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
							checkRequiredFlags, createMachineRole,
						),
					},
					{
						Name:  "list",
						Usage: "List all machine roles for an organization",
						Flags: []cli.Flag{
							orgFlag("Org the machine roles belongs to", false),
						},
						Action: chain(
							ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
							checkRequiredFlags, listMachineRoles,
						),
					},
				},
			},
			{
				Name:  "bootstrap",
				Usage: "Bootstrap a new machine using Torus Gatekeeper",
				Flags: []cli.Flag{
					authProviderFlag("Auth provider for bootstrapping", true),
					urlFlag("Gatekeeper URL for bootstrapping", true),
					roleFlag("Role the machine will belong to", true),
					machineFlag("Machine name to bootstrap", false),
					orgFlag("Org the machine will belong to", false),
					caFlag("CA Bundle to use for certificate verification. Uses system if none is provided", false),
				},
				Action: chain(checkRequiredFlags, bootstrapCmd),
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
	success, err := prompts.Confirm(nil, &preamble, true, true)
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve confirmation", err)
	}
	if !success {
		return errs.ErrAbort
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

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, _, _, err := selectOrg(c, client, ctx.String("org"), false)
	if err != nil {
		return err
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

	teamMap := make(map[identity.ID]envelope.Team, len(orgTree.Teams))
	for _, t := range orgTree.Teams {
		teamMap[*t.Team.ID] = *t.Team
	}

	machine := machineSegment.Machine
	machineBody := machine.Body

	// Created profile
	creator := profileMap[*machineBody.CreatedBy]
	createdBy := creator.Body.Name + " (" + ui.FaintString(creator.Body.Username) + ")"
	createdOn := machineBody.Created.Format(time.RFC3339)

	// Destroyed profile
	destroyedOn := "-"
	destroyedBy := "-"
	if machineBody.State == primitive.MachineDestroyedState {
		destroyer := profileMap[*machineBody.DestroyedBy]
		destroyedOn = machineBody.Destroyed.Format(time.RFC3339)
		destroyedBy = destroyer.Body.Name + " (" + ui.FaintString(destroyer.Body.Name) + ")"
	}

	// Membership info
	var teamNames []string
	for _, m := range machineSegment.Memberships {
		team := teamMap[*m.Body.TeamID]
		if team.Body.TeamType == primitive.MachineTeamType {
			teamNames = append(teamNames, team.Body.Name)
		}
	}
	roleOutput := strings.Join(teamNames, ", ")
	if roleOutput == "" {
		roleOutput = "-"
	}

	fmt.Println("")
	w1 := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintf(w1, "%s:\t%s\n", ui.BoldString("ID"), machine.ID)
	fmt.Fprintf(w1, "%s:\t%s\n", ui.BoldString("Name"), ui.FaintString(machineBody.Name))
	fmt.Fprintf(w1, "%s:\t%s\n", ui.BoldString("Role"), roleOutput)
	fmt.Fprintf(w1, "%s:\t%s\n", ui.BoldString("State"), colorizeMachineState(machineBody.State))
	fmt.Fprintf(w1, "%s:\t%s\n", ui.BoldString("Created By"), createdBy)
	fmt.Fprintf(w1, "%s:\t%s\n", ui.BoldString("Created On"), createdOn)
	fmt.Fprintf(w1, "%s:\t%s\n", ui.BoldString("Destroyed By"), destroyedBy)
	fmt.Fprintf(w1, "%s:\t%s\n", ui.BoldString("Destroyed On"), destroyedOn)
	w1.Flush()
	fmt.Println("")

	w2 := ansiterm.NewTabWriter(os.Stdout, 2, 0, 3, ' ', 0)
	fmt.Fprintf(w2, "%s\t%s\t%s\t%s\n", ui.BoldString("Token ID"), ui.BoldString("State"), ui.BoldString("Created By"), ui.BoldString("Created On"))
	for _, token := range machineSegment.Tokens {
		tokenID := token.Token.ID
		state := colorizeMachineState(token.Token.Body.State)
		creator := profileMap[*token.Token.Body.CreatedBy]
		createdBy := creator.Body.Name + " (" + ui.FaintString(creator.Body.Username) + ")"
		createdOn := token.Token.Body.Created.Format(time.RFC3339)
		fmt.Fprintf(w2, "%s\t%s\t%s\t%s\n", tokenID, state, createdBy, createdOn)
	}

	w2.Flush()
	fmt.Println("")

	return nil
}

func colorizeMachineState(state string) string {
	switch state {
	case "active":
		return ui.ColorString(ui.Green, state)
	case "destroyed":
		return ui.ColorString(ui.Red, state)
	default:
		return state
	}
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

	org, _, _, err := selectOrg(c, client, ctx.String("org"), false)
	if err != nil {
		return err
	}

	state := primitive.MachineActiveState
	if ctx.Bool("destroyed") {
		state = primitive.MachineDestroyedState
	}

	if ctx.String("role") != "" && ctx.Bool("destroyed") {
		return errs.NewExitError(
			"Cannot specify --destroyed and --role at the same time")
	}

	roles, err := client.Teams.List(c, org.ID, ctx.String("role"), primitive.MachineTeamType)
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve metadata", err)
	}

	// If no role is given, we don't want to error for role not found when there
	// are no roles at all. instead we want to error with no machines found.
	if len(roles) < 1 && ctx.String("role") != "" {
		return errs.NewExitError("Machine role not found.")
	}

	var roleID *identity.ID
	if ctx.String("role") != "" {
		roleID = roles[0].ID
	}

	machines, err := client.Machines.List(c, org.ID, &state, nil, roleID)
	if err != nil {
		return err
	}

	if len(machines) == 0 {
		fmt.Println("No machines found.")
		return nil
	}

	roleMap := make(map[identity.ID]primitive.Team, len(roles))
	for _, t := range roles {
		if t.Body.TeamType == primitive.MachineTeamType {
			roleMap[*t.ID] = *t.Body
		}
	}

	fmt.Println("")
	w := ansiterm.NewTabWriter(os.Stdout, 2, 0, 3, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", ui.BoldString("ID"), ui.BoldString("Name"), ui.BoldString("State"), ui.BoldString("Role"), ui.BoldString("Creation Date"))
	for _, machine := range machines {
		mID := machine.Machine.ID.String()
		m := machine.Machine.Body
		roleName := "-"
		for _, m := range machine.Memberships {
			role, ok := roleMap[*m.Body.TeamID]
			if ok {
				roleName = role.Name
			}
		}

		state := colorizeMachineState(state)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", mID, ui.FaintString(m.Name), state, roleName, m.Created.Format(time.RFC3339))
	}
	w.Flush()
	fmt.Println("")

	return nil
}

// bootstrapCmd is the cli.Command for Bootstrapping machine configuration from the Gatekeeper
func bootstrapCmd(ctx *cli.Context) error {
	cloud := ctx.String("auth")

	resp, err := bootstrap.Do(
		bootstrap.Provider(cloud),
		ctx.String("url"),
		ctx.String("machine"),
		ctx.String("org"),
		ctx.String("role"),
		ctx.String("ca"),
	)
	if err != nil {
		return fmt.Errorf("bootstrap provision failed: %s", err)
	}

	envFile := filepath.Join(GlobalRoot, EnvironmentFile)
	err = writeEnvironmentFile(resp.Token, resp.Secret)
	if err != nil {
		return fmt.Errorf("failed to write environment file[%s]: %s", envFile, err)
	}

	fmt.Printf("Machine bootstrapped. Environment configuration saved in %s\n", envFile)
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

	teams, err := client.Teams.List(c, org.ID, "", primitive.AnyTeamType)
	if err != nil {
		return errs.NewErrorExitError("Failed to retrieve roles", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 20, 0, 1, ' ', 0)
	for _, t := range teams {
		if !isMachineTeam(t.Body) {
			continue
		}

		displayTeamType := ""
		if t.Body.TeamType == primitive.SystemTeamType && t.Body.Name == primitive.MachineTeamName {
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

	org, oName, _, err := selectOrg(c, client, ctx.String("org"), false)
	if err != nil {
		return err
	}

	orgID := org.ID
	teamName, err = prompts.RoleName(teamName, true)
	if err != nil {
		return err
	}

	fmt.Println("")
	_, err = client.Teams.Create(c, orgID, teamName, primitive.MachineTeamType)
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return errs.NewExitError("Role already exists")
		}

		return errs.NewErrorExitError("Role creation failed.", err)
	}

	fmt.Printf("Machine role %s created for org %s.\n", teamName, oName)
	hints.Display(hints.Allow, hints.Deny, hints.Policies)
	return nil
}

func createMachine(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, oName, _, err := selectOrg(c, client, ctx.String("org"), false)
	if err != nil {
		return err
	}

	role, roleName, newRole, err := selectRole(c, client, org, ctx.String("role"), true)
	if err != nil {
		return err
	}

	args := ctx.Args()
	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	name, err = promptForMachineName(name, roleName)
	fmt.Println()
	if err != nil {
		return errs.NewErrorExitError(machineCreateFailed, err)
	}

	if newRole {
		role, err = client.Teams.Create(c, org.ID, roleName, primitive.MachineTeamType)
		if err != nil {
			return errs.NewErrorExitError("Could not create machine role", err)
		}

		fmt.Printf("Machine role %s created for org %s.\n\n", roleName, oName)
	}

	machine, tokenSecret, err := createMachineByName(c, client, org.ID, role.ID, name)
	if err != nil {
		return err
	}

	fmt.Print("\nYou will only be shown the secret once, please keep it safe.\n\n")

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)

	tokenID := machine.Tokens[0].Token.ID
	fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Machine ID"), machine.Machine.ID)
	fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Machine Token ID"), tokenID)
	fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Machine Token Secret"), tokenSecret)

	w.Flush()
	hints.Display(hints.Allow, hints.Deny)
	return nil
}

func createMachineByName(c context.Context, client *api.Client,
	orgID, teamID *identity.ID, name string) (*apitypes.MachineSegment, *base64.Value, error) {

	s, p := spinner("Attempting to create machine.")
	s.Start()
	machine, tokenSecret, err := client.Machines.Create(
		c, orgID, teamID, name, p)
	s.Stop()
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

	return prompts.MachineName(name, false)
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

func writeEnvironmentFile(token *identity.ID, secret *base64.Value) error {
	_, err := os.Stat(GlobalRoot)
	if os.IsNotExist(err) {
		os.Mkdir(GlobalRoot, 0700)
	}

	envPath := filepath.Join(GlobalRoot, EnvironmentFile)
	f, err := os.Create(envPath)
	if err != nil {
		return err
	}
	os.Chmod(envPath, 0600)

	w := bufio.NewWriter(f)
	w.WriteString(fmt.Sprintf("TORUS_TOKEN_ID=%s\n", token))
	w.WriteString(fmt.Sprintf("TORUS_TOKEN_SECRET=%s\n", secret))
	w.Flush()

	return nil
}
