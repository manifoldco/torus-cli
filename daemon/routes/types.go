package routes

import "github.com/arigatomachine/cli/daemon/registry"

type Login struct {
	Email      string `json:"email"`
	Passphrase string `json:"passphrase"`
}

type Version struct {
	Version string `json:"version"`
}

type Status struct {
	Token      bool `json:"token"`
	Passphrase bool `json:"passphrase"`
}

type KeyPairGenerate struct {
	OrgID *registry.ID `json:"org_id"`
}

type Error struct {
	Err     string `json:"error"`
	Message string `json:"message"`
}

const (
	TokenTypeLogin = "login"
	TokenTypeAuth  = "auth"
)
