package cmd

import (
	"context"
	"fmt"
	"sync"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
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
				Action:    Chain(EnsureDaemon, orgsCreate),
			},
			{
				Name:   "list",
				Usage:  "List organizations associated with your account",
				Action: Chain(EnsureDaemon, EnsureSession, orgsListCmd),
			},
		},
	}
	Cmds = append(Cmds, orgs)
}

const orgCreateFailed = "Org creation failed, please try again."

func orgsCreate(ctx *cli.Context) error {
	args := ctx.Args()
	usage := usageString(ctx)
	if len(args) > 1 {
		text := "Too many arguments\n\n"
		text += usage
		return cli.NewExitError(text, -1)
	}
	if len(args) < 1 {
		text := "Missing name\n\n"
		text += usage
		return cli.NewExitError(text, -1)
	}

	c := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		return cli.NewExitError(orgCreateFailed, -1)
	}

	client := api.NewClient(cfg)

	_, err = createOrgByName(ctx, c, client, args[0])
	if err != nil {
		return err
	}

	return nil
}

func createOrgByName(ctx *cli.Context, c context.Context, client *api.Client, name string) (*api.OrgResult, error) {
	org, err := client.Orgs.Create(c, name)
	if err != nil {
		return nil, cli.NewExitError(orgCreateFailed, -1)
	}

	err = generateKeypairsForOrg(ctx, c, client, org.ID, false)
	if err != nil {
		msg := fmt.Sprintf("Could not generate keypairs for org. Run '%s keypairs generate' to fix.", ctx.App.Name)
		return nil, cli.NewExitError(msg, -1)
	}

	fmt.Println("Org " + org.Body.Name + " created.")
	return org, nil
}

func orgsListCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	var wg sync.WaitGroup
	wg.Add(2)

	var orgs []api.OrgResult
	var self *api.UserResult
	var oErr, sErr error

	go func() {
		orgs, oErr = client.Orgs.List(context.Background())
		wg.Done()
	}()

	go func() {
		self, sErr = client.Users.Self(context.Background())
		wg.Done()
	}()

	wg.Wait()
	if oErr != nil || sErr != nil {
		return cli.NewExitError("Error fetching orgs list", -1)
	}

	withoutPersonal := orgs
	for i, o := range orgs {
		if o.Body.Name == self.Body.Username {
			fmt.Printf("  %s [personal]\n", o.Body.Name)
			withoutPersonal = append(orgs[:i], orgs[i+1:]...)
		}
	}

	for _, o := range withoutPersonal {
		fmt.Printf("  %s\n", o.Body.Name)
	}

	return nil
}
