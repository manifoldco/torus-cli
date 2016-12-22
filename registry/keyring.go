package registry

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/url"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// ErrMemberNotFound is returned when a keyring member find call fails.
var ErrMemberNotFound = errors.New("keyring membership not found")

// KeyringClient represents the `/keyrings` registry end point for accessing
// keyrings the user or machine belong too.
type KeyringClient struct {
	client  RoundTripper
	Members *KeyringMembersClient
}

// KeyringSection is the shared interface between different KeyringSection
// versions.
type KeyringSection interface {
	GetKeyring() envelope.KeyringInf
	KeyringVersion() int
	FindMember(*identity.ID) (*primitive.KeyringMember, *primitive.MEKShare, error)
	HasRevocations() bool
}

// KeyringSectionV1 represents a section of the CredentialGraph only pertaining to
// a keyring and it's membership.
type KeyringSectionV1 struct {
	Keyring *envelope.KeyringV1        `json:"keyring"`
	Members []envelope.KeyringMemberV1 `json:"members"`
}

// GetKeyring returns the Keyring object in this KeyringSection
func (k *KeyringSectionV1) GetKeyring() envelope.KeyringInf {
	return k.Keyring
}

// KeyringVersion returns the version of the keyring itself (not its schema).
func (k *KeyringSectionV1) KeyringVersion() int {
	return k.Keyring.Body.KeyringVersion
}

// FindMember returns the membership and mekshare for the given user id.
// The data is returned in V2 format.
func (k *KeyringSectionV1) FindMember(id *identity.ID) (*primitive.KeyringMember, *primitive.MEKShare, error) {
	var krm *primitive.KeyringMember
	var mekshare *primitive.MEKShare
	for _, m := range k.Members {
		if *m.Body.OwnerID == *id {
			krm = &primitive.KeyringMember{
				OrgID:           m.Body.OrgID,
				KeyringID:       m.Body.KeyringID,
				OwnerID:         m.Body.OwnerID,
				PublicKeyID:     m.Body.PublicKeyID,
				EncryptingKeyID: m.Body.EncryptingKeyID,
			}

			mekshare = &primitive.MEKShare{
				Key: m.Body.Key,
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
	Keyring *envelope.Keyring             `json:"keyring"`
	Members []KeyringMember               `json:"members"`
	Claims  []envelope.KeyringMemberClaim `json:"claims"`
}

// GetKeyring returns the Keyring object in this KeyringSection
func (k *KeyringSectionV2) GetKeyring() envelope.KeyringInf {
	return k.Keyring
}

// KeyringVersion returns the version of the keyring itself (not its schema).
func (k *KeyringSectionV2) KeyringVersion() int {
	return k.Keyring.Body.KeyringVersion
}

// FindMember returns the membership and mekshare for the given user id.
//
// An owner (user/machine token) may have multiple memberships, one per
// encryption key. There will only be one unrevoked membership.
// Either this unrevoked membership will be returned, or the result will error
// with ErrMemberNotFound.
func (k *KeyringSectionV2) FindMember(id *identity.ID) (*primitive.KeyringMember, *primitive.MEKShare, error) {
	var krm *primitive.KeyringMember
	var mekshare *primitive.MEKShare

outerLoop:
	for _, m := range k.Members {
		if *m.Member.Body.OwnerID == *id {
			// We've found the right owner. Now see if this membership is
			// unrevoked.
			// A revocation is always terminal for a claim chain, so if there's
			// any revocations for this membership, we know it is invalid.
			for _, c := range k.Claims {
				if *c.Body.KeyringMemberID == *m.Member.ID && c.Body.ClaimType == primitive.RevocationClaimType {
					continue outerLoop
				}

			}

			krm = m.Member.Body
			// We never get the MEKShare for another user returned.
			if m.MEKShare != nil {
				mekshare = m.MEKShare.Body
			}
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
		if claim.Body.ClaimType == primitive.RevocationClaimType {
			return true
		}
	}
	return false
}

// KeyringMember holds membership information for v2 keyrings. In v2, a user
// can have their master encryption key share removed.
type KeyringMember struct {
	Member   *envelope.KeyringMember `json:"member"`
	MEKShare *envelope.MEKShare      `json:"mekshare"`
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
		Keyring *envelope.Signed              `json:"keyring"`
		Members json.RawMessage               `json:"members"`
		Claims  []envelope.KeyringMemberClaim `json:"claims"`
	}{}

	_, err = k.client.Do(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	converted := make([]KeyringSection, len(resp))
	for i, k := range resp {
		if k.Keyring.Version == 1 {
			kre := &envelope.KeyringV1{
				ID:        k.Keyring.ID,
				Version:   k.Keyring.Version,
				Signature: k.Keyring.Signature,
				Body:      k.Keyring.Body.(*primitive.KeyringV1),
			}

			s := KeyringSectionV1{
				Keyring: kre,
			}
			err := json.Unmarshal(k.Members, &s.Members)
			if err != nil {
				return nil, err
			}
			converted[i] = &s
		} else {
			kre := &envelope.Keyring{
				ID:        k.Keyring.ID,
				Version:   k.Keyring.Version,
				Signature: k.Keyring.Signature,
				Body:      k.Keyring.Body.(*primitive.Keyring),
			}
			s := KeyringSectionV2{
				Keyring: kre,
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
