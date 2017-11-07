package logic

import (
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/registry"
)

// credentialGraphKeyIndex holds credential graphs and indexes them by the
// encryption key used to include the given identity into the ring as a member.
type credentialGraphKeyIndex struct {
	authID identity.ID

	graphs map[identity.ID][]registry.CredentialGraph
}

func newCredentialGraphKeyIndex(authID identity.ID) *credentialGraphKeyIndex {
	return &credentialGraphKeyIndex{
		authID: authID,
		graphs: make(map[identity.ID][]registry.CredentialGraph),
	}
}

func (cgi *credentialGraphKeyIndex) Add(graphs ...registry.CredentialGraph) error {
	for _, g := range graphs {
		krm, _, err := g.FindMember(&cgi.authID)
		if err != nil {
			return err
		}

		id := *(krm.EncryptingKeyID)
		if _, ok := cgi.graphs[id]; !ok {
			cgi.graphs[id] = []registry.CredentialGraph{g}
			continue
		}

		cgi.graphs[id] = append(cgi.graphs[id], g)
	}

	return nil
}

func (cgi *credentialGraphKeyIndex) GetIndex() map[identity.ID][]registry.CredentialGraph {
	return cgi.graphs
}
