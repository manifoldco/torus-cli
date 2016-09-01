// Package apitypes defines types shared between the daemon and its api client.
package apitypes

import (
	"strings"

	"github.com/arigatomachine/cli/primitive"
)

// ErrorType represents the string error types that the daemon and registry can
// return.
type ErrorType string

// These are the possible error types.
const (
	BadRequestError     = "bad_request"
	UnauthorizedError   = "unauthorized"
	NotFoundError       = "not_found"
	InternalServerError = "internal_server"
	NotImplementedError = "not_implemented"
)

// Error represents standard formatted API errors from the daemon or registry.
type Error struct {
	StatusCode int

	Type string   `json:"type"`
	Err  []string `json:"error"`
}

// Error implements the error interface for formatted API errors.
func (e *Error) Error() string {
	return e.Type + ": " + strings.Join(e.Err, " ")
}

// Version contains the release version of the daemon.
type Version struct {
	Version string `json:"version"`
}

// SessionStatus contains details about the user's daemon session.
type SessionStatus struct {
	Token      bool `json:"token"`
	Passphrase bool `json:"passphrase"`
}

// Login contains the required details for logging in to the api and daemon.
type Login struct {
	Email      string `json:"email"`
	Passphrase string `json:"passphrase"`
}

// Profile contains the fields in the response for the profiles endpoint
type Profile struct {
	ID   string `json:"id"`
	Body *struct {
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"body"`
}

// OrgInvite contains information for sending an Org invite
type OrgInvite struct {
	ID      string               `json:"id"`
	Version int                  `json:"version"`
	Body    *primitive.OrgInvite `json:"body"`
}
