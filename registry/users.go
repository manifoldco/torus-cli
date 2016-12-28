package registry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"

	"github.com/manifoldco/torus-cli/daemon/crypto"
)

// UsersClient represents the  registry `/users` endpoints.
type UsersClient struct {
	client RoundTripper
}

// Create attempts to register a new user
func (u *UsersClient) Create(ctx context.Context, userObj *envelope.User,
	signup apitypes.Signup) (*envelope.User, error) {

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

	user := envelope.User{}
	err := u.client.RoundTrip(ctx, "POST", "/users", v, &userObj, &user)
	if err != nil {
		return nil, err
	}

	err = validateSelf(&user)
	if err != nil {
		log.Printf("Invalid user object: %s", err)
		return nil, err
	}

	return &user, nil
}

// VerifyEmail will confirm the user's email with the registry
func (u *UsersClient) VerifyEmail(ctx context.Context, verifyCode string) error {
	verify := apitypes.VerifyEmail{
		Code: verifyCode,
	}
	return u.client.RoundTrip(ctx, "POST", "/users/verify", nil, &verify, nil)
}

// Update patches the user object with whitelisted fields
func (u *UsersClient) Update(ctx context.Context, userObj interface{}) (*envelope.User, error) {
	user := envelope.User{}
	err := u.client.RoundTrip(ctx, "PATCH", "/users/self", nil, userObj, &user)
	if err != nil {
		return nil, err
	}

	err = validateSelf(&user)
	if err != nil {
		log.Printf("Invalid user object: %s", err)
		return nil, err
	}

	return &user, nil
}

func validateSelf(s *envelope.User) error {
	if s.Version != 1 {
		return errors.New("version must be 1")
	}

	if s.Body == nil {
		return errors.New("missing body")
	}

	if s.Body.Master == nil {
		return errors.New("missing master key section")
	}

	if s.Body.Master.Alg != crypto.Triplesec {
		return &apitypes.Error{
			Type: apitypes.InternalServerError,
			Err: []string{
				fmt.Sprintf("Unknown alg: %s", s.Body.Master.Alg),
			},
		}
	}

	if len(*s.Body.Master.Value) == 0 {
		return errors.New("Zero length master key found")
	}

	return nil
}
