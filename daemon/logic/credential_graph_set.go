package logic

import (
	"errors"
	"sort"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/pathexp"
	"github.com/manifoldco/torus-cli/primitive"

	"github.com/manifoldco/torus-cli/daemon/registry"
)

// credentialGraphSet holds credential graphs and answers questions about their
// liveliness/reachability.
//
// CredentialGraphs are identified by path expression, and versioned.
// credenialGraphSet handles multiple graphs by path expression, and multiple
// versions within each path expression.
type credentialGraphSet struct {
	graphs map[string][]registry.CredentialGraph
}

func newCredentialGraphSet() *credentialGraphSet {
	return &credentialGraphSet{
		graphs: make(map[string][]registry.CredentialGraph),
	}
}

func (cgs *credentialGraphSet) Add(graphs ...registry.CredentialGraph) error {
	for _, c := range graphs {
		var pe *pathexp.PathExp
		env := c.GetKeyring()
		switch env.Version {
		case 1:
			pe = env.Body.(*primitive.KeyringV1).PathExp
		case 2:
			pe = env.Body.(*primitive.Keyring).PathExp

		default:
			return errors.New("Unknown keyring version")
		}

		cgs.graphs[pe.String()] = append(cgs.graphs[pe.String()], c)
	}

	return nil
}

func (credentialGraphSet) activeCreds(parents []identity.ID,
	graph registry.CredentialGraph) ([]envelope.Signed, []identity.ID, error) {

	creds := graph.GetCredentials()

	// maybeActive is a set of potentially active credentials.
	// It will never contain unset credentials as they can't be
	// active.
	maybeActive := make(map[identity.ID]envelope.Signed, len(creds))
	unchecked := make([]identity.ID, 0, len(creds))
	for i := range creds {
		var parent *identity.ID
		cred := creds[i]
		switch cred.Version {
		case 1:
			parent = cred.Body.(*primitive.CredentialV1).Previous
			maybeActive[*cred.ID] = cred
		case 2:
			body := cred.Body.(*primitive.Credential)
			parent = body.Previous

			if body.State == nil || *body.State != "unset" {
				maybeActive[*cred.ID] = cred
			}
		default:
			return nil, parents, errors.New("Unknown credential version")
		}

		if parent != nil {
			unchecked = append(unchecked, *parent)
		}
	}

	// unchecked will contain the existing parents, and the parents
	// of all credentials in this version of the keyring
	unchecked = append(unchecked, parents...)
	parents = []identity.ID{}

	for len(unchecked) > 0 {
		var id identity.ID
		id, unchecked = unchecked[0], unchecked[1:]

		if _, ok := maybeActive[id]; ok {
			delete(maybeActive, id)
		} else {
			parents = append(parents, id)
		}
	}

	var activeCreds []envelope.Signed
	for _, cred := range maybeActive {
		activeCreds = append(activeCreds, cred)
	}
	return activeCreds, parents, nil
}

// Active returns a slice of CredentialGraphs that contain credentials that
// are still reachable.
// A credential is reachable if a newer version of the CredentialGraph has not
// replaced its value, and it is not an `unset` Credential.
func (cgs *credentialGraphSet) Active() ([]registry.CredentialGraph, error) {
	active := make([]registry.CredentialGraph, 0, len(cgs.graphs))

	for _, graphs := range cgs.graphs {

		// parents is the slice of IDs already seen in Previous fields.
		// they cannot be active as they have been overwritten.
		var parents []identity.ID

		sort.Sort(graphSorter(graphs))
		for headOffset, graph := range graphs {
			var activeCreds []envelope.Signed
			var err error
			activeCreds, parents, err = cgs.activeCreds(parents, graph)
			if err != nil {
				return nil, err
			}

			// the most recent version of a keyring is always active
			// (it could be filled entirely with unset values)
			if len(activeCreds) > 0 || headOffset == 0 {
				active = append(active, graph)
			}
		}
	}

	return active, nil
}

// NeedRotation returns a slice of Credentials that need to be rotated.
//
// A Credential needs to be rotated if its most recent set version is in a
// CredentialGraph version that contains a revocation of a user's share to
// that Keyring.
func (cgs *credentialGraphSet) NeedRotation() ([]envelope.Signed, error) {
	var needRotation []envelope.Signed

	for _, graphs := range cgs.graphs {
		var parents []identity.ID

		sort.Sort(graphSorter(graphs))
		for _, graph := range graphs {
			var activeCreds []envelope.Signed
			var err error
			activeCreds, parents, err = cgs.activeCreds(parents, graph)
			if err != nil {
				return nil, err
			}

			if len(activeCreds) > 0 && graph.HasRevocations() {
				needRotation = append(needRotation, activeCreds...)
			}
		}
	}

	return needRotation, nil
}

// Head returns the most recent version of a CredentialGraph that would contain
// the given PathExp.
func (cgs *credentialGraphSet) Head(pe *pathexp.PathExp) (registry.CredentialGraph, error) {

	gpe, err := pe.WithInstance("*")
	if err != nil {
		return nil, err
	}

	graphs, ok := cgs.graphs[gpe.String()]
	if !ok {
		return nil, nil
	}

	sort.Sort(graphSorter(graphs))
	return graphs[0], nil
}

// HeadCredential returns the most recent version of a Credential that shares
// the provided PathExp and Name
//
// A Head Credential need not be in the Head of the CredentialGraph.
func (cgs *credentialGraphSet) HeadCredential(pe *pathexp.PathExp, name string) (*envelope.Signed, error) {
	var head *envelope.Signed
	version := -1

	gpe, err := pe.WithInstance("*")
	if err != nil {
		return nil, err
	}

	graphs, ok := cgs.graphs[gpe.String()]
	if !ok {
		return nil, nil
	}

	sort.Sort(graphSorter(graphs))
	for _, graph := range graphs {
		creds := graph.GetCredentials()
		for _, cred := range creds {
			base, err := baseCredential(&cred)
			if err != nil {
				return nil, err
			}

			if base.PathExp.Equal(pe) && base.Name == name && base.CredentialVersion > version {
				env := cred
				head = &env
				version = base.CredentialVersion
			}
		}
	}

	return head, nil
}

// graphSorter implements sort.Interface, for sorting CredentialGraphs
// by version in decreasing order
type graphSorter []registry.CredentialGraph

func (g graphSorter) Len() int           { return len(g) }
func (g graphSorter) Swap(i, j int)      { g[i], g[j] = g[j], g[i] }
func (g graphSorter) Less(i, j int) bool { return g[i].KeyringVersion() > g[j].KeyringVersion() }
