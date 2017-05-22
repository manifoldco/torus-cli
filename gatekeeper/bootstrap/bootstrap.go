// Package bootstrap provides authentication actions for the bootstrap process
package bootstrap

import (
	"fmt"

	"github.com/manifoldco/torus-cli/gatekeeper/apitypes"
	"github.com/manifoldco/torus-cli/gatekeeper/bootstrap/aws"
)

const (
	// AWSPubic is Amazon's Public Cloud Provider
	AWSPublic = "aws"
)

// Type is the bootstrap Provider type
type Type string

// Provider is an interface that provides bootstrapping and verification of a role
type Provider interface {
	// Bootstrap provides the main bootstrapping functions for the auth Provider
	Bootstrap(url, name, org, role string) (*apitypes.BootstrapResponse, error)

	// Verify does backend verification on the bootstrap Provider
	Verify() error
}

// New returns a new Provider based on the given bootstrap.Type
func New(t Type) (Provider, error) {
	switch string(t) {
	case AWSPublic:
		return aws.New(), nil

	default:
		return nil, fmt.Errorf("Invalid Provider: %s", t)
	}
}
