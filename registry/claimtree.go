package registry

import (
	"context"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
)

// ClaimTreeClient represents the `/claimtree` registry endpoint, used for
// retrieving the public keys and their associated claims for an organization.
type ClaimTreeClient struct {
	client RoundTripper
}

// ClaimTree represents an organizations claim tree which contains public
// signing and encryption keys for every member.
type ClaimTree struct {
	Org        *envelope.Org               `json:"org"`
	PublicKeys []apitypes.PublicKeySegment `json:"public_keys"`
}

// List returns a list of all claimtrees for a given orgID. If no orgID is
// provided then it returns all claimtrees for every organization the user
// belongs too.
//
// If an ownerID is provided then only public keys and claims related to that
// user or machine will be returned.
func (c *ClaimTreeClient) List(ctx context.Context, orgID *identity.ID,
	ownerID *identity.ID) ([]ClaimTree, error) {

	query := &url.Values{}
	if orgID != nil {
		query.Set("org_id", orgID.String())
	}

	if ownerID != nil {
		query.Set("owner_id", ownerID.String())
	}

	var resp []ClaimTree
	err := c.client.RoundTrip(ctx, "GET", "/claimtree", query, nil, &resp)
	return resp, err
}
