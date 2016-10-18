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

func buildGraphWithRevocation(rawPathExp string, version int, secrets ...cred) registry.CredentialGraph {
	cg := buildGraph(rawPathExp, version, secrets...)

	cg.(*registry.CredentialGraphV2).Claims = []envelope.Signed{
		{Body: &primitive.KeyringMemberClaim{
			ClaimType: primitive.RevocationClaimType,
		}},
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

func TestCredentialGraphSetActive(t *testing.T) {
	t.Run("Many PathExps", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		cgs.Add(buildGraph("/o/p/e/s/u/i1", 1, cred{id: id1}))
		cgs.Add(buildGraph("/o/p/e/s/u/i2", 1, cred{id: id2}))

		active, err := cgs.Active()
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if len(active) != 2 {
			t.Fail()
		}
	})

	t.Run("Multi version no shadow", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		cgs.Add(buildGraph("/o/p/e/s/u/*", 2, cred{id: id1}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 3, cred{id: id2}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 1, cred{id: id3}))

		active, err := cgs.Active()
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if len(active) != 3 {
			t.Fail()
		}
	})

	t.Run("Multi version shadowed", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		cgs.Add(buildGraph("/o/p/e/s/u/*", 2, cred{id: id2, prev: id1}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 3, cred{id: id3, prev: id2}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 1, cred{id: id1}))

		active, err := cgs.Active()
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if len(active) != 1 {
			t.Error("Wrong active count. wanted: 1 got:", len(active))
		}

		v := active[0].KeyringVersion()
		if v != 3 {
			t.Error("Wrong keyring version. wanted: 3 got:", v)
		}
	})

	t.Run("Multi version skip shadowed", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		cgs.Add(buildGraph("/o/p/e/s/u/*", 2, cred{id: id2}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 3, cred{id: id3, prev: id1}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 1, cred{id: id1}))

		active, err := cgs.Active()
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if len(active) != 2 {
			t.Error("Wrong active count. wanted: 2 got:", len(active))
		}

		if active[0].KeyringVersion() == 1 || active[1].KeyringVersion() == 1 {
			t.Error("Keyring version 1 should not be active")
		}
	})

	t.Run("Multi version mixed shadowed", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		cgs.Add(buildGraph("/o/p/e/s/u/*", 2, cred{id: id3, prev: id1}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 1, cred{id: id1}, cred{id: id2}))

		active, err := cgs.Active()
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if len(active) != 2 {
			t.Error("Wrong active count. wanted: 2 got:", len(active))
		}
	})

	t.Run("Multi version unset", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		cgs.Add(buildGraph("/o/p/e/s/u/*", 2, cred{id: id3, prev: id1}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 1, cred{id: id1}, cred{id: id2, state: &unset}))

		active, err := cgs.Active()
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if len(active) != 1 {
			t.Error("Wrong active count. wanted: 1 got:", len(active))
		}

		v := active[0].KeyringVersion()
		if v != 2 {
			t.Error("Wrong active keyring version. wanted: 2 got:", v)
		}
	})

	t.Run("unset is not active", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		cgs.Add(buildGraph("/o/p/e/s/u/*", 2, cred{id: id3}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 1, cred{id: id1}, cred{id: id2, prev: id1, state: &unset}))

		active, err := cgs.Active()
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if len(active) != 1 {
			t.Error("Wrong active count. wanted: 1 got:", len(active))
		}

		v := active[0].KeyringVersion()
		if v != 2 {
			t.Error("Wrong active keyring version. wanted: 2 got:", v)
		}
	})

	t.Run("unset shadows old versions", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		cgs.Add(buildGraph("/o/p/e/s/u/*", 3, cred{id: id3}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 2, cred{id: id2, prev: id1, state: &unset}))
		cgs.Add(buildGraph("/o/p/e/s/u/*", 1, cred{id: id1}))

		active, err := cgs.Active()
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if len(active) != 1 {
			t.Error("Wrong active count. wanted: 1 got:", len(active))
		}

		v := active[0].KeyringVersion()
		if v != 3 {
			t.Error("Wrong active keyring version. wanted: 2 got:", v)
		}
	})

	t.Run("head version is always active", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		cgs.Add(buildGraph("/o/p/e/s/u/*", 1, cred{id: id1, state: &unset}))

		active, err := cgs.Active()
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if len(active) != 1 {
			t.Error("Wrong active count. wanted: 1 got:", len(active))
		}
	})
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

func TestCredentialGraphSetNeedRotation(t *testing.T) {
	t.Run("no credentials need rotation", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		pe := "/o/p/e/s/u/i"
		name := "cred"
		cgs.Add(buildGraph("/o/p/e/s/u/*", 3, cred{id: id3, pe: &pe, name: &name}))

		out, err := cgs.NeedRotation()
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if len(out) != 0 {
			t.Error("Credentials reported as needing rotation when there shouldn't be")
		}
	})

	t.Run("version in head keyring needs rotation", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		pe := "/o/p/e/s/u/i"
		name := "cred"

		cgs.Add(buildGraphWithRevocation("/o/p/e/s/u/*", 3, cred{id: id3, pe: &pe, name: &name}))

		out, err := cgs.NeedRotation()
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if len(out) != 1 {
			t.Error("Wrong number of credentials needing revision found")
		}
	})

	t.Run("version in old keyring needs rotation", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		pe := "/o/p/e/s/u/i"
		name := "cred"
		othername := "othercred"

		cgs.Add(buildGraph("/o/p/e/s/u/*", 3, cred{id: id3, pe: &pe, name: &name}))
		cgs.Add(buildGraphWithRevocation("/o/p/e/s/u/*", 2, cred{id: id2, pe: &pe, name: &othername}))

		out, err := cgs.NeedRotation()
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if len(out) != 1 {
			t.Fatal("Wrong number of credentials needing revision found")
		}

		if out[0].ID != id2 {
			t.Error("Wrong credential needing revision returned")
		}
	})

	t.Run("already rotated value is not returned", func(t *testing.T) {
		cgs := newCredentialGraphSet()

		pe := "/o/p/e/s/u/i"
		name := "cred"

		cgs.Add(buildGraph("/o/p/e/s/u/*", 3, cred{id: id3, prev: id2, pe: &pe, name: &name}))
		cgs.Add(buildGraphWithRevocation("/o/p/e/s/u/*", 2, cred{id: id2, pe: &pe, name: &name}))

		out, err := cgs.NeedRotation()
		if err != nil {
			t.Fatal("error seen:", err)
		}

		if len(out) != 0 {
			t.Error("Wrong number of credentials needing revision found")
		}
	})
}
