package registry

import (
	"context"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
)

// MachinesClient represents the `/machines` registry endpoint, used for
// creating, listing, authorizing, and destroying machines and their tokens.
type MachinesClient struct {
	client RoundTripper
}

// MachineCreationSegment represents the request sent to create the registry to
// create a machine and it's first token
type MachineCreationSegment struct {
	apitypes.MachineSegment
	Tokens []MachineTokenCreationSegment `json:"tokens"`
}

// MachineTokenCreationSegment represents the request send to the registry to
// create a Machine Token
type MachineTokenCreationSegment struct {
	Token    *envelope.MachineToken `json:"token"`
	Keypairs []*ClaimedKeyPair      `json:"keypairs"`
}

// Create requests the registry to create a MachineSegment.
//
// The MachineSegment includes the Machine, it's Memberships, and authorization
// tokens.
func (m *MachinesClient) Create(ctx context.Context, machine *envelope.Machine,
	memberships []envelope.Membership, token *MachineTokenCreationSegment) (*apitypes.MachineSegment, error) {

	segment := MachineCreationSegment{
		MachineSegment: apitypes.MachineSegment{
			Machine:     machine,
			Memberships: memberships,
		},
		Tokens: []MachineTokenCreationSegment{*token},
	}

	resp := &apitypes.MachineSegment{}
	err := m.client.RoundTrip(ctx, "POST", "/machines", nil, &segment, &resp)
	return resp, err
}

// Get requests a single machine from the registry
func (m *MachinesClient) Get(ctx context.Context, machineID *identity.ID) (*apitypes.MachineSegment, error) {
	resp := &apitypes.MachineSegment{}
	err := m.client.RoundTrip(ctx, "GET", "/machines/"+(*machineID).String(), nil, nil, &resp)
	return resp, err
}

// List machines in the given org and state
func (m *MachinesClient) List(ctx context.Context, orgID *identity.ID, state *string, name *string, teamID *identity.ID) ([]apitypes.MachineSegment, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Add("org_id", (*orgID).String())
	}
	if state != nil {
		v.Add("state", *state)
	}
	if teamID != nil {
		v.Add("team_id", (*teamID).String())
	}
	if name != nil {
		v.Add("name", *name)
	}

	var results []apitypes.MachineSegment
	err := m.client.RoundTrip(ctx, "GET", "/machines", v, nil, &results)
	return results, err
}

// Destroy machine by ID
func (m *MachinesClient) Destroy(ctx context.Context, machineID *identity.ID) error {
	return m.client.RoundTrip(ctx, "DELETE", "/machines/"+machineID.String(), nil, nil, nil)
}
