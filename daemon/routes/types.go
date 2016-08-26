package routes

import (
	"github.com/arigatomachine/cli/apitypes"

	"github.com/arigatomachine/cli/daemon/identity"
)

type keyPairGenerate struct {
	OrgID *identity.ID `json:"org_id"`
}

type errorMsg struct {
	Type  apitypes.ErrorType `json:"type"`
	Error []string           `json:"error"`
}
