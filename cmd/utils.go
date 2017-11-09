package cmd

import (
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/primitive"
)

// ByTeamPrecedence will implement the sort Interface
type ByTeamPrecedence []envelope.Team

func (tp ByTeamPrecedence) Len() int {
	return len(tp)
}

func (tp ByTeamPrecedence) Swap(i, j int) {
	tp[i], tp[j] = tp[j], tp[i]
}

func (tp ByTeamPrecedence) Less(i, j int) bool {

	// 1) 	If both teams are system teams, must compare owner < admin < member
	//		where "<" means has higher precendence (ie. owner (0) < admin (1) where
	//		owner has a higher precedence).
	//
	// 2) 	Any system team automatically has higher precedence than a non-system team.
	//
	// 3) 	If both teams are non-system user teams, then list in alphabetical order.
	//
	// 4)	If a team is a machine team, it automatically has the lowest precedence.
	//
	// 5) 	If both teams are machine teams, then list in alphabetical order.
	//
	if tp[i].Body.TeamType == "system" && tp[j].Body.TeamType == "system" {
		return primitive.SystemTeams[tp[i].Body.Name] < primitive.SystemTeams[tp[j].Body.Name]
	} else if tp[i].Body.TeamType == "system" {
		return true
	} else if tp[j].Body.TeamType == "system" {
		return false
	} else if tp[i].Body.TeamType == "user" && tp[j].Body.TeamType == "user" {
		return tp[i].Body.Name[0] < tp[j].Body.Name[0]
	} else if tp[i].Body.TeamType == "user" {
		return true
	} else if tp[j].Body.TeamType == "user" {
		return false
	}
	return tp[i].Body.Name[0] < tp[j].Body.Name[0]
}
