package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/errs"
	"github.com/manifoldco/torus-cli/hints"
	"github.com/manifoldco/torus-cli/ui"

	"github.com/urfave/cli"
)

func init() {
	signup := cli.Command{
		Name:     "signup",
		Usage:    "Create a new Torus account",
		Category: "ACCOUNT",
		Action:   chain(ensureDaemon, signupCmd),
	}
	Cmds = append(Cmds, signup)
}

func signupCmd(ctx *cli.Context) error {
	return signup(ctx, false)
}

// signup can be ran as a sub-command when an account is needed prior to running
// a particular action. the subCommand boolean signifies it is running as such
// and not as a generic signup
func signup(ctx *cli.Context, subCommand bool) error {
	args := ctx.Args()
	// Arguments are only used as a subcommand
	if subCommand && (len(args) > 0 && len(args) != 2) {
		var text string
		if len(args) > 2 {
			text = "Too many arguments supplied."
		} else {
			text = "Too few arguments supplied."
		}
		return errs.NewUsageExitError(text, ctx)
	}

	fmt.Println("By completing sign up, you agree to our terms of use (found at https://torus.sh/terms)\nand our privacy policy (found at https://torus.sh/privacy)")
	fmt.Println("")

	name, err := FullNamePrompt("")
	if err != nil {
		return err
	}

	username, err := UsernamePrompt("")
	if err != nil {
		return err
	}

	defaultEmail := ""
	defaultInvite := ""
	if len(args) == 2 {
		defaultEmail = args[0]
		defaultInvite = args[1]
	}

	email, err := EmailPrompt(defaultEmail)
	if err != nil {
		return err
	}

	var inviteCode string
	if subCommand {
		inviteCode, err = InviteCodePrompt(defaultInvite)
		if err != nil {
			return err
		}
	}

	reminderLabel := "Reminder: "
	ui.Hint("Don't forget to keep your password safe and secure! You can't recover your account if your password lost.", false, &reminderLabel)
	password, err := PasswordPrompt(true, nil)
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	signup := apitypes.Signup{
		Name:       name,
		Username:   username,
		Passphrase: password,
		Email:      email,
		InviteCode: inviteCode,
		OrgName:    ctx.String("org"),
		OrgInvite:  subCommand,
	}

	c := context.Background()

	fmt.Println("")
	user, err := client.Users.Create(c, &signup, &progress)
	if err != nil {
		if strings.Contains(err.Error(), "resource exists") {
			return errs.NewExitError("Username or email address in use.")
		}
		return errs.NewExitError("Signup failed, please try again.")
	}

	// Log the user in
	err = performLogin(c, client, user.Email(), password, true)
	if err != nil {
		return err
	}

	// Generate keypairs, look up the user's org
	err = generateKeypairsForOrg(c, ctx, client, nil, true)
	if err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("Your account has been created!")

	if !subCommand {
		fmt.Println("")
		fmt.Println("We have emailed you a verification code.")
		fmt.Println("Please verify your email address by entering the code below.")
		fmt.Println("")

		if err = askToVerify(ctx); err != nil {
			return err
		}

		fmt.Println("")
	}

	hints.Display(hints.GettingStarted)
	return nil
}
