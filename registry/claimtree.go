package registry

import (
	"context"
	"errors"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// ErrKeyNotFound represents a situation where a key could not be found
var ErrKeyNotFound = &apitypes.Error{
	Type: apitypes.NotFoundError,
	Err:  []string{"Could not locate public key segment for specified key id"},
}

// ErrMissingKeyForOwner represents a situation where a key of a specific type could not be found for the owner inside an org
var ErrMissingKeyForOwner = &apitypes.Error{
	Type: apitypes.NotFoundError,
	Err:  []string{"Could not locate public key segment in claimtree for target"},
}

// ErrClaimTreeNotFound represents a situation where a claimtree could not be
// found
var ErrClaimTreeNotFound = &apitypes.Error{
	Type: apitypes.NotFoundError,
	Err:  []string{"Could not find claim tree for org"},
}

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

// Find returns the PublicKeySegment for the given PublicKeyID. Accepts a
// boolean for indicating whether or not to enforce that the key must be
// active.
//
// If a key segment could not be found an error is returned.
func (ct *ClaimTree) Find(id *identity.ID, mustActive bool) (*apitypes.PublicKeySegment, error) {
	for _, pks := range ct.PublicKeys {
		if *pks.PublicKey.ID == *id {
			if mustActive && pks.Revoked() {
				continue
			}

			return &pks, nil
		}
	}

	return nil, ErrKeyNotFound
}

// FindActive returns the PublicKeySegment for a non-revoked Public Key for
// the given owner id.
//
// If an active key cannot be found an error is returned
func (ct *ClaimTree) FindActive(ownerID *identity.ID, t primitive.KeyType) (*apitypes.PublicKeySegment, error) {
	for _, pks := range ct.PublicKeys {
		if *pks.PublicKey.Body.OwnerID != *ownerID {
			continue
		}

		if pks.PublicKey.Body.KeyType != t {
			continue
		}

		if pks.Revoked() {
			continue
		}

		return &pks, nil
	}

	return nil, ErrMissingKeyForOwner
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

// Get returns a claimtree for a specific organization by the given orgID.
//
// If an ownerID is provided then only public keys and claims related to that
// user or machine will be returned.
func (c *ClaimTreeClient) Get(ctx context.Context, orgID *identity.ID,
	ownerID *identity.ID) (*ClaimTree, error) {

	if orgID == nil {
		return nil, errors.New("An org id must be provided")
	}

	cts, err := c.List(ctx, orgID, ownerID)
	if err != nil {
		return nil, err
	}

	if len(cts) != 1 {
		return nil, ErrClaimTreeNotFound
	}

	return &cts[0], nil
}
