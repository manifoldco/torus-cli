package routes

import (
	"net/http"

	"github.com/manifoldco/torus-cli/apitypes"
)

type errorMsg struct {
	Type  apitypes.ErrorType `json:"type"`
	Error []string           `json:"error"`
}

var notFoundError = &apitypes.Error{
	StatusCode: http.StatusNotFound,
	Type:       apitypes.NotFoundError,
	Err:        []string{"Not found"},
}
