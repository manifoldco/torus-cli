package api

import (
	"context"
	"errors"

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
func (o *UsersClient) Self(ctx context.Context) (*UserResult, error) {
	req, _, err := o.client.NewRequest("GET", "/users/self", nil, nil, true)
	if err != nil {
		return nil, err
	}

	result := envelope.Unsigned{}
	_, err = o.client.Do(ctx, req, &result, nil, nil)
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
