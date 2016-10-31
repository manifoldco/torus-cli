package routes

// This file contains routes related to the user's session

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/identity"

	"github.com/manifoldco/torus-cli/daemon/crypto"
	"github.com/manifoldco/torus-cli/daemon/db"
	"github.com/manifoldco/torus-cli/daemon/registry"
	"github.com/manifoldco/torus-cli/daemon/session"
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

		err = attemptLogin(ctx, client, s, db, creds)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func attemptLogin(ctx context.Context, client *registry.Client, s session.Session, db *db.DB, creds apitypes.Login) error {
	salt, loginToken, err := client.Tokens.PostLogin(ctx, creds.Email)
	if err != nil {
		return err
	}

	hmac, err := crypto.DeriveLoginHMAC(ctx, creds.Passphrase, salt, loginToken)
	if err != nil {
		return err
	}

	authToken, err := client.Tokens.PostAuth(ctx, loginToken, hmac)
	if err != nil {
		return err
	}

	self, err := client.Users.GetSelf(ctx, authToken)
	if err != nil {
		return err
	}

	db.Set(self)
	return s.Set("user", self, self, creds.Passphrase, authToken)
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
				logoutErr := s.Logout()
				if logoutErr != nil {
					log.Printf("Error while attempting to destroy session: %s", logoutErr)
					encodeResponseErr(w, logoutErr)
					break
				}

				w.WriteHeader(http.StatusNoContent)
			}
		case nil:
			logoutErr := s.Logout()
			if logoutErr != nil {
				log.Printf("Error while attempting to destroy session: %s", logoutErr)
				encodeResponseErr(w, logoutErr)
				break
			}

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

func signupRoute(client *registry.Client, s session.Session, db *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dec := json.NewDecoder(r.Body)

		signup := apitypes.Signup{}
		err := dec.Decode(&signup)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		passwordObj, masterObj, err := crypto.EncryptPasswordObject(ctx, signup.Passphrase)
		if err != nil {
			log.Printf("Error generating password object: %s", err)
			encodeResponseErr(w, err)
			return
		}

		userBody := registry.SignupBody{
			Username: signup.Username,
			Name:     signup.Name,
			Email:    signup.Email,
			Password: passwordObj,
			Master:   masterObj,
		}

		ID, err := identity.NewMutable(&userBody)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		userObj := registry.Signup{
			ID:      ID.String(),
			Version: 1,
			Body:    &userBody,
		}

		user, err := client.Users.Create(ctx, userObj, signup)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		w.WriteHeader(201)
		enc := json.NewEncoder(w)
		err = enc.Encode(user)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}
	}
}
