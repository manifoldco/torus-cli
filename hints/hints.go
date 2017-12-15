package hints

import (
	"math/rand"
	"time"

	"github.com/manifoldco/torus-cli/ui"
)

// Cmd represents a command that has hints available
type Cmd int

const (
	// Allow command adds hint to `torus allow`
	Allow Cmd = iota

	// Context adds hint to application context
	Context

	// Deny command adds hint to `torus deny`
	Deny

	// Export command adds hints for `torus export`
	Export

	// GettingStarted displays helpful hints when a user signs up
	GettingStarted

	// Import command adds hint to `torus import`
	Import

	// InvitesApprove command adds hint to `torus invites approve`
	InvitesApprove

	// InvitesSend command adds hint to `torus invites send`
	InvitesSend

	// Link command adds hint to `torus link`
	Link

	// Ls command adds hint to `torus ls`
	Ls

	// Path adds hint about the path expression
	Path

	// Policies command adds hint to `torus policies`
	Policies

	// Projects command adds hint to `torus projects`
	Projects

	// Run command adds hint to `torus run`
	Run

	// Set command adds hint to `torus set`
	Set

	// TeamMembers command adds hint to `torus members`
	TeamMembers

	// Unset command adds hint to `torus unset
	Unset

	// View command adds hint to `torus view`
	View

	// PersonalOrg adds hint to `torus orgs list`
	PersonalOrg

	// OrgMembers adds hint to `torus invites list`
	OrgMembers
)

var commandHints map[Cmd][]string

func init() {
	commandHints = map[Cmd][]string{
		Allow: {
			"Grant additional access to secrets for a team or role using `torus allow`",
		},
		GettingStarted: {
			"Link your code repository to an org and project by generating a .torus.json file using the `torus link` command! Share this linking with your teammates by committing to the repository.",
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
			"Display current members of your organization with `torus teams members member`",
		},
		View: {
			"View secret values which have been set using `torus view`",
			"See the exact path for each secret set using `torus view -v`",
		},
		Export: {
			"Export secrets for a specific environment and service using `torus export`",
		},
		PersonalOrg: {
			"A personal org is created for you on sign-up.",
		},
		OrgMembers: {
			"Display current members of your organization with `torus orgs members`",
		},
	}
}

// Display chooses a random hint from the allotted commands and displays it
func Display(possible ...Cmd) {
	hint := randHint(possible)
	if hint != "" {
		ui.Hint(hint, false, nil)
	}
}

// randHint returns random hint from chosen command's list
func randHint(possible []Cmd) string {
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
