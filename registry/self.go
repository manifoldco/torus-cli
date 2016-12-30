package registry

import (
	"context"
	"errors"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/primitive"
)

var errUnknownSessionType = errors.New("Unknown session type")

// SelfClient represents the registry `/self` endpoints.
type SelfClient struct {
	client RequestDoer
}

type rawSelf struct {
	*apitypes.Self

	// Shadow over the pure self value
	Identity *envelope.Unsigned `json:"identity"`
	Auth     *envelope.Unsigned `json:"auth"`
}

// Get returns the current identities associated with this token
func (s *SelfClient) Get(ctx context.Context, token string) (*apitypes.Self, error) {
	raw := &rawSelf{}
	err := tokenRoundTrip(ctx, s.client, token, "GET", "/self", nil, nil, raw)
	if err != nil {
		return nil, err
	}

	self := raw.Self
	switch raw.Type {
	case apitypes.UserSession:
		user, err := envelope.ConvertUser(raw.Identity)
		if err != nil {
			return nil, err
		}

		self.Identity = user
		self.Auth = user
	case apitypes.MachineSession:
		self.Identity = &envelope.Machine{
			ID:      raw.Identity.ID,
			Version: raw.Identity.Version,
			Body:    raw.Identity.Body.(*primitive.Machine),
		}
		self.Auth = &envelope.MachineToken{
			ID:      raw.Auth.ID,
			Version: raw.Auth.Version,
			Body:    raw.Auth.Body.(*primitive.MachineToken),
		}
	default:
		return nil, errUnknownSessionType
	}

	return self, nil
}
