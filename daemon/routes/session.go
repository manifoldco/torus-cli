package routes

// This file contains routes related to the user's session

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/base64"
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

		req := apitypes.Login{}
		err := dec.Decode(&req)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		var creds apitypes.LoginCredential
		switch req.Type {
		case apitypes.UserSession:
			creds = &apitypes.UserLogin{}
		case apitypes.MachineSession:
			creds = &apitypes.MachineLogin{}
		default:
			encodeResponseErr(w, &apitypes.Error{
				Type: apitypes.BadRequestError,
				Err:  []string{"unrecognized login request type"},
			})
			return
		}

		err = json.Unmarshal(req.Credentials, creds)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		if !creds.Valid() {
			encodeResponseErr(w, &apitypes.Error{
				Type: apitypes.BadRequestError,
				Err:  []string{"invalid login credentials provided"},
			})
			return
		}

		var authToken string
		switch creds.Type() {
		case apitypes.UserSession:
			authToken, err = attemptHMACLogin(ctx, client, s, db, creds)
		case apitypes.MachineSession:
			authToken, err = attemptPDPKALogin(ctx, client, s, db, creds)
		}

		self, err := client.Self.Get(ctx, authToken)
		if err != nil {
			log.Printf("Could not retrieve self: %s", err)
			encodeResponseErr(w, err)
			return
		}

		db.Set(self.Identity)
		if self.Type == apitypes.UserSession {
			db.Set(self.Auth)
		}

		err = s.Set(self.Type, self.Identity, self.Auth, creds.Passphrase(), authToken)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func attemptPDPKALogin(ctx context.Context, client *registry.Client, s session.Session, db *db.DB, creds apitypes.LoginCredential) (string, error) {
	salt, loginToken, err := client.Tokens.PostLogin(ctx, creds)
	if err != nil {
		return "", err
	}

	pw := base64.NewValue([]byte(creds.Passphrase()))
	salt64 := base64.NewValue([]byte(salt))
	keypair, err := crypto.DeriveLoginKeypair(ctx, pw, salt64)
	if err != nil {
		return "", err
	}

	log.Printf("Public Key %s", keypair.PublicKey())
	loginTokenSig := keypair.Sign([]byte(loginToken))
	return client.Tokens.PostPDPKAuth(ctx, loginToken, loginTokenSig)
}

func attemptHMACLogin(ctx context.Context, client *registry.Client, s session.Session, db *db.DB, creds apitypes.LoginCredential) (string, error) {
	salt, loginToken, err := client.Tokens.PostLogin(ctx, creds)
	if err != nil {
		return "", err
	}

	hmac, err := crypto.DeriveLoginHMAC(ctx, creds.Passphrase(), salt, loginToken)
	if err != nil {
		return "", err
	}

	return client.Tokens.PostAuth(ctx, loginToken, hmac)
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

func selfRoute(s session.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)

		if s.Type() == apitypes.NotLoggedIn {
			encodeResponseErr(w, &apitypes.Error{
				Type: apitypes.UnauthorizedError,
				Err:  []string{"invalid login"},
			})
			return
		}

		resp := s.Self()
		err := enc.Encode(resp)
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
