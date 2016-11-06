package cmd

import (
	"fmt"

	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

type machineCreateData struct {
	TeamName string
}

func machineCreate() framework.Command {
	machine := machineCreateData{
		TeamName: "test-machine-team-" + utils.GlobalNonce,
	}

	return framework.Command{
		Spawn: fmt.Sprintf("machines create --org username-%s", utils.GlobalNonce),
		Expect: []string{
			"Generating machine token",
			"Generating token keypairs",
			"Creating keyring memberships for token",
			"Uploading keyring memberships",
			"Machine created",
			"You will only be shown the secret once, please keep it safe.",
			"Machine ID:",
			"Machine Token ID:",
			"Machine Token Secret:",
		},
		Prompt: &framework.Prompt{
			Fields: []framework.Field{
				{
					Label:    "Create a new team",
					SendLine: machine.TeamName,
				},
				{
					Label: "Enter machine name",
				},
			},
		},
	}
}

func machineList() framework.Command {
	machineTeamName := "test-machine-team-" + utils.GlobalNonce
	return framework.Command{
		Spawn: fmt.Sprintf("machines list --org username-%s", utils.GlobalNonce),
		Expect: []string{
			"ID[\\s]+NAME[\\s]+STATE[\\s]+TEAM[\\s]+CREATION DATE",
			fmt.Sprintf("[a-z0-9]+[\\s]+%s\\-[a-z0-9]+[\\s]+active[\\s]+%s[\\s]+[0-9]{4}\\-[0-9]{2}\\-[0-9]{2}T[0-9]{2}\\:[0-9]{2}\\:[0-9]{2}Z", machineTeamName, machineTeamName),
		},
	}
}
