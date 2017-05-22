package hints

import (
	"math/rand"
	"time"

	"github.com/manifoldco/torus-cli/ui"
)

type Hint int

const (
	Allow Hint = iota
	Context
	Deny
	InvitesApprove
	InvitesSend
	Link
	Ls
	Path
	Policies
	Projects
	Run
	Set
	TeamMembers
	View
)

var commandHints map[Hint][]string

func init() {
	commandHints = map[Hint][]string{
		Allow: {
			"Grant additional access to secrets for a team or role using `torus allow`",
		},
		Context: {
			"Your linked organization and project are found in .torus.json after running `torus link`",
		},
		Deny: {
			"Restrict access to secrets for a team or role using `torus deny`",
		},
		InvitesApprove: {
			"Approve multiple invites with `torus worklog resolve`",
		},
		InvitesSend: {
			"Invite another user to join your organization with `torus invites send`",
		},
		Link: {
			"Define an organization and project for your current working directory using `torus link`",
		},
		Ls: {
			"Explore the objects and secrets for your org through `torus ls`",
		},
		Path: {
			"Each secret path has 7 segments: /org/project/env/service/identity/instance/secret",
			"Secret paths can contain wildcards, such as `dev-*` for the environment",
		},
		Policies: {
			"Display policies for your organization with `torus policies list`",
			"View details of an existing policy with `torus policies view`",
		},
		Projects: {
			"Create a project for your secrets using `torus projects create`",
		},
		Run: {
			"Start your process with your decrypted secrets using `torus run`",
		},
		TeamMembers: {
			"Display current members of your organization with `torus members member`",
		},
		View: {
			"View secret values which have been set using `torus view`",
			"See the exact path for each secret set using `torus view -v`",
		},
	}
}

// Display chooses a random hint from the allotted commands and displays it
func Display(possible ...Hint) {
	hint := randHint(possible)
	if hint != "" {
		ui.Hint(hint, false)
	}
}

// randHint returns random hint from chosen command's list
func randHint(possible []Hint) string {
	l := len(possible)
	if l == 0 {
		return ""
	}

	// Seed random integer
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	i := r.Intn(l)
	cmd := possible[i]
	if hints, ok := commandHints[cmd]; ok {
		item := r.Intn(len(hints))
		if len(hints) > item {
			return hints[item]
		}
	}
	return ""
}
