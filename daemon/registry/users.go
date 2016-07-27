package registry

import (
	"errors"
	"fmt"
	"log"

	"github.com/arigatomachine/cli/daemon/crypto"
)

type Users struct {
	client *Client
}

func (u *Users) GetSelf(token string) (*SelfResponse, error) {
	req, err := u.client.NewTokenRequest(token, "GET", "/users/self", nil)
	if err != nil {
		log.Printf("Error making api request: %s", err)
		return nil, err
	}

	self := SelfResponse{}
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

func validateSelf(s *SelfResponse) error {
	if s.Version != 1 {
		return errors.New("version must be 1")
	}

	if s.Body == nil {
		return errors.New("missing body")
	}

	if s.Body.Master == nil {
		return errors.New("missing master key section")
	}

	if s.Body.Master.Alg != crypto.Triplesec {
		return fmt.Errorf("Unknown alg: %s", s.Body.Master.Alg)
	}

	if len(*s.Body.Master.Value) == 0 {
		return errors.New("Zero length master key found")
	}

	return nil
}
