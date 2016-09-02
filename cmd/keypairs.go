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
