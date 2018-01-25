package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
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
					orgFlag("org to generate keypairs for", false),
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
			{
				Name:  "revoke",
				Usage: "Revoke the keypairs for an organization (used for testing only)",

				Flags: []cli.Flag{
					orgFlag("org to revoke keypairs for", true),
				},
				Action: chain(
					ensureDaemon, ensureSession, loadDirPrefs, loadPrefDefaults,
					setUserEnv, checkRequiredFlags, revokeKeypairs,
				),
				Hidden: true,
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
	var org *envelope.Org
	org, err = client.Orgs.GetByName(c, ctx.String("org"))
	if err != nil {
		return errs.NewExitError(keypairListFailed)
	}
	if org == nil {
		return errs.NewExitError("Org not found.")
	}

	keypairs, err := client.KeyPairs.List(c, org.ID)
	if err != nil {
		return errs.NewExitError(keypairListFailed)
	}

	fmt.Println("")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tORG\tKEY TYPE\tVALID\tCREATION DATE")
	fmt.Fprintln(w, " \t \t \t \t ")
	for _, keypair := range keypairs.All() {
		pk := keypair.PublicKey.Body
		valid := "YES"
		if keypair.Revoked() {
			valid = "NO"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", keypair.PublicKey.ID,
			org.Body.Name, pk.KeyType, valid, pk.Created.Format(time.RFC3339))
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
	subjectOrgs := make(map[*identity.ID]*envelope.Org)
	regenOrgs := make(map[*identity.ID]string)

	if ctx.Bool("all") {
		// If all flag is supplied, we will get all their orgs
		var orgs []envelope.Org
		orgs, oErr := client.Orgs.List(c)
		if oErr != nil {
			return errs.NewExitError("Could not retrieve orgs, please try again.")
		}
		for _, org := range orgs {
			subjectOrgs[org.ID] = &org
			orgNames[org.ID] = org.Body.Name
		}
	} else {
		// Verify the org they've specified exists
		var org *envelope.Org
		orgName := ctx.String("org")
		if orgName == "" {
			return errs.NewExitError("Missing flags: --org.")
		}
		org, oErr := client.Orgs.GetByName(c, orgName)
		if oErr != nil || org == nil {
			return errs.NewExitError("Org '" + orgName + "' not found.")
		}
		subjectOrgs[org.ID] = org
		orgNames[org.ID] = org.Body.Name
	}

	// Iterate over target orgs and identify which keys exist
	var pErr error
	hasKey := make(map[identity.ID]map[primitive.KeyType]bool, len(subjectOrgs))
	for orgID := range subjectOrgs {
		keypairs, err := client.KeyPairs.List(c, orgID)
		if err != nil {
			pErr = err
			break
		}
		for _, kp := range keypairs.All() {
			if kp.Revoked() {
				continue
			}

			oID := *kp.PublicKey.Body.OrgID
			if hasKey[oID] == nil {
				hasKey[oID] = make(map[primitive.KeyType]bool)
			}
			keyType := kp.PublicKey.Body.KeyType
			hasKey[oID][keyType] = true
		}
	}

	if pErr != nil {
		return errs.NewExitError("Error fetching required context.")
	}

	// Regenerate for orgs which do not have both keys present
	for orgID := range subjectOrgs {
		if !hasKey[*orgID][primitive.EncryptionKeyType] || !hasKey[*orgID][primitive.SigningKeyType] {
			regenOrgs[orgID] = orgNames[orgID]
		}
	}

	var rErr error

	s, p := spinner("Attempting to generate keypairs")
	s.Start()
	for orgID, name := range regenOrgs {
		s.Update("Generating signing and encryption keypairs for org: " + name)
		err := client.KeyPairs.Create(c, orgID, p)
		if err != nil && rErr == nil {
			s.Stop()
			rErr = err
			break
		}
	}
	s.Stop()

	if rErr != nil {
		return errs.NewExitError("Error while regenerating keypairs.")
	}

	if len(regenOrgs) > 0 {
		for _, name := range regenOrgs {
			fmt.Printf("Successfully generated keypairs for %s org\n", name)
		}
	} else {
		fmt.Println("No keypairs missing.")
	}
	return nil
}

func generateKeypairsForOrg(c context.Context, ctx *cli.Context, client *api.Client, orgID *identity.ID, lookupOrg bool) error {
	var err error

	msg := fmt.Sprintf("Could not generate keypairs for org. Run '%s keypairs generate' to fix.", ctx.App.Name)
	outputErr := errs.NewExitError(msg)
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

	s, p := spinner("Attempting to generate keypairs")
	s.Start()
	err = client.KeyPairs.Create(c, orgID, p)
	s.Stop()
	if err != nil {
		return outputErr
	}

	return nil
}

func revokeKeypairs(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	// Verify the org they've specified exists
	orgName := ctx.String("org")
	if orgName == "" {
		return errs.NewExitError("Missing flags: --org.")
	}
	org, err := client.Orgs.GetByName(c, orgName)
	if err != nil || org == nil {
		return errs.NewExitError("Org '" + orgName + "' not found.")
	}

	// Iterate over target orgs and identify which keys exist
	keypairs, err := client.KeyPairs.List(c, org.ID)
	if err != nil {
		return errs.NewErrorExitError("Error fetching keypairs.", err)
	}

	hasKey := make(map[primitive.KeyType]struct{})

	for _, kp := range keypairs.All() {
		if kp.Revoked() {
			continue
		}

		hasKey[kp.PublicKey.Body.KeyType] = struct{}{}
	}

	if _, ok := hasKey[primitive.SigningKeyType]; !ok {
		fmt.Println("No existing keys to revoke.")
		return nil
	}

	s, p := spinner("Attempting to revoke keypairs")
	s.Start()
	err = client.KeyPairs.Revoke(c, org.ID, p)
	s.Stop()
	if err != nil {
		return errs.NewErrorExitError("Error while revoking keypairs.", err)
	}

	fmt.Println("Keypairs revoked.")
	return nil
}
