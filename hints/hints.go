package hints

import (
	"math/rand"
	"time"

	"github.com/manifoldco/torus-cli/ui"
)

var commandHints map[string][]string

// Init generates the commandHints map
func Init() {
	commandHints = map[string][]string{
		"allow": {
			"Grant additional access to secrets for a team or role using `torus allow`",
		},
		"context": {
			"Your linked organization and project are found in .torus.json after running `torus link`",
		},
		"deny": {
			"Restrict access to secrets for a team or role using `torus deny`",
		},
		"invites approve": {
			"Approve multiple invites with `torus worklog resolve`",
		},
		"invites send": {
			"Invite another user to join your organization with `torus invites send`",
		},
		"link": {
			"Define an organization and project for your current working directory using `torus link`",
		},
		"ls": {
			"Explore the objects and secrets for your org through `torus ls`",
		},
		"path": {
			"Each secret path has 7 segments: /org/project/env/service/identity/instance/secret",
			"Secret paths can contain wildcards, such as `dev-*` for the environment",
		},
		"policies": {
			"Display policies for your organization with `torus policies list`",
			"View details of an existing policy with `torus policies view`",
		},
		"projects": {
			"Create a project for your secrets using `torus projects create`",
		},
		"run": {
			"Start your process with your decrypted secrets using `torus run`",
		},
		"system": {
			"Disable hints with `torus prefs set core.hints false`",
		},
		"teams members": {
			"Display current members of your organization with `torus members member`",
		},
		"view": {
			"View secret values which have been set using `torus view`",
			"See the exact path for each secret set using `torus view -v`",
		},
	}
}

// Display chooses a random hint from the allotted commands and displays it
func Display(possible []string) {
	// Seed random integer
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	i := r.Intn(len(possible))
	cmd := possible[i]
	if hints, ok := commandHints[cmd]; ok {
		// Display random hint from chosen command's list
		item := r.Intn(len(hints))
		if len(hints) > item {
			ui.Hint(hints[item], false)
		}
	}
}
