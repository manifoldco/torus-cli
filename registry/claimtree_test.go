package registry

import (
	"testing"

	gm "github.com/onsi/gomega"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

func TestClaimTree(t *testing.T) {
	orgID, err := identity.NewMutable(&primitive.Org{})
	if err != nil {
		t.Fatal("Could not generate org id", err)
	}

	t.Run("can find pubkey by id", func(t *testing.T) {
		gm.RegisterTestingT(t)

		pks := createPKS(&orgID, primitive.EncryptionKeyType, false)
		ct := ClaimTree{PublicKeys: []apitypes.PublicKeySegment{pks}}

		found, err := ct.Find(pks.PublicKey.ID, false)
		gm.Expect(err).To(gm.BeNil())
		gm.Expect(*found).To(gm.Equal(pks))
	})

	t.Run("errors if pubkey is revoked", func(t *testing.T) {
		gm.RegisterTestingT(t)

		pks := createPKS(&orgID, primitive.EncryptionKeyType, true)
		ct := ClaimTree{PublicKeys: []apitypes.PublicKeySegment{pks}}

		found, err := ct.Find(pks.PublicKey.ID, true)
		gm.Expect(err).To(gm.Equal(ErrKeyNotFound))
		gm.Expect(found).To(gm.BeNil())
	})

	t.Run("can find an active key for user", func(t *testing.T) {
		gm.RegisterTestingT(t)

		pksA := createPKS(&orgID, primitive.EncryptionKeyType, false)
		pksB := createPKS(&orgID, primitive.SigningKeyType, false)

		ct := ClaimTree{PublicKeys: []apitypes.PublicKeySegment{pksA, pksB}}

		found, err := ct.FindActive(pksA.PublicKey.Body.OwnerID, primitive.EncryptionKeyType)
		gm.Expect(err).To(gm.BeNil())
		gm.Expect(*found).To(gm.Equal(pksA))

		found, err = ct.FindActive(pksB.PublicKey.Body.OwnerID, primitive.SigningKeyType)
		gm.Expect(err).To(gm.BeNil())
		gm.Expect(*found).To(gm.Equal(pksB))
	})

	t.Run("returns error if an active key cannot be found", func(t *testing.T) {
		gm.RegisterTestingT(t)

		pksA := createPKS(&orgID, primitive.EncryptionKeyType, true)

		ct := ClaimTree{PublicKeys: []apitypes.PublicKeySegment{pksA}}

		found, err := ct.FindActive(pksA.PublicKey.Body.OwnerID, primitive.EncryptionKeyType)
		gm.Expect(err).To(gm.Equal(ErrMissingKeyForOwner))
		gm.Expect(found).To(gm.BeNil())
	})

	t.Run("returns error if segment exists but isnt the right type", func(t *testing.T) {
		gm.RegisterTestingT(t)

		pksA := createPKS(&orgID, primitive.SigningKeyType, false)
		ct := ClaimTree{PublicKeys: []apitypes.PublicKeySegment{pksA}}

		found, err := ct.FindActive(pksA.PublicKey.Body.OwnerID, primitive.EncryptionKeyType)
		gm.Expect(err).To(gm.Equal(ErrMissingKeyForOwner))
		gm.Expect(found).To(gm.BeNil())
	})

	t.Run("errors if pubkey can't be found", func(t *testing.T) {
		gm.RegisterTestingT(t)

		pubID, err := identity.NewImmutable(&primitive.PublicKey{}, "hi")
		gm.Expect(err).To(gm.BeNil())

		ct := ClaimTree{PublicKeys: []apitypes.PublicKeySegment{}}

		_, err = ct.Find(&pubID, false)
		gm.Expect(err).To(gm.Equal(ErrKeyNotFound))
	})
}
