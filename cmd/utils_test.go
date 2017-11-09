package cmd

import (
	"sort"
	"testing"

	gm "github.com/onsi/gomega"

	"github.com/manifoldco/torus-cli/primitive"
)

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
			Version: 6,
			Body: &primitive.Team{
				Name:     "apple",
				TeamType: primitive.MachineTeamType,
			},
		},
		// User -> cat team
		{
			Version: 3,
			Body: &primitive.Team{
				Name:     "cat",
				TeamType: primitive.UserTeamType,
			},
		},
		// User -> dog team
		{
			Version: 4,
			Body: &primitive.Team{
				Name:     "dog",
				TeamType: primitive.UserTeamType,
			},
		},
		// User elephant team
		{
			Version: 5,
			Body: &primitive.Team{
				Name:     "elephant",
				TeamType: primitive.UserTeamType,
			},
		},
		// System -> Member team
		{
			Version: 2,
			Body: &primitive.Team{
				Name:     primitive.MemberTeamName,
				TeamType: primitive.SystemTeamType,
			},
		},
		// System -> Admin team
		{
			Version: 1,
			Body: &primitive.Team{
				Name:     primitive.AdminTeamName,
				TeamType: primitive.SystemTeamType,
			},
		},
		// System -> Owner team
		{
			Version: 0,
			Body: &primitive.Team{
				Name:     primitive.OwnerTeamName,
				TeamType: primitive.SystemTeamType,
			},
		},
	}

	sort.Sort(ByTeamPrecedence(testTeamList))

	idList := []int{}
	for _, team := range testTeamList {
		idList = append(idList, int(team.Version))
	}

	listSortedCorrectly := sort.IntsAreSorted(idList)
	gm.Expect(listSortedCorrectly).To(gm.BeTrue())
}
