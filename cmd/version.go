package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"
)

func init() {
	version := cli.Command{
		Name:     "version",
		Usage:    "Display versions of utility components",
		Category: "SYSTEM",
		Action:   VersionLookup,
	}
	Cmds = append(Cmds, version)
}

// VersionLookup ensures the environment is ready and then executes version cmd
func VersionLookup(ctx *cli.Context) error {
	return chain(
		ensureDaemon, listVersionsCmd,
	)(ctx)
}

func listVersionsCmd(ctx *cli.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)
	c := context.Background()
	daemonVersion, registryVersion := retrieveVersions(c, client)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\n", "CLI", cfg.Version)
	fmt.Fprintf(w, "%s\t%s\n", "Daemon", daemonVersion.Version)
	fmt.Fprintf(w, "%s\t%s\n", "Registry", registryVersion.Version)
	w.Flush()

	return nil
}

func retrieveVersions(c context.Context, client *api.Client) (*apitypes.Version, *apitypes.Version) {
	var wg sync.WaitGroup
	wg.Add(1)

	var daemonVersion *apitypes.Version

	go func() {
		var dErr error
		daemonVersion, dErr = client.Version.Get(c)
		if dErr != nil {
			daemonVersion = &apitypes.Version{
				Version: "unknown (" + dErr.Error() + ")",
			}
		}
		wg.Done()
	}()

	registryVersion, rErr := client.Version.GetRegistry(c)
	if rErr != nil {
		registryVersion = &apitypes.Version{
			Version: "unknown (" + rErr.Error() + ")",
		}
	}
	wg.Wait()

	return daemonVersion, registryVersion

}
