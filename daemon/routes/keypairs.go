package routes

// This file contains routes related to keypairs

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/envelope"
	"github.com/arigatomachine/cli/daemon/observer"
	"github.com/arigatomachine/cli/daemon/primitive"
	"github.com/arigatomachine/cli/daemon/registry"
	"github.com/arigatomachine/cli/daemon/session"
)

func keypairsGenerateRoute(client *registry.Client, s session.Session,
	db *db.DB, engine *crypto.Engine, o *observer.Observer) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var step uint = 1
		ctx := r.Context()

		dec := json.NewDecoder(r.Body)
		genReq := keyPairGenerate{}
		err := dec.Decode(&genReq)
		if err != nil || genReq.OrgID == nil {
			encodeResponseErr(w, err)
			return
		}

		n, err := o.Notifier(ctx, 5)
		if err != nil {
			log.Printf("Error creating Notifier: %s", err)
			encodeResponseErr(w, err)
			return
		}

		kp, err := engine.GenerateKeyPairs(ctx)
		if err != nil {
			log.Printf("Error generating keypairs: %s", err)
			encodeResponseErr(w, err)
			return
		}

		n.Notify(observer.Progress, "Keypairs generated", true)

		pubsig, err := packagePublicKey(ctx, engine, s.ID(), genReq.OrgID,
			signingKeyType, kp.Signature.Public, nil, &kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		privsig, err := packagePrivateKey(ctx, engine, s.ID(), genReq.OrgID,
			kp.Signature.PNonce, kp.Signature.Private, pubsig.ID, pubsig.ID,
			&kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		sigclaim, err := engine.SignedEnvelope(
			ctx, primitive.NewClaim(genReq.OrgID, s.ID(), pubsig.ID, pubsig.ID,
				primitive.SignatureClaimType),
			pubsig.ID, &kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		n.Notify(observer.Progress, "Signing keys signed", true)

		pubsig, privsig, claims, err := client.KeyPairs.Post(ctx, pubsig,
			privsig, sigclaim)
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

		n.Notify(observer.Progress, "Signing keys uploaded", true)

		pubenc, err := packagePublicKey(ctx, engine, s.ID(), genReq.OrgID,
			encryptionKeyType, kp.Encryption.Public[:], pubsig.ID,
			&kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		privenc, err := packagePrivateKey(ctx, engine, s.ID(), genReq.OrgID,
			kp.Encryption.PNonce, kp.Encryption.Private, pubenc.ID, pubsig.ID,
			&kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		encclaim, err := engine.SignedEnvelope(
			ctx, primitive.NewClaim(genReq.OrgID, s.ID(), pubenc.ID, pubenc.ID,
				primitive.SignatureClaimType),
			pubsig.ID, &kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		n.Notify(observer.Progress, "Encryption keys signed", true)

		pubenc, privenc, claims, err = client.KeyPairs.Post(ctx, pubenc,
			privenc, encclaim)
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

		n.Notify(observer.Progress, "Encryption keys uploaded", true)
		step++

		w.WriteHeader(http.StatusNoContent)
	}
}
