package routes

import (
	"github.com/manifoldco/torus-cli/apitypes"
)

type errorMsg struct {
	Type  apitypes.ErrorType `json:"type"`
	Error []string           `json:"error"`
}

var notFoundError = &apitypes.Error{
	Type: apitypes.NotFoundError,
	Err:  []string{"Not found"},
}
