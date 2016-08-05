package routes

import "github.com/arigatomachine/cli/daemon/identity"

type errorType string

const (
	badRequestError   = "bad_request"
	unauthorizedError = "unauthorized"
	notFoundError     = "not_found"

	internalServerError = "internal_server"
	notImplementedError = "not_implemented"
)

type login struct {
	Email      string `json:"email"`
	Passphrase string `json:"passphrase"`
}

type version struct {
	Version string `json:"version"`
}

type status struct {
	Token      bool `json:"token"`
	Passphrase bool `json:"passphrase"`
}

type keyPairGenerate struct {
	OrgID *identity.ID `json:"org_id"`
}

type errorMsg struct {
	Type  errorType `json:"type"`
	Error string    `json:"error"`
}
