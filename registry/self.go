package registry

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
)

var errUnknownSessionType = errors.New("Unknown session type")

// SelfClient represents the registry `/self` endpoints.
type SelfClient struct {
	client RoundTripper
}

type rawSelf struct {
	*apitypes.Self

	// Shadow over the pure self value
	Identity json.RawMessage `json:"identity"`
	Auth     json.RawMessage `json:"auth"`
}

// Get returns the current identities associated with this token
func (s *SelfClient) Get(ctx context.Context, token string) (*apitypes.Self, error) {
	req, err := s.client.NewRequest("GET", "/self", nil, nil)
	replaceAuthToken(req, token)
	if err != nil {
		log.Printf("Error making Self request: %s", err)
		return nil, err
	}

	raw := &rawSelf{}
	_, err = s.client.Do(ctx, req, raw)
	if err != nil {
		return nil, err
	}

	self := raw.Self
	switch raw.Type {
	case apitypes.UserSession:
		self.Identity = &envelope.User{}
		self.Auth = &envelope.User{}
	case apitypes.MachineSession:
		self.Identity = &envelope.Machine{}
		self.Auth = &envelope.MachineToken{}
	default:
		return nil, errUnknownSessionType
	}

	err = json.Unmarshal(raw.Identity, self.Identity)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(raw.Auth, self.Auth)
	if err != nil {
		return nil, err
	}

	return self, nil
}
