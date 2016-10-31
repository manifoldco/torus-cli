package apitypes

import (
	"github.com/manifoldco/torus-cli/envelope"
)

// PublicKeySegment represents a sub section of a claimtree targeting a
// specific public key and it's claims.
type PublicKeySegment struct {
	Key    *envelope.Signed  `json:"public_key"`
	Claims []envelope.Signed `json:"claims"`
}
