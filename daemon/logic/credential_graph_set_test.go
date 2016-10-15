package logic

import (
	"testing"

	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/pathexp"
	"github.com/arigatomachine/cli/primitive"

	"github.com/arigatomachine/cli/daemon/registry"
)

var (
	id1 = mustID("04100000000000000000000000001")
	id2 = mustID("04100000000000000000000000010")
	id3 = mustID("04100000000000000000000000100")

	unset = "unset"
)

type cred struct {
	id    *identity.ID
	prev  *identity.ID
	state *string
	pe    *string
	name  *string
}

func mustID(raw string) *identity.ID {
	id, err := identity.DecodeFromString(raw)
	if err != nil {
		panic(err)
	}

	return &id
}

func mustPathExp(raw string) *pathexp.PathExp {
	pe, err := pathexp.Parse(raw)
	if err != nil {
		panic(err)
	}

	return pe
}

func buildGraph(rawPathExp string, version int, secrets ...cred) registry.CredentialGraph {
	pe := mustPathExp(rawPathExp)
	cg := &registry.CredentialGraphV2{
		KeyringSectionV2: registry.KeyringSectionV2{
			Keyring: &envelope.Signed{
				Version: 2,
				Body: &primitive.Keyring{
					BaseKeyring: primitive.BaseKeyring{
						PathExp:        pe,
						KeyringVersion: version,
					},
				},
			},
		},
	}

	for _, secret := range secrets {
		base := primitive.BaseCredential{
			Previous: secret.prev,
		}

		if secret.pe != nil {
			base.PathExp = mustPathExp(*secret.pe)
		}
		if secret.name != nil {
			base.Name = *secret.name
		}

		cred := envelope.Signed{
			ID:      secret.id,
			Version: 2,
			Body: &primitive.Credential{
				BaseCredential: base,
				State:          secret.state,
			},
		}
		cg.Credentials = append(cg.Credentials, cred)
	}

	return cg
}

func TestCredentialGraphSetAdd(t *testing.T) {
	cgs := newCredentialGraphSet()
	cg := buildGraph("/o/p/e/s/u/i", 1)
	cgs.Add(cg)

	if len(cgs.graphs) != 1 {
		t.Error("CredentialGraph was not added")
	}
}

func TestCredentialGraphSetHead(t *testing.T) {
	t.Run("no match", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		cgs.Add(buildGraph("/o/p/e/s1/u/*", 2))
		cgs.Add(buildGraph("/o/p/e/s1/u/*", 1))

		out, err := cgs.Head(mustPathExp("/o/p/e/s2/u/i"))
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if out != nil {
			t.Error("Head CredentialGraph found when there should not be one")
		}
	})

	t.Run("match", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		cgs.Add(buildGraph("/o/p/e/s/u/*", 2))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 3))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 1))

		out, err := cgs.Head(mustPathExp("/o/p/e/s/u/i"))
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if out == nil {
			t.Fatal("Head CredentialGraph not found when there should be one")
		}

		if out.KeyringVersion() != 3 {
			t.Error("Wrong CredentialGraph version returned")
		}
	})

}

func TestCredentialGraphSetHeadCredential(t *testing.T) {
	t.Run("no match", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		pe := "/o/p/e/s/u/i"
		name := "other"
		cgs.Add(buildGraph("/o/p/e/s/u/*", 2, cred{id: id3, prev: id2, pe: &pe, name: &name}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 1, cred{id: id2, pe: &pe, name: &name}))

		out, err := cgs.HeadCredential(mustPathExp(pe), "nomatchie")
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if out != nil {
			t.Error("Head credential found when there should not be one")
		}
	})

	t.Run("in head CredentialGraph", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		pe := "/o/p/e/s/u/i"
		name := "cred"

		cgs.Add(buildGraph("/o/p/e/s/u/*", 2, cred{id: id3, prev: id2, pe: &pe, name: &name}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 1, cred{id: id2, pe: &pe, name: &name}))

		out, err := cgs.HeadCredential(mustPathExp(pe), name)
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if out == nil {
			t.Fatal("Head credential not found when there should be one")
		}

		if out.ID != id3 {
			t.Error("Wrong head credential found. wanted:", id3, "got:", out.ID)
		}
	})

	t.Run("in older CredentialGraph", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		pe := "/o/p/e/s/u/i"
		name := "cred"
		othername := "othercred"

		cgs.Add(buildGraph("/o/p/e/s/u/*", 3, cred{id: id3, pe: &pe, name: &othername}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 2, cred{id: id2, prev: id1, pe: &pe, name: &name}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 1, cred{id: id1, pe: &pe, name: &name}))

		out, err := cgs.HeadCredential(mustPathExp(pe), name)
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if out == nil {
			t.Fatal("Head credential not found when there should be one")
		}

		if out.ID != id2 {
			t.Error("Wrong head credential found. wanted:", id2, "got:", out.ID)
		}
	})
}
