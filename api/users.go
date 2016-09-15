package api

import (
	"context"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
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

// Self retrieves the currently logged in user
func (u *UsersClient) Self(ctx context.Context) (*UserResult, error) {
	req, _, err := u.client.NewRequest("GET", "/users/self", nil, nil, true)
	if err != nil {
		return nil, err
	}

	user := UserResult{}
	_, err = u.client.Do(ctx, req, &user, nil, nil)
	return &user, err
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
	if err != nil {
		return err
	}

	return nil
}
