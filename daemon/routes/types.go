package routes

import (
	"net/http"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/identity"
)

type keyPairGenerate struct {
	OrgID *identity.ID `json:"org_id"`
}

type machineCreate struct {
	Name  string       `json:"name"`
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
