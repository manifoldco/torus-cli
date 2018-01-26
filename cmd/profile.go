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
	"github.com/manifoldco/torus-cli/ui"

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

	printProfile(session)
	return nil
}

func printProfile(session *api.Session) {
	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	if session.Type() == apitypes.MachineSession {
		fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Machine ID"), session.ID())
		fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Machine Token ID"), session.AuthID())
		fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Machine Name"), ui.FaintString(session.Username()))
		fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Machine State"), colorizeAccountState(session))
	} else {
		fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Name"), session.Name())
		fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Username"), ui.FaintString(session.Username()))
		fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Email"), session.Email())
		fmt.Fprintf(w, "%s:\t%s\n", ui.BoldString("Status"), colorizeAccountState(session))
	}

	w.Flush()
}

func colorizeAccountState(session *api.Session) string {
	if session.Type() == apitypes.MachineSession {
		return colorizeMachineState(session.State())
	}

	return colorizeUserState(session.State())
}

func colorizeUserState(state string) string {
	switch state {
	case "active":
		return ui.ColorString(ui.Green, state)
	case "unverified":
		return ui.ColorString(ui.Yellow, state)
	default:
		return state
	}
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
	err = AskPerform("Would you like to change your password", "")
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
	mustVerify := false
	if ogEmail != email {
		delta.Email = email
		mustVerify = true
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
	abortErr := ConfirmDialogue(ctx, nil, &preamble, "", false)
	if abortErr != nil {
		return abortErr
	}

	// Update the profile
	result, err := client.Users.Update(c, delta)
	if err != nil {
		return errs.NewErrorExitError("Failed to update profile.", err)
	}
	currentEmail := result.Email()

	// If the password has changed, log the user in again
	if newPassword != "" {
		err = performLogin(c, client, currentEmail, newPassword, false)
		if err != nil {
			return err
		}
	}

	if mustVerify {
		fmt.Println("")
		fmt.Println("A verification code has been sent to your new email address.")
		fmt.Println("Please verify this change by entering the code below.")
		fmt.Println("")

		if err = askToVerify(ctx); err != nil {
			return err
		}
	}

	// Output the final session details
	updatedSession, err := client.Session.Who(c)
	if err != nil {
		return errs.NewErrorExitError("Error fetching user details", err)
	}

	fmt.Println("")

	printProfile(updatedSession)
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
