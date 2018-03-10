// Package registry provides access to the Torus registry REST API.
package registry

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/manifoldco/torus-cli/apitypes"
)

// TokenHolder holds an authorization token
type TokenHolder interface {
	HasToken() bool
	Token() []byte
}

// Client exposes the registry REST API.
type Client struct {
	KeyPairs        *KeyPairsClient
	Tokens          *TokensClient
	Users           *UsersClient
	Teams           *TeamsClient
	Memberships     *MembershipsClient
	Credentials     *Credentials
	Orgs            *OrgsClient
	OrgInvites      *OrgInvitesClient
	Policies        *PoliciesClient
	Projects        *ProjectsClient
	Environments    *EnvironmentsClient
	Services        *ServicesClient
	Keyring         *KeyringClient
	KeyringMember   *KeyringMemberClientV1
	Claims          *ClaimsClient
	ClaimTree       *ClaimTreeClient
	CredentialGraph *CredentialGraphClient
	Machines        *MachinesClient
	Profiles        *ProfilesClient
	Self            *SelfClient
	Version         *VersionClient
}

// NewClient returns a new Client.
func NewClient(prefix string, apiVersion string, version string,
	token TokenHolder, t http.RoundTripper) *Client {

	rt := &registryRoundTripper{
		DefaultRequestDoer: DefaultRequestDoer{
			Client: &http.Client{
				Transport: t,
				Timeout:   time.Minute,
			},
			Host: prefix,
		},

		apiVersion: apiVersion,
		version:    version,
		holder:     token,
	}

	return NewClientWithRoundTripper(rt)
}

// NewClientWithRoundTripper returns a new Client using the provided
// RoundTripper. This is used in the api package to embed registry endpoints.
func NewClientWithRoundTripper(rt RoundTripper) *Client {
	c := &Client{}

	c.KeyPairs = &KeyPairsClient{client: rt}
	c.Tokens = &TokensClient{client: rt}
	c.Users = &UsersClient{client: rt}
	c.Teams = &TeamsClient{client: rt}
	c.Memberships = &MembershipsClient{client: rt}
	c.Credentials = &Credentials{client: rt}
	c.Orgs = &OrgsClient{client: rt}
	c.OrgInvites = &OrgInvitesClient{client: rt}
	c.Policies = &PoliciesClient{client: rt}
	c.Projects = &ProjectsClient{client: rt}
	c.Environments = &EnvironmentsClient{client: rt}
	c.Services = &ServicesClient{client: rt}
	c.Claims = &ClaimsClient{client: rt}
	c.ClaimTree = &ClaimTreeClient{client: rt}
	c.Keyring = &KeyringClient{client: rt}
	c.Keyring.Members = &KeyringMembersClient{client: rt}
	c.KeyringMember = &KeyringMemberClientV1{client: rt}
	c.CredentialGraph = &CredentialGraphClient{client: rt}
	c.Machines = &MachinesClient{client: rt}
	c.Profiles = &ProfilesClient{client: rt}
	c.Self = &SelfClient{client: rt}
	c.Version = &VersionClient{client: rt}

	return c
}

type registryRoundTripper struct {
	DefaultRequestDoer

	apiVersion string
	version    string
	holder     TokenHolder
}

// Augment the default NewRequest to set additional required headers
func (rt *registryRoundTripper) NewRequest(method, path string,
	query *url.Values, body interface{}) (*http.Request, error) {

	req, err := rt.DefaultRequestDoer.NewRequest(method, path, query, body)

	if err != nil {
		return nil, err
	}

	if rt.holder.HasToken() {
		req.Header.Set("Authorization", "Bearer "+string(rt.holder.Token()))
	}

	req.Header.Set("User-Agent", "Torus-Daemon/"+rt.version)
	req.Header.Set("X-Registry-Version", rt.apiVersion)

	return req, nil
}

// Augment the default Do to set a timeout.
func (rt *registryRoundTripper) Do(ctx context.Context, r *http.Request,
	v interface{}) (*http.Response, error) {

	ctx, cancelFunc := context.WithTimeout(ctx, 60*time.Second)
	r = r.WithContext(ctx)
	defer cancelFunc()

	resp, err := rt.DefaultRequestDoer.Do(ctx, r, v)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			err = &apitypes.Error{
				Type: apitypes.RequestTimeoutError,
				Err:  []string{"Request timed out"},
			}
		}

		return nil, err
	}

	return resp, nil
}

func (rt *registryRoundTripper) RoundTrip(ctx context.Context, method, path string, query *url.Values, body, response interface{}) error {
	req, err := rt.NewRequest(method, path, query, body)
	if err != nil {
		return err
	}

	_, err = rt.Do(ctx, req, response)
	return err
}
