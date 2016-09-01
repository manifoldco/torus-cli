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
				Name:   "list",
				Usage:  "List organizations associated with your account",
				Action: Chain(EnsureDaemon, EnsureSession, orgsListCmd),
			},
		},
	}
	Cmds = append(Cmds, orgs)
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
