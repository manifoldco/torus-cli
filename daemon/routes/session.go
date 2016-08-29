package routes

// This file contains routes related to the user's session

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arigatomachine/cli/apitypes"

	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/registry"
	"github.com/arigatomachine/cli/daemon/session"
)

func loginRoute(client *registry.Client, s session.Session,
	db *db.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dec := json.NewDecoder(r.Body)

		creds := apitypes.Login{}
		err := dec.Decode(&creds)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		if creds.Email == "" || creds.Passphrase == "" {
			w.WriteHeader(http.StatusBadRequest)
			enc := json.NewEncoder(w)
			enc.Encode(&errorMsg{
				Type:  apitypes.BadRequestError,
				Error: []string{"email and passphrase required"},
			})
			return
		}

		salt, loginToken, err := client.Tokens.PostLogin(ctx, creds.Email)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		hmac, err := crypto.DeriveLoginHMAC(ctx, creds.Passphrase, salt, loginToken)
		if err != nil {
			log.Printf("Error generating login token hmac: %s", err)
			encodeResponseErr(w, err)
			return
		}

		authToken, err := client.Tokens.PostAuth(ctx, loginToken, hmac)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		self, err := client.Users.GetSelf(ctx, authToken)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		db.Set(self)
		s.Set(self.ID, creds.Passphrase, authToken)

		w.WriteHeader(http.StatusNoContent)
	}
}

func logoutRoute(client *registry.Client, s session.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := s.Token()

		if tok == "" {
			encodeResponseErr(w, notFoundError)
			return
		}

		err := client.Tokens.Delete(r.Context(), tok)
		switch err := err.(type) {
		case *apitypes.Error:
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
	}
}

func sessionRoute(s session.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		if !(s.HasToken() && s.HasPassphrase()) {
			w.WriteHeader(http.StatusNotFound)
			err := enc.Encode(&errorMsg{
				Type:  apitypes.UnauthorizedError,
				Error: []string{"Not logged in"},
			})
			if err != nil {
				encodeResponseErr(w, err)
			}
			return
		}

		err := enc.Encode(&apitypes.SessionStatus{
			Token:      s.HasToken(),
			Passphrase: s.HasPassphrase(),
		})

		if err != nil {
			encodeResponseErr(w, err)
		}
	}
}
