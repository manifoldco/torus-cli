package api

import (
	"context"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/registry"
)

// UsersClient makes requests to the registry's and daemon's users endpoints
type UsersClient struct {
	*registry.UsersClient
	client *apiRoundTripper
}

func newUsersClient(upstream *registry.UsersClient, rt *apiRoundTripper) *UsersClient {
	return &UsersClient{upstream, rt}
}

// Create will have the daemon create a new user request
func (u *UsersClient) Create(ctx context.Context, signup *apitypes.Signup, output *ProgressFunc) (envelope.UserInf, error) {
	e := envelope.Unsigned{}
	err := u.client.DaemonRoundTrip(ctx, "POST", "/signup", nil, &signup, &e, nil)
	if err != nil {
		return nil, err
	}

	return envelope.ConvertUser(&e)
}

// Update performs a profile update to the user object
func (u *UsersClient) Update(ctx context.Context, delta apitypes.ProfileUpdate) (envelope.UserInf, error) {
	e := envelope.Unsigned{}
	err := u.client.DaemonRoundTrip(ctx, "PATCH", "/self", nil, &delta, &e, nil)
	if err != nil {
		return nil, err
	}

	return envelope.ConvertUser(&e)
}
