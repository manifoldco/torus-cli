package apitypes

import (
	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
)

// MachineSegment represents a machine, its tokens, and their connected keypairs
type MachineSegment struct {
	Machine     *envelope.Machine     `json:"machine"`
	Memberships []envelope.Membership `json:"memberships"`
	Tokens      []struct {
		Token    *envelope.MachineToken `json:"token"`
		Keypairs []PublicKeySegment     `json:"keypairs"`
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
