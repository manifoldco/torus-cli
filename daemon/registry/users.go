package registry

import (
	"errors"
	"fmt"
	"log"

	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/envelope"
	"github.com/arigatomachine/cli/daemon/primitive"
)

type Users struct {
	client *Client
}

func (u *Users) GetSelf(token string) (*envelope.Unsigned, error) {
	req, err := u.client.NewTokenRequest(token, "GET", "/users/self", nil)
	if err != nil {
		log.Printf("Error making api request: %s", err)
		return nil, err
	}

	self := envelope.Unsigned{}
	_, err = u.client.Do(req, &self)
	if err != nil {
		log.Printf("Error making api request: %s", err)
		return nil, err
	}

	err = validateSelf(&self)
	if err != nil {
		log.Printf("Invalid user self: %s", err)
		return nil, err
	}

	return &self, nil
}

func validateSelf(s *envelope.Unsigned) error {
	if s.Version != 1 {
		return errors.New("version must be 1")
	}

	body := s.Body.(*primitive.User)

	if body == nil {
		return errors.New("missing body")
	}

	if body.Master == nil {
		return errors.New("missing master key section")
	}

	if body.Master.Alg != crypto.Triplesec {
		return fmt.Errorf("Unknown alg: %s", body.Master.Alg)
	}

	if len(*body.Master.Value) == 0 {
		return errors.New("Zero length master key found")
	}

	return nil
}
