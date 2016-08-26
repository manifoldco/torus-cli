// Package apitypes defines types shared between the daemon and its api client.
package apitypes

import "strings"

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
