package routes

import (
	"time"

	"github.com/arigatomachine/cli/daemon/base64"
	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/envelope"
	"github.com/arigatomachine/cli/daemon/identity"
	"github.com/arigatomachine/cli/daemon/primitive"
)

func packagePublicKey(engine *crypto.Engine, ownerID, orgID *identity.ID,
	keyType string, public []byte, sigID *identity.ID,
	sigKP *crypto.SignatureKeyPair) (*envelope.Signed, error) {

	alg := crypto.Curve25519
	if keyType == SigningKeyType {
		alg = crypto.EdDSA
	}

	now := time.Now().UTC()

	kv := base64.Value(public)
	body := primitive.PublicKey{
		OrgID:     orgID,
		OwnerID:   ownerID,
		KeyType:   keyType,
		Algorithm: alg,

		Key: primitive.PublicKeyValue{
			Value: &kv,
		},

		Created: now,
		Expires: now.Add(time.Hour * 8760), // one year
	}

	return engine.SignedEnvelope(&body, sigID, sigKP)
}

func packagePrivateKey(engine *crypto.Engine, ownerID, orgID *identity.ID,
	pnonce, private []byte, pubID, sigID *identity.ID,
	sigKP *crypto.SignatureKeyPair) (*envelope.Signed, error) {

	kv := base64.Value(private)
	pv := base64.Value(pnonce)
	body := primitive.PrivateKey{
		OrgID:       orgID,
		OwnerID:     ownerID,
		PNonce:      &pv,
		PublicKeyID: pubID,

		Key: primitive.PrivateKeyValue{
			Algorithm: crypto.Triplesec,
			Value:     &kv,
		},
	}

	return engine.SignedEnvelope(&body, sigID, sigKP)
}
