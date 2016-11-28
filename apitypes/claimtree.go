package apitypes

import (
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/primitive"
)

// PublicKeySegment represents a sub section of a claimtree targeting a
// specific public key and it's claims.
type PublicKeySegment struct {
	PublicKey *envelope.Signed  `json:"public_key"`
	Claims    []envelope.Signed `json:"claims"`
}

// Revoked returns a bool indicating if any revocation claims exist against this
// PublicKey
func (pks *PublicKeySegment) Revoked() bool {
	for _, claim := range pks.Claims {
		if claim.Body.(*primitive.Claim).ClaimType == primitive.RevocationClaimType {
			return true
		}
	}

	return false
}
