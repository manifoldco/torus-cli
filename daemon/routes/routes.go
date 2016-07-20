package routes

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

		req, err := client.NewRequest("POST", "/tokens",
			&registry.LoginTokenRequest{
				Type:  TokenTypeLogin,
				Email: creds.Email,
			})
		if err != nil {
			log.Printf("Error building http request: %s", err)
			encodeResponseErr(w, err)
			return
		}

		salt := registry.LoginTokenResponse{}
		resp, err := client.Do(req, &salt)
		if err != nil && resp != nil && resp.StatusCode != 201 {
			log.Printf("Failed to get login token from server: %s", err)
			encodeResponseErr(w, err)
			return
		} else if err != nil {
			log.Printf("Error making api request: %s", err)
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

		req, err = client.NewTokenRequest(salt.Token, "POST", "/tokens",
			&registry.AuthTokenRequest{Type: "auth", TokenHMAC: hmac})
		if err != nil {
			log.Printf("Error building http request: %s", err)
			encodeResponseErr(w, err)
			return
		}

		auth := registry.AuthTokenResponse{}
		resp, err = client.Do(req, &auth)
		if err != nil {
			log.Printf("Error making api request: %s", err)
			encodeResponseErr(w, err)
			return
		}

		req, err = client.NewTokenRequest(auth.Token, "GET", "/users/self", nil)
		if err != nil {
			log.Printf("Error making api request: %s", err)
			encodeResponseErr(w, err)
			return
		}

		self := registry.SelfResponse{}
		resp, err = client.Do(req, &self)
		if err != nil {
			log.Printf("Error making api request: %s", err)
			encodeResponseErr(w, err)
			return
		}

		err = validateSelf(&self)
		if err != nil {
			log.Printf("Invalid user self: %s", err)
			encodeResponseErr(w, err)
			return
		}

		mk, err := base64.RawURLEncoding.DecodeString(self.Body.Master.Value)
		if err != nil {
			log.Printf("Could not decode master key: %s", err)
			encodeResponseErr(w, err)
			return
		}

		db.SetMasterKey(mk)
		s.Set(creds.Passphrase, auth.Token)

		w.WriteHeader(http.StatusNoContent)
	})

	mux.PostFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		tok := s.Token()

		if tok == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		req, err := client.NewTokenRequest(tok, "DELETE", "/tokens/"+tok, nil)
		if err != nil {
			log.Printf("Error building http request: %s", err)
			encodeResponseErr(w, err)
			return
		}

		resp, err := client.Do(req, nil)
		if err != nil && resp == nil {
			log.Printf("Error making api request: %s", err)
			encodeResponseErr(w, err)
			return
		}

		if resp.StatusCode >= 200 || resp.StatusCode < 300 {
			s.Logout()
			w.WriteHeader(http.StatusNoContent)
		} else if resp.StatusCode >= 500 {
			// On a 5XX response, we don't know for sure that the server has
			// successfully removed the auth token. Keep the copy in the daemon,
			// so the user may try again.
			encodeResponseErr(w, err)
		} else {
			// A 4XX error indicates either the token isn't found, or we're
			// not allowed to remove it (or the server is a teapot).
			//
			// In any case, the daemon has gotten out of sync with the server.
			// Remove our local copy of the auth token.
			log.Printf("Got 4XX removing auth token. Treating as success")
			s.Logout()
			w.WriteHeader(http.StatusNoContent)
		}
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

func validateSelf(s *registry.SelfResponse) error {
	if s.Version != 1 {
		return errors.New("version must be 1")
	}

	if s.Body == nil {
		return errors.New("missing body")
	}

	if s.Body.Master == nil {
		return errors.New("missing master key section")
	}

	if s.Body.Master.Alg != "triplesec-v3" {
		return fmt.Errorf("Unknown alg: %s", s.Body.Master.Alg)
	}

	if len(s.Body.Master.Value) == 0 {
		return errors.New("Zero length master key found")
	}

	return nil
}
