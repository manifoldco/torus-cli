package routes

import "github.com/arigatomachine/cli/daemon/identity"

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
	Err     string `json:"error"`
	Message string `json:"message"`
}
