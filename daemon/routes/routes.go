package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-zoo/bone"

	"github.com/arigatomachine/cli/daemon/config"
	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/envelope"
	"github.com/arigatomachine/cli/daemon/primitive"
	"github.com/arigatomachine/cli/daemon/registry"
	"github.com/arigatomachine/cli/daemon/session"
)

// NewRouteMux returns a *bone.Mux responsible for handling the cli to daemon
// http api.
func NewRouteMux(c *config.Config, s session.Session, db *db.DB,
	t *http.Transport) *bone.Mux {

	engine := crypto.NewEngine(s, db)
	client := registry.NewClient(c.API, s, t)
	mux := bone.New()

	mux.PostFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)

		creds := login{}
		err := dec.Decode(&creds)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		if creds.Email == "" || creds.Passphrase == "" {
			w.WriteHeader(http.StatusBadRequest)
			enc := json.NewEncoder(w)
			enc.Encode(&errorMsg{
				Type:  badRequestError,
				Error: "email and passphrase required",
			})
			return
		}

		salt, loginToken, err := client.Tokens.PostLogin(creds.Email)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		hmac, err := crypto.DeriveLoginHMAC(creds.Passphrase, salt, loginToken)
		if err != nil {
			log.Printf("Error generating login token hmac: %s", err)
			encodeResponseErr(w, err)
			return
		}

		authToken, err := client.Tokens.PostAuth(loginToken, hmac)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		self, err := client.Users.GetSelf(authToken)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		db.Set(self)
		s.Set(self.ID, creds.Passphrase, authToken)

		w.WriteHeader(http.StatusNoContent)
	})

	mux.PostFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		tok := s.Token()

		if tok == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		err := client.Tokens.Delete(tok)
		switch err := err.(type) {
		case *registry.Error:
			switch {
			case err.StatusCode >= 500:
				// On a 5XX response, we don't know for sure that the server
				// has successfully removed the auth token. Keep the copy in
				// the daemon, so the user may try again.
				encodeResponseErr(w, err)
			case err.StatusCode >= 400:
				// A 4XX error indicates either the token isn't found, or we're
				// not allowed to remove it (or the server is a teapot).
				//
				// In any case, the daemon has gotten out of sync with the
				// server. Remove our local copy of the auth token.
				log.Printf("Got 4XX removing auth token. Treating as success")
				s.Logout()
				w.WriteHeader(http.StatusNoContent)
			}
		case nil:
			s.Logout()
			w.WriteHeader(http.StatusNoContent)
		default:
			encodeResponseErr(w, err)
		}
	})

	mux.PostFunc("/keypairs/generate", func(w http.ResponseWriter,
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
	})

	mux.PostFunc("/credentials", func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		cred := &envelope.Unsigned{}

		err := dec.Decode(&cred)
		if err != nil {
			log.Printf("error decoding credential: %s", err)
			encodeResponseErr(w, err)
			return
		}

		// Get matching credentials
		credBody := cred.Body.(*primitive.Credential)
		creds, err := client.Credentials.List(credBody.Name, "", credBody.PathExp)
		if err != nil {
			log.Printf("error retrieving previous cred: %s", err)
			encodeResponseErr(w, err)
			return
		}

		if len(creds) == 0 {
			credBody.Previous = nil
			credBody.CredentialVersion = 1
		} else {
			previousCred := creds[len(creds)-1]
			previousCredBody := previousCred.Body.(*primitive.Credential)

			credBody.Previous = previousCred.ID
			credBody.CredentialVersion = previousCredBody.CredentialVersion + 1
		}

		cred, err = client.Credentials.Create(cred)
		if err != nil {
			log.Printf("error creating credential: %s", err)
			encodeResponseErr(w, err)
			return
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(cred)
		if err != nil {
			log.Printf("error encoding credential create resp: %s", err)
			encodeResponseErr(w, err)
			return
		}
	})

	mux.GetFunc("/credentials", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		trees, err := client.CredentialTree.List(
			q.Get("Name"), q.Get("path"), q.Get("pathexp"), s.ID())
		if err != nil {
			log.Printf("error retrieving credential trees: %s", err)
			encodeResponseErr(w, err)
			return
		}

		// Loop over the trees and unpack the credentials; later on we will
		// actually do real work and decrypt each of these credentials but for
		// now we just need ot return a list of them!
		creds := []envelope.Unsigned{}
		for _, tree := range trees {
			creds = append(creds, tree.Credentials...)
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(creds)
		if err != nil {
			log.Printf("error encoding credentials: %s", err)
			encodeResponseErr(w, err)
			return
		}
	})

	mux.GetFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		if !(s.HasToken() && s.HasPassphrase()) {
			w.WriteHeader(http.StatusNotFound)
			err := enc.Encode(&errorMsg{
				Type:  unauthorizedError,
				Error: "Not logged in",
			})
			if err != nil {
				encodeResponseErr(w, err)
			}
			return
		}

		err := enc.Encode(&status{
			Token:      s.HasToken(),
			Passphrase: s.HasPassphrase(),
		})

		if err != nil {
			encodeResponseErr(w, err)
		}
	})

	mux.GetFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		err := enc.Encode(&version{Version: c.Version})
		if err != nil {
			encodeResponseErr(w, err)
		}
	})

	return mux
}

// if encoding has errored, our struct is either bad, or our writer
// is broken. Try writing an error back to the client, but ignore any
// problems (ie the writer is broken).
func encodeResponseErr(w http.ResponseWriter, err error) {
	enc := json.NewEncoder(w)

	rErr, ok := err.(*registry.Error)
	if ok {
		w.WriteHeader(rErr.StatusCode)
		enc.Encode(rErr)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(&errorMsg{
			Type:  internalServerError,
			Error: "Internal server error",
		})
	}
}
