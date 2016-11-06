package cmd

import (
	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

func daemonStatus() framework.Command {
	return framework.Command{
		Spawn: "daemon status",
		Expect: []string{
			`Daemon is running\. pid\: [0-9]+[\s]+version\: v` + utils.SemverRegex,
		},
	}
}

func daemonStop() framework.Command {
	return framework.Command{
		Spawn: "daemon stop",
		Expect: []string{
			"Daemon stopped gracefully.",
		},
	}
}

func daemonStart() framework.Command {
	return framework.Command{
		Spawn: "daemon start",
		Expect: []string{
			"Daemon started.",
		},
	}
}
