package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-zoo/bone"

	"github.com/arigatomachine/cli/daemon/config"
	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/registry"
	"github.com/arigatomachine/cli/daemon/session"
)

func NewRouteMux(c *config.Config, s session.Session, db *db.DB,
	t *http.Transport) *bone.Mux {

	engine := crypto.NewEngine(s, db)
	client := registry.NewClient(c.API, s, t)
	mux := bone.New()

	mux.PostFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)

		creds := Login{}
		err := dec.Decode(&creds)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		if creds.Email == "" || creds.Passphrase == "" {
			w.WriteHeader(http.StatusBadRequest)
			enc := json.NewEncoder(w)
			enc.Encode(&Error{Err: "email and passphrase required"})
			return
		}

		salt, err := client.Tokens.PostLogin(creds.Email)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		hmac, err := crypto.DeriveLoginHMAC(creds.Passphrase, salt.Salt,
			salt.Token)
		if err != nil {
			log.Printf("Error generating login token hmac: %s", err)
			encodeResponseErr(w, err)
			return
		}

		auth, err := client.Tokens.PostAuth(salt.Token, hmac)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		self, err := client.Users.GetSelf(auth.Token)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		db.SetMasterKey(*self.Body.Master.Value)
		s.Set(self.ID, creds.Passphrase, auth.Token)

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
		genReq := KeyPairGenerate{}
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

		userID, err := registry.NewIDFromString(s.ID())
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		pubsig, err := packagePublicKey(engine, userID, genReq.OrgID,
			registry.SigningKeyType,
			kp.Signature.Public, nil, &kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		privsig, err := packagePrivateKey(engine, userID, genReq.OrgID,
			kp.Signature.PNonce, kp.Signature.Private, pubsig.ID, pubsig.ID,
			&kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		sigclaim, err := registry.NewEnvelope(engine,
			registry.NewClaim(genReq.OrgID, userID, pubsig.ID, pubsig.ID,
				registry.SignatureClaimType),
			pubsig.ID, &kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		_, _, _, err = client.KeyPairs.Post(pubsig, privsig, sigclaim)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		pubenc, err := packagePublicKey(engine, userID, genReq.OrgID,
			registry.EncryptionKeyType,
			kp.Encryption.Public[:], pubsig.ID, &kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		privenc, err := packagePrivateKey(engine, userID, genReq.OrgID,
			kp.Encryption.PNonce, kp.Encryption.Private, pubenc.ID, pubsig.ID,
			&kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		encclaim, err := registry.NewEnvelope(engine,
			registry.NewClaim(genReq.OrgID, userID, pubenc.ID, pubenc.ID,
				registry.SignatureClaimType),
			pubsig.ID, &kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		_, _, _, err = client.KeyPairs.Post(pubenc, privenc, encclaim)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.GetFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		if !(s.HasToken() && s.HasPassphrase()) {
			w.WriteHeader(http.StatusNotFound)
			err := enc.Encode(&Error{Err: "Not logged in"})
			if err != nil {
				encodeResponseErr(w, err)
			}
			return
		}

		err := enc.Encode(&Status{
			Token:      s.HasToken(),
			Passphrase: s.HasPassphrase(),
		})

		if err != nil {
			encodeResponseErr(w, err)
		}
	})

	mux.GetFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		err := enc.Encode(&Version{Version: c.Version})
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
		enc.Encode(&Error{Err: "Internal server error"})
	}
}
