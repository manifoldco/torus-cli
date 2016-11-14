package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"

	"github.com/urfave/cli"
)

func init() {
	profile := cli.Command{
		Name:     "profile",
		Usage:    "Manage your Torus account",
		Category: "ACCOUNT",
		Subcommands: []cli.Command{
			{
				Name:  "view",
				Usage: "View your profile",
				Action: chain(
					ensureDaemon, ensureSession, setUserEnv, profileView,
				),
			},
			{
				Name:  "update",
				Usage: "Update your profile",
				Action: chain(
					ensureDaemon, ensureSession, setUserEnv, profileEdit,
				),
			},
		},
	}
	Cmds = append(Cmds, profile)
}

// profileView is used to view your account profile
func profileView(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	session, err := client.Session.Who(c)
	if err != nil {
		return errs.NewErrorExitError("Error fetching user details", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	if session.Type() == apitypes.MachineSession {
		fmt.Fprintf(w, "Machine ID:\t%s\n", session.ID())
		fmt.Fprintf(w, "Machine Token ID:\t%s\n", session.AuthID())
		fmt.Fprintf(w, "Machine Name:\t%s\n\n", session.Username())
	} else {
		fmt.Fprintf(w, "Name:\t%s\n", session.Name())
		fmt.Fprintf(w, "Email:\t%s\n", session.Email())
		fmt.Fprintf(w, "Username:\t%s\n\n", session.Username())
	}

	w.Flush()

	return nil
}

// profileEdit is used to update name and email for an account
func profileEdit(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	session, err := client.Session.Who(c)
	if err != nil {
		return errs.NewErrorExitError("Error fetching user details", err)
	}
	if session.Type() == apitypes.MachineSession {
		return errs.NewExitError("Machines cannot update profile")
	}

	ogName := session.Name()
	name, err := FullNamePrompt(ogName)
	if err != nil {
		return err
	}

	ogEmail := session.Email()
	email, err := EmailPrompt(ogEmail)
	if err != nil {
		return err
	}

	warning := "\nYou are about to update your profile to the values above."
	if email != ogEmail {
		warning = "\nYou will be required to re-verify your email address before taking any further actions within Torus."
	}

	if ogEmail == email && ogName == name {
		fmt.Println("\nNo changes made :)")
		return nil
	}

	err = ConfirmDialogue(ctx, nil, &warning)
	if err != nil {
		return err
	}

	delta := apitypes.ProfileUpdate{}
	if ogEmail != email {
		delta.Email = email
	}
	if ogName != name {
		delta.Name = name
	}

	_, err = client.Users.Update(c, delta)
	if err != nil {
		return errs.NewErrorExitError("Failed to update profile.", err)
	}
	updatedSession, err := client.Session.Who(c)
	if err != nil {
		return errs.NewErrorExitError("Error fetching user details", err)
	}

	fmt.Println("")
	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Name:\t%s\n", updatedSession.Name())
	fmt.Fprintf(w, "Email:\t%s\n", updatedSession.Email())
	fmt.Fprintf(w, "Username:\t%s\n\n", updatedSession.Username())
	w.Flush()

	return nil
}
