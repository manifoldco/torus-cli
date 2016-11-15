package api

import (
	"context"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// UsersClient makes proxied requests to the registry's users endpoints
type UsersClient struct {
	client *Client
}

// UserResult is the payload returned for a user object
type UserResult struct {
	ID      *identity.ID    `json:"id"`
	Version uint8           `json:"version"`
	Body    *primitive.User `json:"body"`
}

// Signup will have the daemon create a new user request
func (u *UsersClient) Signup(ctx context.Context, signup *apitypes.Signup, output *ProgressFunc) (*UserResult, error) {
	req, _, err := u.client.NewRequest("POST", "/signup", nil, &signup, false)
	if err != nil {
		return nil, err
	}

	user := UserResult{}
	_, err = u.client.Do(ctx, req, &user, nil, output)
	return &user, err
}

// VerifyEmail will confirm the user's email with the registry
func (u *UsersClient) VerifyEmail(ctx context.Context, verifyCode string) error {
	verify := apitypes.VerifyEmail{
		Code: verifyCode,
	}
	req, _, err := u.client.NewRequest("POST", "/users/verify", nil, &verify, true)
	if err != nil {
		return err
	}

	_, err = u.client.Do(ctx, req, nil, nil, nil)
	return err
}

// Update performs a profile update to the user object
func (u *UsersClient) Update(ctx context.Context, delta apitypes.ProfileUpdate) (*UserResult, error) {
	req, _, err := u.client.NewRequest("PATCH", "/self", nil, &delta, false)
	if err != nil {
		return nil, err
	}

	user := UserResult{}
	_, err = u.client.Do(ctx, req, &user, nil, nil)
	return &user, err
}
