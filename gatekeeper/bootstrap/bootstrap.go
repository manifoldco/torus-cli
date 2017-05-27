// Package bootstrap provides authentication actions for the bootstrap process
package bootstrap

import (
	"fmt"

	"github.com/manifoldco/torus-cli/gatekeeper/apitypes"
	"github.com/manifoldco/torus-cli/gatekeeper/bootstrap/aws"
)

// Provider represents the Provider type for bootstrapping
type Provider string

const (
	// AWSPublic is Amazon's Public Cloud Provider
	AWSPublic Provider = "aws"
)

// Do will execute the bootstrap request for the given provider
func Do(provider Provider, url, name, org, role string) (*apitypes.BootstrapResponse, error) {
	switch provider {
	case AWSPublic:
		return aws.Bootstrap(url, name, org, role)

	default:
		return nil, fmt.Errorf("invalid provider: %s", provider)
	}
}
