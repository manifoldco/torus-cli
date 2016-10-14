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
					orgFlag("org to show keypairs for", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, listKeypairs,
				),
			},
			{
				Name:  "generate",
				Usage: "Generate keyparis for an organization",
				Flags: []cli.Flag{
					orgFlag("org to show keypairs for", true),
					cli.BoolFlag{
						Name:  "all",
						Usage: "Perform command for all orgs without valid keypairs",
					},
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, generateKeypairs,
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

func generateKeypairs(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	orgNames := make(map[*identity.ID]string)
	subjectOrgs := make(map[*identity.ID]*api.OrgResult)
	regenOrgs := make(map[*identity.ID]string)

	if ctx.Bool("all") {
		// If all flag is supplied, we will get all their orgs
		var orgs []api.OrgResult
		orgs, oErr := client.Orgs.List(c)
		if oErr != nil {
			return cli.NewExitError("Could not retrieve orgs, please try again", -1)
		}
		for _, org := range orgs {
			subjectOrgs[org.ID] = &org
			orgNames[org.ID] = org.Body.Name
		}
	} else {
		// Verify the org they've specified exists
		var org *api.OrgResult
		orgName := ctx.String("org")
		org, oErr := client.Orgs.GetByName(c, orgName)
		if oErr != nil || org == nil {
			return cli.NewExitError("Org '"+orgName+"' not found", -1)
		}
		subjectOrgs[org.ID] = org
		orgNames[org.ID] = org.Body.Name
	}

	// Iterate over target orgs and identify which keys exist
	var pErr error
	hasKey := make(map[string]map[string]bool, len(subjectOrgs))
	for orgID := range subjectOrgs {
		keypairs, err := client.Keypairs.List(c, orgID)
		if err != nil {
			pErr = err
			break
		}
		for _, kp := range keypairs {
			oID := kp.PublicKey.Body.OrgID.String()
			if hasKey[oID] == nil {
				hasKey[oID] = make(map[string]bool)
			}
			keyType := kp.PublicKey.Body.KeyType
			hasKey[oID][keyType] = true
		}
	}

	if pErr != nil {
		return cli.NewExitError("Error fetching required context", -1)
	}

	// Regenerate for orgs which do not have both keys present
	for orgID := range subjectOrgs {
		if !hasKey[orgID.String()]["encryption"] || !hasKey[orgID.String()]["signing"] {
			regenOrgs[orgID] = orgNames[orgID]
		}
	}

	var rErr error

	for orgID, name := range regenOrgs {
		fmt.Println("Generating signing and encryption keypairs for org: " + name)
		err := client.Keypairs.Generate(c, orgID, &progress)
		if err != nil && rErr == nil {
			rErr = err
			break
		}
	}

	if rErr != nil {
		return cli.NewExitError("Error while regenerating keypairs", -1)
	}

	if len(regenOrgs) > 0 {
		fmt.Println("Keypair generation successful.")
	} else {
		fmt.Println("No keypairs missing.")
	}
	return nil
}

func generateKeypairsForOrg(c context.Context, ctx *cli.Context, client *api.Client, orgID *identity.ID, lookupOrg bool) error {
	var err error

	msg := fmt.Sprintf("Could not generate keypairs for org. Run '%s keypairs generate' to fix.", ctx.App.Name)
	outputErr := cli.NewExitError(msg, -1)
	if orgID == nil && !lookupOrg {
		return outputErr
	}

	// Lookup org if not supplied
	if orgID == nil && lookupOrg {
		orgs, err := client.Orgs.List(c)
		if err != nil || len(orgs) < 1 {
			return outputErr
		}
		org := orgs[0]
		orgID = org.ID
	}

	err = client.Keypairs.Generate(c, orgID, &progress)
	if err != nil {
		return outputErr
	}

	return nil
}
