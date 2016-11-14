package registry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/primitive"

	"github.com/manifoldco/torus-cli/daemon/crypto"
)

// Users represents the  registry `/users` endpoints.
type Users struct {
	client *Client
}

// Create attempts to register a new user
func (u *Users) Create(ctx context.Context, userObj Signup, signup apitypes.Signup) (*envelope.Unsigned, error) {
	v := &url.Values{}
	if signup.InviteCode != "" {
		v.Set("code", signup.InviteCode)
	}
	if signup.OrgInvite && signup.OrgName != "" {
		v.Set("org", signup.OrgName)
	}
	if signup.Email != "" {
		v.Set("email", signup.Email)
	}
	req, err := u.client.NewRequest("POST", "/users", v, userObj)
	if err != nil {
		log.Printf("Error making api request: %s", err)
		return nil, err
	}

	user := envelope.Unsigned{}
	_, err = u.client.Do(ctx, req, &user)
	if err != nil {
		log.Printf("Error making api request: %s", err)
		return nil, err
	}

	err = validateSelf(&user)
	if err != nil {
		log.Printf("Invalid user object: %s", err)
		return nil, err
	}

	return &user, nil
}

// Update patches the user object with whitelisted fields
func (u *Users) Update(ctx context.Context, userObj interface{}) (*envelope.Unsigned, error) {
	req, err := u.client.NewRequest("PATCH", "/users/self", nil, userObj)
	if err != nil {
		log.Printf("Error making api request: %s", err)
		return nil, err
	}

	user := envelope.Unsigned{}
	_, err = u.client.Do(ctx, req, &user)
	if err != nil {
		log.Printf("Error making api request: %s", err)
		return nil, err
	}

	err = validateSelf(&user)
	if err != nil {
		log.Printf("Invalid user object: %s", err)
		return nil, err
	}

	return &user, nil
}

func validateSelf(s *envelope.Unsigned) error {
	if s.Version != 1 {
		return errors.New("version must be 1")
	}

	body := s.Body.(*primitive.User)

	if body == nil {
		return errors.New("missing body")
	}

	if body.Master == nil {
		return errors.New("missing master key section")
	}

	if body.Master.Alg != crypto.Triplesec {
		return &apitypes.Error{
			Type: apitypes.InternalServerError,
			Err: []string{
				fmt.Sprintf("Unknown alg: %s", body.Master.Alg),
			},
		}
	}

	if len(*body.Master.Value) == 0 {
		return errors.New("Zero length master key found")
	}

	return nil
}

// Signup contains fields for signup
type Signup struct {
	ID      string      `json:"id"`
	Version int         `json:"version"`
	Body    *SignupBody `json:"body"`
}

// SignupBody contains fields for Signup object body during signup
type SignupBody struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	// State is not a field because the server determines it, client cannot
	Password *primitive.UserPassword `json:"password"`
	Master   *primitive.MasterKey    `json:"master"`
}

// Version returns the object version
func (SignupBody) Version() int {
	return 1
}

// Type returns the User byte
func (SignupBody) Type() byte {
	return 0x01
}

// Mutable indicates this object is Mutable type
func (SignupBody) Mutable() {}
