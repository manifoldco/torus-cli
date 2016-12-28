package api

import (
	"context"
	"crypto/rand"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/registry"
)

const tokenSecretSize = 18

// MachinesClient makes requests to the Daemon on behalf of the user to
// manipulate Machine resources.
type MachinesClient struct {
	*registry.MachinesClient
	client *apiRoundTripper
}

func newMachinesClient(upstream *registry.MachinesClient, rt *apiRoundTripper) *MachinesClient {
	return &MachinesClient{upstream, rt}
}

// Create a new machine in the given org
func (m *MachinesClient) Create(ctx context.Context, orgID, teamID *identity.ID,
	name string, output ProgressFunc) (*apitypes.MachineSegment, *base64.Value, error) {

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

	result := &apitypes.MachineSegment{}
	err = m.client.DaemonRoundTrip(ctx, "POST", "/machines", nil, &mcr, &result, output)
	return result, secret, err
}

func createTokenSecret() (*base64.Value, error) {
	value := make([]byte, tokenSecretSize)
	_, err := rand.Read(value)
	if err != nil {
		return nil, err
	}

	return base64.NewValue(value), nil
}
