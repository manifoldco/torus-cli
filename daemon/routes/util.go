package routes

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/registry"
)

func packagePublicKey(engine *crypto.Engine, ownerID, orgID *registry.ID,
	keyType string, public []byte, sigID *registry.ID,
	sigKP *crypto.SignatureKeyPair) (*registry.Envelope, error) {

	alg := "curve25519"
	if keyType == registry.SigningKeyType {
		alg = "eddsa"
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

	return envelope(engine, &body, sigID, sigKP)
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
			Algorithm: "triplesec-v3",
			Value:     &kv,
		},
	}

	return envelope(engine, &body, sigID, sigKP)
}

func envelope(engine *crypto.Engine, body registry.AgObject, sigID *registry.ID,
	sigKP *crypto.SignatureKeyPair) (*registry.Envelope, error) {

	b, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}

	s, err := engine.Sign(*sigKP, append([]byte(strconv.Itoa(1)), b...))
	if err != nil {
		return nil, err
	}

	sv := registry.Base64Value(s)
	sig := registry.Signature{
		PublicKeyID: sigID,
		Algorithm:   "eddsa",
		Value:       &sv,
	}

	id, err := registry.NewID(1, body, &sig)
	if err != nil {
		return nil, err
	}

	return &registry.Envelope{
		ID:        &id,
		Version:   1,
		Body:      body,
		Signature: sig,
	}, nil
}
