package routes

// This file contains routes related to keypairs

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/envelope"
	"github.com/arigatomachine/cli/daemon/primitive"
	"github.com/arigatomachine/cli/daemon/registry"
	"github.com/arigatomachine/cli/daemon/session"
)

func keypairsGenerateRoute(client *registry.Client, s session.Session,
	db *db.DB, engine *crypto.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter,
		r *http.Request) {

		dec := json.NewDecoder(r.Body)
		genReq := keyPairGenerate{}
		err := dec.Decode(&genReq)
		if err != nil || genReq.OrgID == nil {
			encodeResponseErr(w, err)
			return
		}

		kp, err := engine.GenerateKeyPairs()
		if err != nil {
			log.Printf("Error generating keypairs: %s", err)
			encodeResponseErr(w, err)
			return
		}

		pubsig, err := packagePublicKey(engine, s.ID(), genReq.OrgID,
			signingKeyType, kp.Signature.Public, nil, &kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		privsig, err := packagePrivateKey(engine, s.ID(), genReq.OrgID,
			kp.Signature.PNonce, kp.Signature.Private, pubsig.ID, pubsig.ID,
			&kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		sigclaim, err := engine.SignedEnvelope(
			primitive.NewClaim(genReq.OrgID, s.ID(), pubsig.ID, pubsig.ID,
				primitive.SignatureClaimType),
			pubsig.ID, &kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		pubsig, privsig, claims, err := client.KeyPairs.Post(pubsig, privsig,
			sigclaim)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		objs := make([]envelope.Envelope, len(claims)+2)
		objs[0] = pubsig
		objs[1] = privsig
		for i, claim := range claims {
			objs[i+2] = &claim
		}
		err = db.Set(objs...)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		pubenc, err := packagePublicKey(engine, s.ID(), genReq.OrgID,
			encryptionKeyType, kp.Encryption.Public[:], pubsig.ID,
			&kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		privenc, err := packagePrivateKey(engine, s.ID(), genReq.OrgID,
			kp.Encryption.PNonce, kp.Encryption.Private, pubenc.ID, pubsig.ID,
			&kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		encclaim, err := engine.SignedEnvelope(
			primitive.NewClaim(genReq.OrgID, s.ID(), pubenc.ID, pubenc.ID,
				primitive.SignatureClaimType),
			pubsig.ID, &kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		pubenc, privenc, claims, err = client.KeyPairs.Post(pubenc, privenc,
			encclaim)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		objs = make([]envelope.Envelope, len(claims)+2)
		objs[0] = pubenc
		objs[1] = privenc
		for i, claim := range claims {
			objs[i+2] = &claim
		}
		err = db.Set(objs...)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
