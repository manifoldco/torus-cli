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

// ErrPublicKeyNotFound represents an error where a given public key inside a
// Keypairs struct could not be found.
var ErrPublicKeyNotFound = &apitypes.Error{
	Type: apitypes.NotFoundError,
	Err:  []string{"Could not find public key"},
}

// ErrMissingValidKeypair represents an error where a valid signing or
// encryption keypair could not be found for an organization
var ErrMissingValidKeypair = &apitypes.Error{
	Type: apitypes.NotFoundError,
	Err:  []string{"Missing encryption or signing keypairs"},
}

// ErrMissingKeysForOrg returns an error where the given org id is not present
// in the keypairs map
var ErrMissingKeysForOrg = &apitypes.Error{
	Type: apitypes.NotFoundError,
	Err:  []string{"Could not find keypairs for org"},
}

// Keypairs contains a slice of a users claimed keypairs for many different
// organizations
type Keypairs struct {
	keypairs map[identity.ID][]ClaimedKeyPair
	idIndex  map[identity.ID]ClaimedKeyPair
}

// NewKeypairs returns an empty keypairs struct
func NewKeypairs() *Keypairs {
	return &Keypairs{
		keypairs: map[identity.ID][]ClaimedKeyPair{},
		idIndex:  map[identity.ID]ClaimedKeyPair{},
	}
}

// Add adds the given keypairs to the list of keypairs
func (kp *Keypairs) Add(keypairs ...ClaimedKeyPair) error {
	for _, k := range keypairs {
		kp.idIndex[*k.PublicKey.ID] = k

		orgID := *k.PublicKey.Body.OrgID
		_, ok := kp.keypairs[orgID]
		if !ok {
			kp.keypairs[orgID] = []ClaimedKeyPair{k}
			continue
		}

		kp.keypairs[orgID] = append(kp.keypairs[orgID], k)
	}

	return nil
}

// Get returns the ClaimedKeyPair for the given public key id
func (kp *Keypairs) Get(publicKeyID *identity.ID) (*ClaimedKeyPair, error) {
	if publicKeyID == nil {
		return nil, errors.New("Invalid PublicKeyID provided to get keypair")
	}

	ckp, ok := kp.idIndex[*publicKeyID]
	if !ok {
		return nil, ErrPublicKeyNotFound
	}

	return &ckp, nil
}

// Select returns a keypair for the given type inside the specified
// organization. If a valid key (non-revoked) cannot be found an error is
// returned.
func (kp *Keypairs) Select(orgID *identity.ID, t primitive.KeyType) (*ClaimedKeyPair, error) {
	if orgID == nil {
		return nil, errors.New("Invalid OrgID Provided to select keypairs")
	}

	possible, ok := kp.keypairs[*orgID]
	if !ok {
		return nil, ErrMissingKeysForOrg
	}

	for _, k := range possible {
		if k.PublicKey.Body.KeyType == t && !k.Revoked() {
			return &k, nil
		}
	}

	return nil, ErrMissingValidKeypair
}

// All returns all keypairs including those which have been revoked.
func (kp *Keypairs) All() []ClaimedKeyPair {
	out := []ClaimedKeyPair{}
	for _, ckp := range kp.idIndex {
		out = append(out, ckp)
	}

	return out
}

// ClaimedKeyPair contains a public/private keypair, and all the Claims made
// against it (system and user signatures).
type ClaimedKeyPair struct {
	apitypes.PublicKeySegment
	PrivateKey *envelope.PrivateKey `json:"private_key"`
}

// KeyPairsClient represents the `/keypairs` registry endpoint, used for
// accessing users' signing and encryption keypairs.
type KeyPairsClient struct {
	client RoundTripper
}

// Create creates a new keypair on the registry.
//
// The keypair includes the user's public key, private key, and a self-signed
// claim on the public key.
//
// keys may be either signing or encryption keys.
func (k *KeyPairsClient) Create(ctx context.Context, pubKey *envelope.PublicKey,
	privKey *envelope.PrivateKey, claim *envelope.Claim) (*envelope.PublicKey,
	*envelope.PrivateKey, []envelope.Claim, error) {

	req := ClaimedKeyPair{
		PublicKeySegment: apitypes.PublicKeySegment{
			PublicKey: pubKey,
			Claims:    []envelope.Claim{*claim},
		},
		PrivateKey: privKey,
	}
	resp := ClaimedKeyPair{}
	err := k.client.RoundTrip(ctx, "POST", "/keypairs", nil, &req, &resp)
	if err != nil {
		return nil, nil, nil, err
	}

	return resp.PublicKey, resp.PrivateKey, resp.Claims, nil
}

// List returns all KeyPairs for the logged in user in the given, or all orgs
// if orgID is nil.
func (k *KeyPairsClient) List(ctx context.Context, orgID *identity.ID) (*Keypairs, error) {
	query := &url.Values{}
	if orgID != nil {
		query.Set("org_id", orgID.String())
	}

	var resp []ClaimedKeyPair
	err := k.client.RoundTrip(ctx, "GET", "/keypairs", query, nil, &resp)
	if err != nil {
		return nil, err
	}

	kp := NewKeypairs()
	err = kp.Add(resp...)
	if err != nil {
		return nil, err
	}

	return kp, nil
}
