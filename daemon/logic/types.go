package logic

import (
	"github.com/arigatomachine/cli/identity"
)

// PlaintextCredentialEnvelope is an unencrypted credential object
type PlaintextCredentialEnvelope struct {
	ID      *identity.ID         `json:"id"`
	Version uint8                `json:"version"`
	Body    *PlaintextCredential `json:"body"`
}

// PlaintextCredential is the body of an unencrypted Credential
type PlaintextCredential struct {
	Name      string       `json:"name"`
	OrgID     *identity.ID `json:"org_id"`
	PathExp   string       `json:"pathexp"`
	ProjectID *identity.ID `json:"project_id"`
	Value     string       `json:"value"`
}
