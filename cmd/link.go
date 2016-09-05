package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/arigatomachine/cli/api"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/dirprefs"
	"github.com/arigatomachine/cli/prefs"
)

func init() {
	link := cli.Command{
		Name:     "link",
		Usage:    "Link your current directory to Arigato",
		Category: "CONTEXT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Overwrite existing organization and project links.",
			},
		},
		Action: Chain(EnsureDaemon, EnsureSession, linkCmd),
	}

	Cmds = append(Cmds, link)
}

func linkCmd(ctx *cli.Context) error {
	preferences, err := prefs.NewPreferences(true)
	if err != nil {
		return err
	}

	dPrefs, err := dirprefs.Load(false)
	if dPrefs != nil && dPrefs.Path != "" && !ctx.Bool("force") {
		msg := fmt.Sprintf(
			"This directory is already linked. Use '%s status' to view.",
			ctx.App.Name,
		)
		return cli.NewExitError(msg, -1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()

	org, project, createdNew, err := SelectCreateOrgAndProject(client, c, ctx, "", "")
	if err != nil {
		return err
	}
	if createdNew == true {
		fmt.Println("")
	}

	// write out the link
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	oName := org.Body.Name
	pName := project.Body.Name

	dPrefs.Organization = oName
	dPrefs.Project = pName
	dPrefs.Path = filepath.Join(cwd, ".arigato.json")

	err = dPrefs.Save()
	if err != nil {
		return err
	}

	// Display the output
	fmt.Println("This directory its subdirectories have been linked to:")
	w := tabwriter.NewWriter(os.Stdout, 2, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Org:\t%s\n", oName)
	fmt.Fprintf(w, "Project:\t%s\n", pName)
	w.Flush()
	fmt.Printf("\nUse '%s status' to view your full working context.\n", ctx.App.Name)

	if !preferences.Core.Context {
		fmt.Printf("Warning: context is disabled. Use '%s prefs' to enable it.\n", ctx.App.Name)
	}

	return nil
}
