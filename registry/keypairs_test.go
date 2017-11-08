package registry

import (
	"testing"

	gm "github.com/onsi/gomega"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

func createPKS(orgID *identity.ID, t primitive.KeyType, revoked bool) apitypes.PublicKeySegment {
	ownerID, err := identity.NewMutable(&primitive.User{})
	if err != nil {
		panic(err)
	}

	pubk := &envelope.PublicKey{
		Version: 1,
		Body: &primitive.PublicKey{
			OrgID:   orgID,
			KeyType: t,
			OwnerID: &ownerID,
		},
	}

	pubkID, err := identity.NewImmutable(pubk.Body, "haha")
	if err != nil {
		panic(err)
	}
	pubk.ID = &pubkID

	pks := apitypes.PublicKeySegment{
		PublicKey: pubk,
		Claims:    []envelope.Claim{},
	}

	if revoked {
		pks.Claims = append(pks.Claims, envelope.Claim{
			Version: 1,
			Body: &primitive.Claim{
				OrgID:       orgID,
				PublicKeyID: pubk.ID,
				ClaimType:   primitive.RevocationClaimType,
			},
		})
	}

	return pks
}

func createCKP(orgID *identity.ID, t primitive.KeyType, revoked bool) ClaimedKeyPair {
	pks := createPKS(orgID, t, revoked)
	privk := &envelope.PrivateKey{
		Version: 1,
		Body: &primitive.PrivateKey{
			OrgID:       orgID,
			OwnerID:     pks.PublicKey.Body.OwnerID,
			PublicKeyID: pks.PublicKey.ID,
		},
	}
	privkID, err := identity.NewImmutable(privk.Body, "sdfs")
	if err != nil {
		panic(err)
	}
	privk.ID = &privkID

	return ClaimedKeyPair{
		PublicKeySegment: pks,
		PrivateKey:       privk,
	}
}

func TestKeypairs(t *testing.T) {
	t.Run("can add keypairs and subsequently find them", func(t *testing.T) {
		gm.RegisterTestingT(t)

		orgID, err := identity.NewMutable(&primitive.Org{})
		gm.Expect(err).To(gm.BeNil())

		kp := NewKeypairs()
		ckp := createCKP(&orgID, primitive.EncryptionKeyType, false)

		err = kp.Add(ckp)
		gm.Expect(err).To(gm.BeNil())

		found, err := kp.Get(ckp.PublicKey.ID)
		gm.Expect(err).To(gm.BeNil())
		gm.Expect(found.PublicKey.ID).To(gm.Equal(ckp.PublicKey.ID))

		all := kp.All()
		gm.Expect(len(all)).To(gm.Equal(1))
		gm.Expect(all[0].PublicKey.ID).To(gm.Equal(ckp.PublicKey.ID))
	})

	t.Run("can select keypairs", func(t *testing.T) {
		gm.RegisterTestingT(t)

		orgID, err := identity.NewMutable(&primitive.Org{})
		gm.Expect(err).To(gm.BeNil())

		kp := NewKeypairs()

		ckpA := createCKP(&orgID, primitive.EncryptionKeyType, false)
		ckpB := createCKP(&orgID, primitive.SigningKeyType, false)
		ckpC := createCKP(&orgID, primitive.EncryptionKeyType, true)

		err = kp.Add(ckpA, ckpB, ckpC)
		gm.Expect(err).To(gm.BeNil())

		found, err := kp.Select(&orgID, primitive.EncryptionKeyType)
		gm.Expect(err).To(gm.BeNil())
		gm.Expect(found.PublicKey.ID).To(gm.Equal(ckpA.PublicKey.ID))

		found, err = kp.Select(&orgID, primitive.SigningKeyType)
		gm.Expect(err).To(gm.BeNil())
		gm.Expect(found.PublicKey.ID).To(gm.Equal(ckpB.PublicKey.ID))
	})

	t.Run("will not select revoked keypairs", func(t *testing.T) {
		gm.RegisterTestingT(t)

		orgID, err := identity.NewMutable(&primitive.Org{})
		gm.Expect(err).To(gm.BeNil())

		kp := NewKeypairs()

		ckpA := createCKP(&orgID, primitive.EncryptionKeyType, true)
		kp.Add(ckpA)

		_, err = kp.Select(&orgID, primitive.EncryptionKeyType)
		gm.Expect(err).To(gm.Equal(ErrMissingValidKeypair))
	})

	t.Run("will error if org does not have keys", func(t *testing.T) {
		gm.RegisterTestingT(t)

		orgID, err := identity.NewMutable(&primitive.Org{})
		gm.Expect(err).To(gm.BeNil())

		kp := NewKeypairs()
		_, err = kp.Select(&orgID, primitive.SigningKeyType)
		gm.Expect(err).To(gm.Equal(ErrMissingKeysForOrg))
	})

	t.Run("will error if key does not exist", func(t *testing.T) {
		gm.RegisterTestingT(t)

		pubkID, err := identity.NewImmutable(&primitive.PublicKey{}, "has")
		gm.Expect(err).To(gm.BeNil())

		kp := NewKeypairs()
		_, err = kp.Get(&pubkID)
		gm.Expect(err).To(gm.Equal(ErrPublicKeyNotFound))
	})
}
