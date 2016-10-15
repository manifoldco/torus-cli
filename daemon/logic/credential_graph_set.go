package logic

import (
	"errors"
	"sort"

	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/pathexp"
	"github.com/arigatomachine/cli/primitive"

	"github.com/arigatomachine/cli/daemon/registry"
)

// credentialGraphSet holds credential graphs and answers questions about their
// liveliness/reachability.
//
// CredentialGraphs are identified by path expression, and versioned.
// credenialGraphSet handles multiple graphs by path expression, and multiple
// versions within each path expression.
type credentialGraphSet struct {
	graphs map[pathexp.PathExp][]registry.CredentialGraph
}

func newCredentialGraphSet() *credentialGraphSet {
	return &credentialGraphSet{
		graphs: make(map[pathexp.PathExp][]registry.CredentialGraph),
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

		cgs.graphs[*pe] = append(cgs.graphs[*pe], c)
	}

	return nil
}

// Head returns the most recent version of a CredentialGraph that would contain
// the given PathExp.
func (cgs *credentialGraphSet) Head(pe *pathexp.PathExp) (registry.CredentialGraph, error) {

	gpe, err := pe.WithInstance("*")
	if err != nil {
		return nil, err
	}

	graphs, ok := cgs.graphs[*gpe]
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

	graphs, ok := cgs.graphs[*gpe]
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
				head = &cred
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
