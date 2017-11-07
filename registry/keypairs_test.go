package registry

import (
	"testing"

	gm "github.com/onsi/gomega"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

func createCKP(orgID *identity.ID, t primitive.KeyType, revoked bool) ClaimedKeyPair {
	pubk := &envelope.PublicKey{
		Version: 1,
		Body: &primitive.PublicKey{
			OrgID:   orgID,
			KeyType: t,
		},
	}

	pubkID, err := identity.NewImmutable(pubk.Body, "haha")
	if err != nil {
		panic(err)
	}
	pubk.ID = &pubkID

	privk := &envelope.PrivateKey{
		Version: 1,
		Body: &primitive.PrivateKey{
			OrgID:       orgID,
			PublicKeyID: pubk.ID,
		},
	}
	privkID, err := identity.NewImmutable(privk.Body, "sdfs")
	if err != nil {
		panic(err)
	}
	privk.ID = &privkID

	ckp := ClaimedKeyPair{
		PublicKeySegment: apitypes.PublicKeySegment{
			PublicKey: pubk,
			Claims:    []envelope.Claim{},
		},
		PrivateKey: privk,
	}

	if revoked {
		ckp.Claims = append(ckp.Claims, envelope.Claim{
			Version: 1,
			Body: &primitive.Claim{
				OrgID:       orgID,
				PublicKeyID: pubk.ID,
				ClaimType:   primitive.RevocationClaimType,
			},
		})
	}

	return ckp
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
