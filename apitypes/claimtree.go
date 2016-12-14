package apitypes

import (
	"errors"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/primitive"
)

// ErrClaimCycleFound is returned when a cycle is found within the claims.
// this *should* be impossible, as they are signed.
var ErrClaimCycleFound = errors.New("Cycle detected in signed claims")

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

// HeadClaim returns the most recent Claim made against this PublicKey
func (pks *PublicKeySegment) HeadClaim() (*envelope.Signed, error) {
	// The head claim is the one that is not the previous claim of any others
outerLoop:
	for _, c1 := range pks.Claims {
		for _, c2 := range pks.Claims {
			if *c2.Body.(*primitive.Claim).Previous == *c1.ID {
				// Something else is newer than c1
				continue outerLoop
			}

		}
		// nothing is newer than c1. return it
		return &c1, nil
	}

	return nil, ErrClaimCycleFound
}
