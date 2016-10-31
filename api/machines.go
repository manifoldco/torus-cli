package api

import (
	"context"
	"crypto/rand"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/identity"
)

const tokenSecretSize = 18

// MachinesClient makes requests to the Daemon on behalf of the user to
// manipualte Machine resources.
type MachinesClient struct {
	client *Client
}

// Create a new machine in the given org
func (m *MachinesClient) Create(ctx context.Context, orgID, teamID *identity.ID,
	name string, output *ProgressFunc) (*apitypes.MachineSegment, *base64.Value, error) {

	secret, err := createTokenSecret()
	if err != nil {
		return nil, nil, err
	}

	mcr := apitypes.MachinesCreateRequest{
		Name:   name,
		OrgID:  orgID,
		TeamID: teamID,
		Secret: secret,
	}

	req, reqID, err := m.client.NewRequest("POST", "/machines", nil, &mcr, false)
	if err != nil {
		return nil, nil, err
	}

	result := &apitypes.MachineSegment{}
	_, err = m.client.Do(ctx, req, result, &reqID, output)
	if err != nil {
		return nil, nil, err
	}

	return result, secret, nil
}

func createTokenSecret() (*base64.Value, error) {
	value := make([]byte, tokenSecretSize)
	_, err := rand.Read(value)
	if err != nil {
		return nil, err
	}

	return base64.NewValue(value), nil
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

	req, reqID, err := m.client.NewRequest("GET", "/machines", v, nil, true)
	if err != nil {
		return nil, err
	}

	var results []*apitypes.MachineSegment
	_, err = m.client.Do(ctx, req, &results, &reqID, nil)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// Get machine by ID
func (m *MachinesClient) Get(ctx context.Context, machineID *identity.ID) (*apitypes.MachineSegment, error) {
	req, reqID, err := m.client.NewRequest("GET", "/machines/"+machineID.String(), nil, nil, true)
	if err != nil {
		return nil, err
	}

	result := &apitypes.MachineSegment{}
	_, err = m.client.Do(ctx, req, result, &reqID, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}
