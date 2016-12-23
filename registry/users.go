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

// UsersClient represents the  registry `/users` endpoints.
type UsersClient struct {
	client RoundTripper
}

// Create attempts to register a new user
func (u *UsersClient) Create(ctx context.Context, userObj *SignupEnvelope,
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
	req, err := u.client.NewRequest("POST", "/users", v, userObj)
	if err != nil {
		log.Printf("Error making api request: %s", err)
		return nil, err
	}

	user := envelope.User{}
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

// VerifyEmail will confirm the user's email with the registry
func (u *UsersClient) VerifyEmail(ctx context.Context, verifyCode string) error {
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

// Update patches the user object with whitelisted fields
func (u *UsersClient) Update(ctx context.Context, userObj interface{}) (*envelope.User, error) {
	req, err := u.client.NewRequest("PATCH", "/users/self", nil, userObj)
	if err != nil {
		log.Printf("Error making api request: %s", err)
		return nil, err
	}

	user := envelope.User{}
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

type omit struct{}

// SignupEnvelope contains fields for signup
type SignupEnvelope struct {
	envelope.User
	Body *SignupBody `json:"body"`
}

// SignupBody contains fields for Signup object body during signup
type SignupBody struct {
	primitive.User

	State omit `json:"state,omitempty"`
}
