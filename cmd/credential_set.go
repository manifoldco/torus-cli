package cmd

import (
	"errors"
	"sort"

	"github.com/manifoldco/torus-cli/apitypes"
)

// credentialSet represents a set of credentials.
// It ensures credentials are unique by name. The credential with the most
// specific PathExp wins. If two credentials have the same specificity level,
// the first one added wins.
//
// credentials are returned in lexicographically sorted order, by name.
type credentialSet map[string]*apitypes.CredentialEnvelope

// Add adds a credential to the credentialSet, replacing an existing credential
// of the same name, if the new credential is more specific.
//
// If the added credential is an unset value, it is ignored.
func (c credentialSet) Add(cred apitypes.CredentialEnvelope) error {
	if (*cred.Body).GetValue() == nil {
		return errors.New("Cannot add an unset credential")
	}

	name := (*cred.Body).GetName()
	if existing, ok := c[name]; ok {
		// The new credential is either as specific, or less specific than
		// the existing one. Keep the existing one.
		eBody := *existing.Body
		if (*cred.Body).GetPathExp().CompareSpecificity(eBody.GetPathExp()) != 1 {
			return nil
		}
	}

	c[name] = &cred
	return nil
}

// ToSlice returns a slice of the credentials in the set, in lexicographically
// sorted order by name.
func (c credentialSet) ToSlice() []apitypes.CredentialEnvelope {
	creds := make([]apitypes.CredentialEnvelope, len(c))
	i := 0
	for _, cred := range c {
		creds[i] = *cred
		i++
	}

	sort.Sort(credSorter(creds))
	return creds
}

// credSorter implements sort.Interface, for sorting credentials
// lexicographically by name.
type credSorter []apitypes.CredentialEnvelope

func (c credSorter) Len() int      { return len(c) }
func (c credSorter) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c credSorter) Less(i, j int) bool {
	a := *c[i].Body
	b := *c[j].Body
	return a.GetName() < b.GetName()
}
