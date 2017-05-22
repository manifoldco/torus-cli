package cmd

import (
	"fmt"
	"os"

	"github.com/manifoldco/go-base64"
	"github.com/urfave/cli"

	"bufio"
	"path/filepath"

	"github.com/manifoldco/torus-cli/gatekeeper/bootstrap"
	"github.com/manifoldco/torus-cli/identity"
)

const (
	// GlobalRoot is the global root of the Torus config
	GlobalRoot = "/etc/torus"

	// EnvironmentFile is the environment file that stores machine information
	EnvironmentFile = "token.environment"
)

func init() {
	bootstrap := cli.Command{
		Name:     "bootstrap",
		Usage:    "Bootstrap a new machine using Torus Gatekeeper",
		Category: "SYSTEM",
		Flags: []cli.Flag{
			authProviderFlag("Auth provider for bootstrapping", true),
			machineFlag("Machine name to bootstrap", false),
			urlFlag("Gatekeeper URL for bootstrapping", true),
			orgFlag("Org the machine will belong to", false),
			roleFlag("Role the machine will belong to", false),
		},
		Action: chain(checkRequiredFlags, bootstrapCmd),
	}

	Cmds = append(Cmds, bootstrap)
}

// bootstrapCmd is the cli.Command for Bootstrapping machine configuration from the Gatekeeper
func bootstrapCmd(ctx *cli.Context) error {
	cloud := bootstrap.Type(ctx.String("auth"))

	provider, err := bootstrap.New(cloud)
	if err != nil {
		return fmt.Errorf("bootstrap init failed: %s", err)
	}

	resp, err := provider.Bootstrap(
		ctx.String("url"),
		ctx.String("name"),
		ctx.String("org"),
		ctx.String("role"),
	)
	if err != nil {
		return fmt.Errorf("bootstrap provision failed: %s", err)
	}

	envFile := filepath.Join(GlobalRoot, EnvironmentFile)
	if err = writeEnvironmentFile(resp.Token, resp.Secret); err != nil {
		return fmt.Errorf("failed to write environment file[%s]: %s", envFile, err)
	}

	fmt.Printf("Machine bootstrapped. Environment configuration saved in %s", envFile)
	return nil
}

func writeEnvironmentFile(token *identity.ID, secret *base64.Value) error {
	err := os.Mkdir(GlobalRoot, os.FileMode(0740))
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(GlobalRoot, EnvironmentFile))
	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)
	w.WriteString(fmt.Sprintf("TORUS_TOKEN_ID=%s", token))
	w.WriteString(fmt.Sprintf("TORUS_TOKEN_SECURE=%s", secret))

	return nil
}
