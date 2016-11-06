package cmd

import (
	"fmt"

	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

type policyCreateData struct {
	Name string
}

func policyView() framework.Command {
	policy := policyCreateData{
		Name: "system-default-member",
	}

	return framework.Command{
		Spawn: fmt.Sprintf("policies view %s", policy.Name),
		Expect: []string{
			`Name\:[\s]+system-default-member`,
			`Description\:[\s]+Provides members with their default permissions`,
			`allow \-r\-\-l \/\$\{org\}\/\*\/\[dev\-\$\{username\}\|dev\-\@\]`,
			`allow crudl \/\$\{org\}\/\*\/\[dev\-\$\{username\}\|dev\-\@\]\/\*`,
			`allow crudl \/\$\{org\}\/\*\/\[dev\-\$\{username\}\|dev\-\@\]\/\*\/\*`,
			`allow crudl \/\$\{org\}\/\*\/\[dev\-\$\{username\}\|dev\-\@\]\/\*\/\*\/\*`,
			`allow crudl \/\$\{org\}\/\*\/\[dev\-\$\{username\}\|dev\-\@\]\/\*\/\*\/\*\/\*`,
		},
	}
}

func policyList() framework.Command {
	return framework.Command{
		Spawn: "policies list",
		Expect: []string{
			`POLICY NAME[\s]+TYPE[\s]+ATTACHED TO`,
			`system-default-admin[\s]+system[\s]+admin`,
			`system-default-member[\s]+system[\s]+member`,
			`system-default-owner[\s]+system[\s]+owner`,
		},
	}
}

func policyAllow() framework.Command {
	username := "username-" + utils.GlobalNonce
	project := "project-" + utils.GlobalNonce
	environment := "dev-" + username
	service := "service-" + utils.GlobalNonce
	path := fmt.Sprintf("/%s/%s/%s/%s/*/*/allow", username, project, environment, service)
	pathEscaped := fmt.Sprintf(`\/%s\/%s\/%s\/%s\/\*\/\*\/allow`, username, project, environment, service)
	team := "team-" + utils.GlobalNonce
	return framework.Command{
		Spawn: fmt.Sprintf("allow crudl %s %s", path, team),
		Expect: []string{
			"Policy generated and attached to the " + team + " team.",
			`Effect\:[\s]+allow`,
			`Action\(s\)\:[\s]+create\, read\, update\, delete\, list`,
			`Resource\:[\s]+` + pathEscaped,
		},
	}
}

func policyDeny() framework.Command {
	username := "username-" + utils.GlobalNonce
	project := "project-" + utils.GlobalNonce
	environment := "dev-" + username
	service := "service-" + utils.GlobalNonce
	path := fmt.Sprintf("/%s/%s/%s/%s/*/*/deny", username, project, environment, service)
	pathEscaped := fmt.Sprintf(`\/%s\/%s\/%s\/%s\/\*\/\*\/deny`, username, project, environment, service)
	team := "team-" + utils.GlobalNonce
	return framework.Command{
		Spawn: fmt.Sprintf("deny crudl %s %s", path, team),
		Expect: []string{
			"Policy generated and attached to the " + team + " team.",
			`Effect\:[\s]+deny`,
			`Action\(s\)\:[\s]+create\, read\, update\, delete\, list`,
			`Resource\:[\s]+` + pathEscaped,
		},
	}
}

func policyListGenerated() framework.Command {
	return framework.Command{
		Spawn: "policies list",
		Expect: []string{
			`POLICY NAME[\s]+TYPE[\s]+ATTACHED TO`,
			`generated\-allow\-[0-9]+[\s]+user[\s]+team\-` + utils.GlobalNonce,
			`generated\-deny\-[0-9]+[\s]+user[\s]+team\-` + utils.GlobalNonce,
			`system-default-admin[\s]+system[\s]+admin`,
			`system-default-member[\s]+system[\s]+member`,
			`system-default-owner[\s]+system[\s]+owner`,
		},
	}
}
