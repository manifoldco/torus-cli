package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
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
		return errs.NewExitError("Machines cannot update profiles")
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

	var newPassword string
	err = AskPerform(ctx, "Would you like to change your password?")
	if err == nil {
		password, err := changePassword(&c, client, session)
		if err != nil {
			return err
		}
		newPassword = password
	}

	// Don't perform any action if no changes occurred
	if ogEmail == email && ogName == name && newPassword == "" {
		fmt.Println("\nNo changes made :)")
		return nil
	}

	// Construct the update payload
	delta := apitypes.ProfileUpdate{}
	if ogEmail != email {
		delta.Email = email
	}
	if ogName != name {
		delta.Name = name
	}
	if newPassword != "" {
		delta.Password = newPassword
	}

	preamble := "\nYou are about to update your profile to the values entered above."
	if email != ogEmail {
		preamble = "\nYou will be required to re-verify your email address before taking any further actions within Torus."
	}
	abortErr := ConfirmDialogue(ctx, nil, &preamble, false)
	if abortErr != nil {
		return abortErr
	}

	// Update the profile
	_, err = client.Users.Update(c, delta)
	if err != nil {
		return errs.NewErrorExitError("Failed to update profile.", err)
	}

	// If the password has changed, log the user in again
	if newPassword != "" {
		err = performLogin(c, client, session.Email(), newPassword, false)
		if err != nil {
			return err
		}
	}

	// Output the final session details
	updatedSession, err := client.Session.Who(c)
	if err != nil {
		return errs.NewErrorExitError("Error fetching user details", err)
	}

	fmt.Println("")
	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Name:\t%s\n", updatedSession.Name())
	fmt.Fprintf(w, "Email:\t%s\n", updatedSession.Email())
	fmt.Fprintf(w, "Username:\t%s\n", updatedSession.Username())
	fmt.Fprintf(w, "Password:\t%s\n\n", strings.Repeat(string(PasswordMask), 10))
	w.Flush()

	return nil
}

func changePassword(c *context.Context, client *api.Client, session *api.Session) (string, error) {
	// Retrieve current password value
	oldLabel := "Current Password"
	currentPassword, err := PasswordPrompt(false, &oldLabel)
	if err != nil {
		return "", err
	}

	// Test the user's current password
	err = testLogin(*c, client, session.Email(), currentPassword)
	if err != nil {
		return "", errs.NewExitError("Invalid password.")
	}

	// Obtain new value for password
	newLabel := "New Password"
	return PasswordPrompt(true, &newLabel)
}
