package routes

import (
	"context"
	"fmt"
	"time"

	"github.com/arigatomachine/cli/daemon/base64"
	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/envelope"
	"github.com/arigatomachine/cli/daemon/identity"
	"github.com/arigatomachine/cli/daemon/primitive"
	"github.com/arigatomachine/cli/daemon/registry"
)

// public key types
const (
	encryptionKeyType = "encryption"
	signingKeyType    = "signing"
)

func packagePublicKey(ctx context.Context, engine *crypto.Engine, ownerID,
	orgID *identity.ID, keyType string, public []byte, sigID *identity.ID,
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

	return engine.SignedEnvelope(ctx, &body, sigID, sigKP)
}

func packagePrivateKey(ctx context.Context, engine *crypto.Engine, ownerID,
	orgID *identity.ID, pnonce, private []byte, pubID, sigID *identity.ID,
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

	return engine.SignedEnvelope(ctx, &body, sigID, sigKP)
}

func findEncryptionPublicKey(trees []registry.ClaimTree, orgID *identity.ID,
	userID *identity.ID) (*envelope.Signed, error) {

	// Loop over claimtree looking for the users encryption key
	var encKey *envelope.Signed
	for _, tree := range trees {
		if *tree.Org.ID != *orgID {
			continue
		}

		for _, segment := range tree.PublicKeys {
			key := segment.Key
			keyBody := key.Body.(*primitive.PublicKey)
			if *keyBody.OwnerID != *userID {
				continue
			}

			if keyBody.KeyType != encryptionKeyType {
				continue
			}

			encKey = key
		}
	}

	if encKey == nil {
		err := fmt.Errorf("No encryption pubkey found for: %s", userID.String())
		return nil, err
	}

	return encKey, nil
}

func findEncryptionPublicKeyByID(trees []registry.ClaimTree, orgID *identity.ID,
	ID *identity.ID) (*envelope.Signed, error) {

	// Loop over claimtree looking for the users encryption key
	var encKey *envelope.Signed
	for _, tree := range trees {
		if *tree.Org.ID != *orgID {
			continue
		}

		for _, segment := range tree.PublicKeys {
			key := segment.Key
			keyBody := key.Body.(*primitive.PublicKey)
			if *key.ID != *ID {
				continue
			}

			if keyBody.KeyType != encryptionKeyType {
				continue
			}

			encKey = key
		}
	}

	if encKey == nil {
		err := fmt.Errorf("No encryption pubkey found for: %s", ID.String())
		return nil, err
	}

	return encKey, nil
}
