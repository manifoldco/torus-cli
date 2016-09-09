package api

import (
	"context"
	"errors"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/envelope"
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

	result := envelope.Unsigned{}
	_, err = u.client.Do(ctx, req, &result, nil, nil)
	if err != nil {
		return nil, err
	}

	user := UserResult{}
	user.ID = result.ID
	user.Version = result.Version

	userBody, ok := result.Body.(*primitive.User)
	if !ok {
		return nil, errors.New("invalid user body")
	}
	user.Body = userBody

	return &user, nil
}

// Signup will have the daemon create a new user request
func (u *UsersClient) Signup(ctx context.Context, signup *apitypes.Signup, output *ProgressFunc) (*UserResult, error) {
	req, _, err := u.client.NewRequest("POST", "/signup", nil, &signup, false)
	if err != nil {
		return nil, err
	}

	result := envelope.Unsigned{}
	_, err = u.client.Do(ctx, req, &result, nil, output)
	if err != nil {
		return nil, err
	}

	user := UserResult{}
	user.ID = result.ID
	user.Version = result.Version

	userBody, ok := result.Body.(*primitive.User)
	if !ok {
		return nil, errors.New("invalid user body")
	}
	user.Body = userBody

	return &user, nil
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
