package apitypes

import (
	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// MachineSegment represents a machine, its tokens, and their connected keypairs
type MachineSegment struct {
	Machine *struct {
		ID   *identity.ID       `json:"id"`
		Body *primitive.Machine `json:"body"`
	} `json:"machine"`
	Memberships []*struct {
		ID   *identity.ID          `json:"id"`
		Body *primitive.Membership `json:"body"`
	} `json:"memberships"`
	Tokens []*struct {
		Token *struct {
			ID   *identity.ID            `json:"id"`
			Body *primitive.MachineToken `json:"body"`
		} `json:"token"`
		Keypairs []PublicKeySegment `json:"keypairs"`
	} `json:"tokens"`
}

// MachinesCreateRequest represents a request by a client to create a machine
// for a specific org, team using the given name and secret.
type MachinesCreateRequest struct {
	Name   string        `json:"name"`
	OrgID  *identity.ID  `json:"org_id"`
	TeamID *identity.ID  `json:"team_id"`
	Secret *base64.Value `json:"secret"`
}
