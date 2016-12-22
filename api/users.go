package api

import (
	"context"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
)

// upstreamUsersClient makes proxied requests to the registry's users endpoints
type upstreamUsersClient struct {
	client RoundTripper
}

// UsersClient makes requests to the registry's and daemon's users endpoints
type UsersClient struct {
	upstreamUsersClient
	client *apiRoundTripper
}

func newUsersClient(rt *apiRoundTripper) *UsersClient {
	return &UsersClient{upstreamUsersClient{rt}, rt}
}

// Signup will have the daemon create a new user request
func (u *UsersClient) Signup(ctx context.Context, signup *apitypes.Signup, output *ProgressFunc) (*envelope.User, error) {
	req, _, err := u.client.NewDaemonRequest("POST", "/signup", nil, &signup)
	if err != nil {
		return nil, err
	}

	user := envelope.User{}
	_, err = u.client.Do(ctx, req, &user)
	return &user, err
}

// VerifyEmail will confirm the user's email with the registry
func (u *upstreamUsersClient) VerifyEmail(ctx context.Context, verifyCode string) error {
	verify := apitypes.VerifyEmail{
		Code: verifyCode,
	}
	req, err := u.client.NewRequest("POST", "/users/verify", nil, &verify)
	if err != nil {
		return err
	}

	_, err = u.client.Do(ctx, req, nil)
	return err
}

// Update performs a profile update to the user object
func (u *UsersClient) Update(ctx context.Context, delta apitypes.ProfileUpdate) (*envelope.User, error) {
	req, _, err := u.client.NewDaemonRequest("PATCH", "/self", nil, &delta)
	if err != nil {
		return nil, err
	}

	user := envelope.User{}
	_, err = u.client.Do(ctx, req, &user)
	return &user, err
}
