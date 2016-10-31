package cmd

import (
	"context"
	"fmt"
	"sync"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/identity"
)

func init() {
	orgs := cli.Command{
		Name:     "orgs",
		Usage:    "View and create organizations",
		Category: "ORGANIZATIONS",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new organization",
				ArgsUsage: "<name>",
				Action:    chain(ensureDaemon, orgsCreate),
			},
			{
				Name:   "list",
				Usage:  "List organizations associated with your account",
				Action: chain(ensureDaemon, ensureSession, orgsListCmd),
			},
			{
				Name:      "remove",
				Usage:     "Remove a user from an org",
				ArgsUsage: "<username>",
				Flags: []cli.Flag{
					orgFlag("org to remove the user from", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, orgsRemove,
				),
			},
		},
	}
	Cmds = append(Cmds, orgs)
}

const orgCreateFailed = "Org creation failed, please try again."

func orgsCreate(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 1 {
		return errs.NewUsageExitError("Too many arguments", ctx)
	}

	var name string
	var err error

	if len(args) == 1 {
		name = args[0]
	}

	label := "Org name"
	autoAccept := name != ""
	name, err = NamePrompt(&label, name, autoAccept)
	if err != nil {
		return handleSelectError(err, orgCreateFailed)
	}

	c := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		return errs.NewExitError(orgCreateFailed)
	}

	client := api.NewClient(cfg)

	_, err = createOrgByName(c, ctx, client, name)
	return err
}

func createOrgByName(c context.Context, ctx *cli.Context, client *api.Client, name string) (*api.OrgResult, error) {
	org, err := client.Orgs.Create(c, name)
	if err != nil {
		return nil, errs.NewExitError(orgCreateFailed)
	}

	err = generateKeypairsForOrg(c, ctx, client, org.ID, false)
	if err != nil {
		msg := fmt.Sprintf("Could not generate keypairs for org. Run '%s keypairs generate' to fix.", ctx.App.Name)
		return nil, errs.NewExitError(msg)
	}

	fmt.Println("Org " + org.Body.Name + " created.")
	return org, nil
}

func orgsListCmd(ctx *cli.Context) error {
	orgs, session, err := orgsList()
	if err != nil {
		return err
	}

	withoutPersonal := orgs

	if session.Type() == apitypes.UserSession {
		for i, o := range orgs {
			if o.Body.Name == session.Username() {
				fmt.Printf("  %s [personal]\n", o.Body.Name)
				withoutPersonal = append(orgs[:i], orgs[i+1:]...)
			}
		}
	}

	for _, o := range withoutPersonal {
		fmt.Printf("  %s\n", o.Body.Name)
	}

	return nil
}

func orgsList() ([]api.OrgResult, api.Session, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	client := api.NewClient(cfg)

	var wg sync.WaitGroup
	wg.Add(2)

	var orgs []api.OrgResult
	var session api.Session
	var oErr, sErr error

	go func() {
		orgs, oErr = client.Orgs.List(context.Background())
		wg.Done()
	}()

	go func() {
		session, sErr = client.Session.Who(context.Background())
		wg.Done()
	}()

	wg.Wait()
	if oErr != nil || sErr != nil {
		return nil, nil, errs.NewExitError("Error fetching orgs list")
	}

	return orgs, session, nil
}

func orgsRemove(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) < 1 || args[0] == "" {
		return errs.NewUsageExitError("Missing username", ctx)
	}
	if len(args) > 1 {
		return errs.NewUsageExitError("Too many arguments", ctx)
	}
	username := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	const userNotFound = "User not found."
	const orgsRemoveFailed = "Could remove user from the org. Please try again."

	org, err := client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return errs.NewExitError(orgsRemoveFailed)
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}

	profile, err := client.Profiles.ListByName(c, username)
	if apitypes.IsNotFoundError(err) {
		return errs.NewExitError(userNotFound)
	}
	if err != nil {
		return errs.NewExitError(orgsRemoveFailed)
	}
	if profile == nil {
		return errs.NewExitError(userNotFound)
	}

	err = client.Orgs.RemoveMember(c, *org.ID, *profile.ID)
	if apitypes.IsNotFoundError(err) {
		fmt.Println("User is not a member of the org.")
		return nil
	}
	if err != nil {
		return errs.NewExitError(orgsRemoveFailed)
	}

	fmt.Println("User has been removed from the org.")
	return nil
}

func getOrg(ctx context.Context, client *api.Client, name string) (*api.OrgResult, error) {
	org, err := client.Orgs.GetByName(ctx, name)
	if err != nil {
		return nil, errs.NewExitError("Unable to lookup org. Please try again.")
	}
	if org == nil {
		return nil, errs.NewExitError("Org not found.")
	}

	return org, nil
}

func getOrgTree(c context.Context, client *api.Client, orgID identity.ID) ([]api.OrgTreeSegment, error) {
	orgTree, err := client.Orgs.GetTree(c, orgID)
	if err != nil {
		return nil, errs.NewExitError("Unable to lookup org tree. Please try again.")
	}
	if orgTree == nil {
		return nil, errs.NewExitError("Org tree not found.")
	}
	return orgTree, nil
}
