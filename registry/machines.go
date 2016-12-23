package registry

import (
	"context"
	"log"
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

	req, err := m.client.NewRequest("POST", "/machines", nil, &segment)
	if err != nil {
		log.Printf("Error building POST Machines Request: %s", err)
		return nil, err
	}

	resp := &apitypes.MachineSegment{}
	_, err = m.client.Do(ctx, req, resp)
	if err != nil {
		log.Printf("Failed to create machine: %s", err)
		return nil, err
	}

	return resp, nil
}

// Get requests a single machine from the registry
func (m *MachinesClient) Get(ctx context.Context, machineID *identity.ID) (*apitypes.MachineSegment, error) {
	req, err := m.client.NewRequest("GET", "/machines/"+(*machineID).String(), nil, nil)
	if err != nil {
		log.Printf("Error building GET Machines Request: %s", err)
		return nil, err
	}

	resp := &apitypes.MachineSegment{}
	_, err = m.client.Do(ctx, req, resp)
	if err != nil {
		log.Printf("Failed to retrieve machine: %s", err)
		return nil, err
	}

	return resp, nil
}

// List machines in the given org and state
func (m *MachinesClient) List(ctx context.Context, orgID *identity.ID, state *string, name *string, teamID *identity.ID) ([]*apitypes.MachineSegment, error) {
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

	req, err := m.client.NewRequest("GET", "/machines", v, nil)
	if err != nil {
		return nil, err
	}

	var results []*apitypes.MachineSegment
	_, err = m.client.Do(ctx, req, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// Destroy machine by ID
func (m *MachinesClient) Destroy(ctx context.Context, machineID *identity.ID) error {
	req, err := m.client.NewRequest("DELETE", "/machines/"+machineID.String(), nil, nil)
	if err != nil {
		return err
	}

	_, err = m.client.Do(ctx, req, nil)

	return err
}
