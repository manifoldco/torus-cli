// Package bootstrap provides authentication actions for the bootstrap process
package bootstrap

import (
	"fmt"

	"github.com/manifoldco/torus-cli/gatekeeper/apitypes"
	"github.com/manifoldco/torus-cli/gatekeeper/bootstrap/aws"
)

const (
	// AWSPublic is Amazon's Public Cloud Provider
	AWSPublic = "aws"
)

// Handler the function returned by the Bootstrap server that can
// be used to bootstrap some information (e.g. machine) for that service.
type Handler func(url, name, org, role string) (*apitypes.BootstrapResponse, error)

// New returns a new Provider based on the given bootstrap.Type
func New(t string) (Handler, error) {
	switch t {
	case AWSPublic:
		return aws.Bootstrap, nil

	default:
		return nil, fmt.Errorf("Invalid Provider: %s", t)
	}
}
