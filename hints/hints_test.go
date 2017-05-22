package hints

import (
	"testing"
)

func TestRandHint(t *testing.T) {
	type tc struct {
		desc string
		cmds []Hint
		hint string
	}

	testCases := []tc{
		{desc: "Empty command list", cmds: []Hint{}, hint: ""},
		{desc: "Command without hint", cmds: []Hint{Set}, hint: ""},
		{
			desc: "Command with one hint",
			cmds: []Hint{Allow},
			hint: "Grant additional access to secrets for a team or role using `torus allow`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := randHint(tc.cmds)
			if got != tc.hint {
				t.Errorf("randHint(%v) expected to be %q, got %q", tc.cmds, tc.hint, got)
			}
		})
	}

	t.Run("Command with multiple hints", func(t *testing.T) {
		got := randHint([]Hint{View})
		exp1 := "View secret values which have been set using `torus view`"
		exp2 := "See the exact path for each secret set using `torus view -v`"
		if got != exp1 && got != exp2 {
			t.Errorf("randHint(View) expected to be either %q or %q, got %s", exp1, exp2, got)
		}
	})

	t.Run("Multiple commands", func(t *testing.T) {
		got := randHint([]Hint{Allow, Deny})
		exp1 := "Grant additional access to secrets for a team or role using `torus allow`"
		exp2 := "Restrict access to secrets for a team or role using `torus deny`"
		if got != exp1 && got != exp2 {
			t.Errorf("randHint(Allow, Deny) expected to be either %q or %q, got %s", exp1, exp2, got)
		}
	})
}
