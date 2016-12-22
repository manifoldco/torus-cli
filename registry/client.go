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
	Token() string
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
	OrgInvite       *OrgInvitesClient
	Projects        *ProjectsClient
	Keyring         *KeyringClient
	KeyringMember   *KeyringMemberClientV1
	Claims          *ClaimsClient
	ClaimTree       *ClaimTreeClient
	CredentialGraph *CredentialGraphClient
	Machines        *MachinesClient
	Self            *SelfClient
}

// NewClient returns a new Client.
func NewClient(prefix string, apiVersion string, version string,
	token TokenHolder, t *http.Transport) *Client {

	rt := &registryRoundTripper{
		DefaultRoundTripper: DefaultRoundTripper{
			Client: &http.Client{Transport: t},
			Host:   prefix,
		},

		apiVersion: apiVersion,
		version:    version,
		holder:     token,
	}

	c := &Client{}

	c.KeyPairs = &KeyPairsClient{client: rt}
	c.Tokens = &TokensClient{client: rt}
	c.Users = &UsersClient{client: rt}
	c.Teams = &TeamsClient{client: rt}
	c.Memberships = &MembershipsClient{client: rt}
	c.Credentials = &Credentials{client: rt}
	c.Orgs = &OrgsClient{client: rt}
	c.OrgInvite = &OrgInvitesClient{client: rt}
	c.Projects = &ProjectsClient{client: rt}
	c.Claims = &ClaimsClient{client: rt}
	c.ClaimTree = &ClaimTreeClient{client: rt}
	c.Keyring = &KeyringClient{client: rt}
	c.Keyring.Members = &KeyringMembersClient{client: rt}
	c.KeyringMember = &KeyringMemberClientV1{client: rt}
	c.CredentialGraph = &CredentialGraphClient{client: rt}
	c.Machines = &MachinesClient{client: rt}
	c.Self = &SelfClient{client: rt}

	return c
}

type registryRoundTripper struct {
	DefaultRoundTripper

	apiVersion string
	version    string
	holder     TokenHolder
}

// Augment the default NewRequest to set additional required headers
func (rt *registryRoundTripper) NewRequest(method, path string,
	query *url.Values, body interface{}) (*http.Request, error) {

	req, err := rt.DefaultRoundTripper.NewRequest(method, path, query, body)

	if err != nil {
		return nil, err
	}

	if tok := rt.holder.Token(); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	req.Header.Set("User-Agent", "Torus-Daemon/"+rt.version)
	req.Header.Set("X-Registry-Version", rt.apiVersion)

	return req, nil
}

// Augment the default Do to set a timeout.
func (rt *registryRoundTripper) Do(ctx context.Context, r *http.Request,
	v interface{}) (*http.Response, error) {

	ctx, cancelFunc := context.WithTimeout(ctx, 6*time.Second)
	r = r.WithContext(ctx)
	defer cancelFunc()

	resp, err := rt.DefaultRoundTripper.Do(ctx, r, v)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			err = &apitypes.Error{
				StatusCode: http.StatusRequestTimeout,
				Type:       "request_timeout",
				Err:        []string{"Request timed out"},
			}
		}

		return nil, err
	}

	return resp, nil
}
