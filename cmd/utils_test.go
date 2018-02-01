package cmd

import (
	"sort"
	"testing"

	gm "github.com/onsi/gomega"

	"github.com/manifoldco/torus-cli/primitive"
)

var expectedOrder = []string{
	primitive.OwnerTeamName,
	primitive.AdminTeamName,
	primitive.MemberTeamName,
	"cat",
	"dog",
	"elephant",
	"apple",
}

// TestTeamPrecedenceSort tests that the ByTeamPrecedence correctly
// implements the Sort interface.
//
// The Version number in each enevelope.Team structure is used as a hard-coded
// expected index in the final, sorted list of teams.
func TestTeamPrecedenceSort(t *testing.T) {
	gm.RegisterTestingT(t)
	testTeamList := ByTeamPrecedence{
		// Machine -> machine team
		{
			Body: &primitive.Team{
				Name:     "apple",
				TeamType: primitive.MachineTeamType,
			},
		},
		// User -> cat team
		{
			Body: &primitive.Team{
				Name:     "cat",
				TeamType: primitive.UserTeamType,
			},
		},
		// User -> dog team
		{
			Body: &primitive.Team{
				Name:     "dog",
				TeamType: primitive.UserTeamType,
			},
		},
		// User elephant team
		{
			Body: &primitive.Team{
				Name:     "elephant",
				TeamType: primitive.UserTeamType,
			},
		},
		// System -> Member team
		{
			Body: &primitive.Team{
				Name:     primitive.MemberTeamName,
				TeamType: primitive.SystemTeamType,
			},
		},
		{
			Body: &primitive.Team{
				Name:     primitive.AdminTeamName,
				TeamType: primitive.SystemTeamType,
			},
		},
		// System -> Owner team
		{
			Body: &primitive.Team{
				Name:     primitive.OwnerTeamName,
				TeamType: primitive.SystemTeamType,
			},
		},
	}

	sort.Sort(ByTeamPrecedence(testTeamList))

	var foundOrder []string
	for _, team := range testTeamList {
		foundOrder = append(foundOrder, team.Body.Name)
	}

	gm.Expect(foundOrder).Should(gm.Equal(expectedOrder))
}

func TestPathExpParsing(t *testing.T) {
	gm.RegisterTestingT(t)

	t.Run("handles **", func(t *testing.T) {
		gm.RegisterTestingT(t)

		pe, err := parsePathExp("/org/silly/**")
		gm.Expect(err).To(gm.BeNil())
		gm.Expect(pe.String()).To(gm.Equal("/org/silly/*/*/*/*"))
	})

	t.Run("handles no path with instance or identity", func(t *testing.T) {
		gm.RegisterTestingT(t)

		pe, err := parsePathExp("/org/silly/env/service")
		gm.Expect(err).To(gm.BeNil())
		gm.Expect(pe.String()).To(gm.Equal("/org/silly/env/service/*/*"))
	})

	t.Run("handles full path", func(t *testing.T) {
		gm.RegisterTestingT(t)

		path := "/org/silly/env/service/identity/instance"
		pe, err := parsePathExp(path)
		gm.Expect(err).To(gm.BeNil())
		gm.Expect(pe.String()).To(gm.Equal(path))
	})
}
