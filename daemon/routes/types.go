package routes

import (
	"net/http"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/identity"
)

type keyPairGenerate struct {
	OrgID *identity.ID `json:"org_id"`
}

type errorMsg struct {
	Type  apitypes.ErrorType `json:"type"`
	Error []string           `json:"error"`
}

var notFoundError = &apitypes.Error{
	StatusCode: http.StatusNotFound,
	Type:       apitypes.NotFoundError,
	Err:        []string{"Not found"},
}
