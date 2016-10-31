package registry

import (
	"context"
	"log"

	"github.com/manifoldco/torus-cli/apitypes"
)

// SelfClient represents the registry `/self` endpoints.
type SelfClient struct {
	client *Client
}

// Get returns the current identities associated with this token
func (s *SelfClient) Get(ctx context.Context, token string) (*apitypes.Self, error) {
	req, err := s.client.NewTokenRequest(token, "GET", "/self", nil, nil)
	if err != nil {
		log.Printf("Error making Self request: %s", err)
		return nil, err
	}

	self := &apitypes.Self{}
	_, err = s.client.Do(ctx, req, self)
	if err != nil {
		return nil, err
	}

	return self, nil
}
