package logic

import (
	"testing"

	gm "github.com/onsi/gomega"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
	"github.com/manifoldco/torus-cli/registry"
)

func addMembership(cg registry.CredentialGraph, authID, encKeyID *identity.ID, revoked bool) {
	krm := &envelope.KeyringMember{
		Version: 2,
		Body: &primitive.KeyringMember{
			EncryptingKeyID: encKeyID,
			OwnerID:         authID,
		},
	}
	krmID, err := identity.NewImmutable(krm.Body, "haha")
	if err != nil {
		panic(err)
	}
	krm.ID = &krmID

	mek := &envelope.MEKShare{
		Body: &primitive.MEKShare{
			KeyringMemberID: krm.ID,
			OwnerID:         authID,
		},
	}

	mekID, err := identity.NewImmutable(mek.Body, "yoyo")
	if err != nil {
		panic(err)
	}
	mek.ID = &mekID

	cgv2 := cg.(*registry.CredentialGraphV2)
	cgv2.Members = append(cgv2.Members, registry.KeyringMember{
		Member:   krm,
		MEKShare: mek,
	})

	if revoked {
		cgv2.Claims = append(cgv2.Claims, envelope.KeyringMemberClaim{
			Version: 1,
			Body: &primitive.KeyringMemberClaim{
				KeyringMemberID: krm.ID,
				ClaimType:       primitive.RevocationClaimType,
			},
		})
	}
}

func TestCredentialGraphKeyIndex(t *testing.T) {
	t.Run("can index graph", func(t *testing.T) {
		gm.RegisterTestingT(t)

		cgA := buildGraph("/o/p/e/s/u/i", 1)
		cgB := buildGraph("/o/p/e/s/u2/i", 1)

		userA, err := identity.NewMutable(&primitive.User{})
		gm.Expect(err).To(gm.BeNil())

		keyA, err := identity.NewImmutable(&primitive.PublicKey{}, "hi")
		gm.Expect(err).To(gm.BeNil())

		addMembership(cgA, &userA, &keyA, false)
		addMembership(cgB, &userA, &keyA, false)

		userB, err := identity.NewMutable(&primitive.User{})
		gm.Expect(err).To(gm.BeNil())

		keyB, err := identity.NewImmutable(&primitive.PublicKey{}, "yo")
		gm.Expect(err).To(gm.BeNil())

		addMembership(cgA, &userB, &keyB, false)
		addMembership(cgB, &userB, &keyB, false)

		idx := newCredentialGraphKeyIndex(userA)
		idx.Add(cgA, cgB)

		keyMap := idx.GetIndex()
		gm.Expect(len(keyMap)).To(gm.Equal(1))

		_, ok := keyMap[keyB]
		gm.Expect(ok).To(gm.BeFalse(), "User B should not be in the idx")

		graphs, ok := keyMap[keyA]
		gm.Expect(ok).To(gm.BeTrue())
		gm.Expect(len(graphs)).To(gm.Equal(2))
	})

	t.Run("returns error if auth id can't be found", func(t *testing.T) {
		gm.RegisterTestingT(t)

		cgA := buildGraph("/o/p/e/s/u/i", 1)

		userA, err := identity.NewMutable(&primitive.User{})
		gm.Expect(err).To(gm.BeNil())

		idx := newCredentialGraphKeyIndex(userA)
		err = idx.Add(cgA)

		gm.Expect(err).ToNot(gm.BeNil())
		gm.Expect(err).To(gm.Equal(registry.ErrMemberNotFound))
	})

	t.Run("handles case where user has multiple keys encoded in ring", func(t *testing.T) {
		gm.RegisterTestingT(t)

		cgA := buildGraph("/o/p/e/s/u/i", 1)

		userA, err := identity.NewMutable(&primitive.User{})
		gm.Expect(err).To(gm.BeNil())

		keyA, err := identity.NewImmutable(&primitive.PublicKey{}, "hi")
		gm.Expect(err).To(gm.BeNil())

		keyB, err := identity.NewImmutable(&primitive.PublicKey{}, "hey")
		gm.Expect(err).To(gm.BeNil())

		addMembership(cgA, &userA, &keyA, false)
		addMembership(cgA, &userA, &keyB, false)

		idx := newCredentialGraphKeyIndex(userA)
		err = idx.Add(cgA)
		gm.Expect(err).To(gm.BeNil())

		gm.Expect(len(idx.GetIndex())).To(gm.Equal(1), "Expected stable selection of key")
	})

	t.Run("handles case where users key is revoked", func(t *testing.T) {
		gm.RegisterTestingT(t)

		cgA := buildGraph("/o/p/e/s/u/i", 2)

		userA, err := identity.NewMutable(&primitive.User{})
		gm.Expect(err).To(gm.BeNil())

		keyA, err := identity.NewImmutable(&primitive.PublicKey{}, "hi")
		gm.Expect(err).To(gm.BeNil())

		addMembership(cgA, &userA, &keyA, true)

		idx := newCredentialGraphKeyIndex(userA)
		err = idx.Add(cgA)

		gm.Expect(err).ToNot(gm.BeNil())
		gm.Expect(err).To(gm.Equal(registry.ErrMemberNotFound))
	})

	t.Run("handles case with child graphs", func(t *testing.T) {
		cgA := buildGraph("/o/p/e/s/u/i", 1)
		cgB := buildGraph("/o/p/e/s/u/i", 2)

		userA, err := identity.NewMutable(&primitive.User{})
		gm.Expect(err).To(gm.BeNil())

		keyA, err := identity.NewImmutable(&primitive.PublicKey{}, "hi")
		gm.Expect(err).To(gm.BeNil())

		addMembership(cgA, &userA, &keyA, false)
		addMembership(cgB, &userA, &keyA, false)

		idx := newCredentialGraphKeyIndex(userA)
		err = idx.Add(cgA, cgB)
		gm.Expect(err).To(gm.BeNil())

		keyMap := idx.GetIndex()
		gm.Expect(len(keyMap)).To(gm.Equal(1))
		gm.Expect(len(keyMap[keyA])).To(gm.Equal(2))
	})
}
