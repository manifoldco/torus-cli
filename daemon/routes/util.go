package routes

import (
	"time"

	"github.com/arigatomachine/cli/daemon/base64"
	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/envelope"
	"github.com/arigatomachine/cli/daemon/identity"
	"github.com/arigatomachine/cli/daemon/primitive"
)

// public key types
const (
	encryptionKeyType = "encryption"
	signingKeyType    = "signing"
)

func packagePublicKey(engine *crypto.Engine, ownerID, orgID *identity.ID,
	keyType string, public []byte, sigID *identity.ID,
	sigKP *crypto.SignatureKeyPair) (*envelope.Signed, error) {

	alg := crypto.Curve25519
	if keyType == signingKeyType {
		alg = crypto.EdDSA
	}

	now := time.Now().UTC()

	body := primitive.PublicKey{
		OrgID:     orgID,
		OwnerID:   ownerID,
		KeyType:   keyType,
		Algorithm: alg,

		Key: primitive.PublicKeyValue{
			Value: base64.NewValue(public),
		},

		Created: now,
		Expires: now.Add(time.Hour * 8760), // one year
	}

	return engine.SignedEnvelope(&body, sigID, sigKP)
}

func packagePrivateKey(engine *crypto.Engine, ownerID, orgID *identity.ID,
	pnonce, private []byte, pubID, sigID *identity.ID,
	sigKP *crypto.SignatureKeyPair) (*envelope.Signed, error) {

	body := primitive.PrivateKey{
		OrgID:       orgID,
		OwnerID:     ownerID,
		PNonce:      base64.NewValue(pnonce),
		PublicKeyID: pubID,

		Key: primitive.PrivateKeyValue{
			Algorithm: crypto.Triplesec,
			Value:     base64.NewValue(private),
		},
	}

	return engine.SignedEnvelope(&body, sigID, sigKP)
}
