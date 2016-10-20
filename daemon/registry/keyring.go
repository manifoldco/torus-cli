package registry

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/url"

	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"
)

// ErrMemberNotFound is returned when a keyring member find call fails.
var ErrMemberNotFound = errors.New("Keyring membership not found.")

// KeyringClient represents the `/keyrings` registry end point for accessing
// keyrings the user or machine belong too.
type KeyringClient struct {
	client  *Client
	Members *KeyringMembersClient
}

// KeyringSection is the shared interface between different KeyringSection
// versions.
type KeyringSection interface {
	GetKeyring() *envelope.Signed
	FindMember(*identity.ID) (*primitive.KeyringMember, *primitive.MEKShare, error)
	HasRevocations() bool
}

// KeyringSectionV1 represents a section of the CredentialGraph only pertaining to
// a keyring and it's membership.
type KeyringSectionV1 struct {
	Keyring *envelope.Signed  `json:"keyring"`
	Members []envelope.Signed `json:"members"`
}

// GetKeyring returns the Keyring object in this KeyringSection
func (k *KeyringSectionV1) GetKeyring() *envelope.Signed {
	return k.Keyring
}

// FindMember returns the membership and mekshare for the given user id.
// The data is returned in V2 format.
func (k *KeyringSectionV1) FindMember(id *identity.ID) (*primitive.KeyringMember, *primitive.MEKShare, error) {
	var krm *primitive.KeyringMember
	var mekshare *primitive.MEKShare
	for _, m := range k.Members {
		mbody := m.Body.(*primitive.KeyringMemberV1)
		if *mbody.OwnerID == *id {
			krm = &primitive.KeyringMember{
				OrgID:           mbody.OrgID,
				KeyringID:       mbody.KeyringID,
				OwnerID:         mbody.OwnerID,
				PublicKeyID:     mbody.PublicKeyID,
				EncryptingKeyID: mbody.EncryptingKeyID,
			}

			mekshare = &primitive.MEKShare{
				Key: mbody.Key,
			}
			break
		}
	}

	if krm == nil {
		return nil, nil, ErrMemberNotFound
	}

	return krm, mekshare, nil
}

// HasRevocations indicates that a Keyring holds revoked user keys. We don't
// track in V1 so it is always false.
func (KeyringSectionV1) HasRevocations() bool {
	return false
}

// KeyringSectionV2 represents a Keyring and its members.
type KeyringSectionV2 struct {
	Keyring *envelope.Signed  `json:"keyring"`
	Members []KeyringMember   `json:"members"`
	Claims  []envelope.Signed `json:"claims"`
}

// GetKeyring returns the Keyring object in this KeyringSection
func (k *KeyringSectionV2) GetKeyring() *envelope.Signed {
	return k.Keyring
}

// FindMember returns the membership and mekshare for the given user id.
func (k *KeyringSectionV2) FindMember(id *identity.ID) (*primitive.KeyringMember, *primitive.MEKShare, error) {
	var krm *primitive.KeyringMember
	var mekshare *primitive.MEKShare
	for _, m := range k.Members {
		mbody := m.Member.Body.(*primitive.KeyringMember)
		if *mbody.OwnerID == *id {
			krm = mbody
			mekshare = m.MEKShare.Body.(*primitive.MEKShare)
			break
		}
	}

	if krm == nil {
		return nil, nil, ErrMemberNotFound
	}

	return krm, mekshare, nil
}

// HasRevocations indicates that a Keyring holds revoked user keys.
func (k *KeyringSectionV2) HasRevocations() bool {
	for _, claim := range k.Claims {
		if claim.Body.(*primitive.KeyringMemberClaim).ClaimType == primitive.RevocationClaimType {
			return true
		}
	}
	return false
}

// KeyringMember holds membership information for v2 keyrings. In v2, a user
// can have their master encryption key share removed.
type KeyringMember struct {
	Member   *envelope.Signed `json:"member"`
	MEKShare *envelope.Signed `json:"mekshare"`
}

// List retrieves an array of KeyringSections from the registry.
func (k *KeyringClient) List(ctx context.Context, orgID *identity.ID,
	ownerID *identity.ID) ([]KeyringSection, error) {

	query := &url.Values{}

	if orgID != nil {
		query.Set("org_id", orgID.String())
	}
	if ownerID != nil {
		query.Set("owner_id", ownerID.String())
	}

	req, err := k.client.NewRequest("GET", "/keyrings", query, nil)
	if err != nil {
		log.Printf("Error building http request for GET /keyrings: %s", err)
		return nil, err
	}

	resp := []struct {
		Keyring *envelope.Signed  `json:"keyring"`
		Members json.RawMessage   `json:"members"`
		Claims  []envelope.Signed `json:"claims"`
	}{}

	_, err = k.client.Do(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	converted := make([]KeyringSection, len(resp))
	for i, k := range resp {
		if k.Keyring.Version == 1 {
			s := KeyringSectionV1{
				Keyring: k.Keyring,
			}
			err := json.Unmarshal(k.Members, &s.Members)
			if err != nil {
				return nil, err
			}
			converted[i] = &s
		} else {
			s := KeyringSectionV2{
				Keyring: k.Keyring,
				Claims:  k.Claims,
			}
			err := json.Unmarshal(k.Members, &s.Members)
			if err != nil {
				return nil, err
			}
			converted[i] = &s
		}
	}

	return converted, nil
}
