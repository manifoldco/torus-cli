package routes

import (
	"time"

	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/registry"
)

func packagePublicKey(engine *crypto.Engine, ownerID, orgID *registry.ID,
	keyType string, public []byte, sigID *registry.ID,
	sigKP *crypto.SignatureKeyPair) (*registry.Envelope, error) {

	alg := crypto.Curve25519
	if keyType == registry.SigningKeyType {
		alg = crypto.EdDSA
	}

	now := time.Now().UTC()

	kv := registry.Base64Value(public)
	body := registry.PublicKey{
		OrgID:     orgID,
		OwnerID:   ownerID,
		KeyType:   keyType,
		Algorithm: alg,

		Key: registry.PublicKeyValue{
			Value: &kv,
		},

		Created: now,
		Expires: now.Add(time.Hour * 8760), // one year
	}

	return registry.NewEnvelope(engine, &body, sigID, sigKP)
}

func packagePrivateKey(engine *crypto.Engine, ownerID, orgID *registry.ID,
	pnonce, private []byte, pubID, sigID *registry.ID,
	sigKP *crypto.SignatureKeyPair) (*registry.Envelope, error) {

	kv := registry.Base64Value(private)
	pv := registry.Base64Value(pnonce)
	body := registry.PrivateKey{
		OrgID:       orgID,
		OwnerID:     ownerID,
		PNonce:      &pv,
		PublicKeyID: pubID,

		Key: registry.PrivateKeyValue{
			Algorithm: crypto.Triplesec,
			Value:     &kv,
		},
	}

	return registry.NewEnvelope(engine, &body, sigID, sigKP)
}
