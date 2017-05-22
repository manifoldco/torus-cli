package apitypes

import (
	"time"

	"github.com/manifoldco/go-base64"
	"github.com/manifoldco/torus-cli/identity"
)

// MachineBootstrap represents a request by a Gatekeeper bootstrap client to
// create a machine with a given org based on AWS Bootstrap information
type MachineBootstrap struct {
	Name string `json:"name"`
	Org  string `json:"org_id"`
	Team string `json:"team_id"`
}

// BootstrapResponse represents the Response object returned to Gatekeeper bootstrap request
type BootstrapResponse struct {
	Token  identity.ID  `json:"token"`
	Secret base64.Value `json:"secret"`
	Error  string       `json:"error"`
}

// AWSBootstrapRequest represents a Bootstrap request from an AWS client
type AWSBootstrapRequest struct {
	Identity      []byte    `json:"identity"`
	Signature     []byte    `json:"signature"`
	ProvisionTime time.Time `json:"provision_time"`

	Machine MachineBootstrap `json:"machine"`
}
