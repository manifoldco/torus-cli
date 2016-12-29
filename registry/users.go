package registry

import (
	"context"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
)

// UsersClient represents the  registry `/users` endpoints.
type UsersClient struct {
	client RoundTripper
}

// Create attempts to register a new user
func (u *UsersClient) Create(ctx context.Context, userObj *envelope.User,
	signup apitypes.Signup) (envelope.UserInf, error) {

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

	e := envelope.Unsigned{}
	err := u.client.RoundTrip(ctx, "POST", "/users", v, &userObj, &e)
	if err != nil {
		return nil, err
	}

	return envelope.ConvertUser(&e)
}

// VerifyEmail will confirm the user's email with the registry
func (u *UsersClient) VerifyEmail(ctx context.Context, verifyCode string) error {
	verify := apitypes.VerifyEmail{
		Code: verifyCode,
	}
	return u.client.RoundTrip(ctx, "POST", "/users/verify", nil, &verify, nil)
}

// Update patches the user object with whitelisted fields
func (u *UsersClient) Update(ctx context.Context, userObj interface{}) (envelope.UserInf, error) {
	e := envelope.Unsigned{}
	err := u.client.RoundTrip(ctx, "PATCH", "/users/self", nil, userObj, &e)
	if err != nil {
		return nil, err
	}

	return envelope.ConvertUser(&e)
}
