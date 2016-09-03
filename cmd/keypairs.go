package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/identity"
)

func init() {
	keypairs := cli.Command{
		Name:     "keypairs",
		Usage:    "View and generate organization keypairs",
		Category: "ORGANIZATIONS",
		Subcommands: []cli.Command{
			{
				Name:  "list",
				Usage: "List your keypairs for an organization",
				Flags: []cli.Flag{
					OrgFlag("org to show keypairs for", true),
				},
				Action: Chain(
					EnsureDaemon, EnsureSession, LoadDirPrefs, LoadPrefDefaults,
					SetUserEnv, checkRequiredFlags, listKeypairs,
				),
			},
		},
	}
	Cmds = append(Cmds, keypairs)
}

const keypairListFailed = "Could not list keypairs, please try again."

func listKeypairs(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Look up the target org
	var org *api.OrgResult
	org, err = client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return cli.NewExitError(keypairListFailed, -1)
	}
	if org == nil {
		return cli.NewExitError("Org not found.", -1)
	}

	keypairs, err := client.Keypairs.List(c, org.ID)
	if err != nil {
		return cli.NewExitError(keypairListFailed, -1)
	}

	fmt.Println("")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintln(w, "ID\tORG\tKEY TYPE\tVALID\tCREATION DATE")
	fmt.Fprintln(w, " \t \t \t \t ")
	for _, keypair := range keypairs {
		fmt.Fprintln(w, keypair.PublicKey.ID.String()+"\t"+org.Body.Name+"\t"+keypair.PublicKey.Body.KeyType+"\tYES\t"+keypair.PublicKey.Body.Created.Format(time.RFC3339))
	}
	w.Flush()
	fmt.Println("")

	return nil
}

func generateKeypairsForOrg(ctx *cli.Context, c context.Context, client *api.Client, targetOrg *api.OrgResult, lookupOrg bool) error {
	var err error
	var progress api.ProgressFunc = func(evt *api.Event, err error) {
		if evt != nil {
			fmt.Println(evt.Message)
		}
	}

	msg := fmt.Sprintf("Could not generate keypairs for org. Run '%s keypairs generate' to fix.", ctx.App.Name)
	outputErr := cli.NewExitError(msg, -1)
	if targetOrg == nil && lookupOrg == false {
		return outputErr
	}

	// Lookup org if not supplied
	var orgID *identity.ID
	if targetOrg == nil && lookupOrg == true {
		orgs, err := client.Orgs.List(c)
		if err != nil || len(orgs) < 1 {
			return outputErr
		}
		org := orgs[0]
		orgID = org.ID
	} else {
		orgID = targetOrg.ID
	}

	err = client.Keypairs.Generate(c, orgID, &progress)
	if err != nil {
		return outputErr
	}

	return nil
}
