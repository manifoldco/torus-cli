package cmd

import (
	"github.com/manifoldco/torus-cli/tools/expect/framework"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

func keypairsList() framework.Command {
	return framework.Command{
		Spawn: "keypairs list",
		Expect: []string{
			`ID[\s]+ORG[\s]+KEY\sTYPE[\s]+VALID[\s]+CREATION DATE`,
			keypairRegexForType("signing"),
			keypairRegexForType("encryption"),
		},
	}
}

func keypairRegexForType(kptype string) string {
	return `[a-z0-9]+[\s]+username\-` + utils.GlobalNonce + `\s+` + kptype + `\s+YES\s+[0-9]{4}\-[0-9]{2}\-[0-9]{2}T[0-9]{2}\:[0-9]{2}\:[0-9]{2}Z`
}
