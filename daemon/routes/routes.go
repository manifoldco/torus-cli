package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-zoo/bone"

	"github.com/arigatomachine/cli/daemon/config"
	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/session"
)

func NewRouteMux(c *config.Config, s session.Session,
	t *http.Transport) *bone.Mux {

	client := http.Client{Transport: t}
	mux := bone.New()

	mux.PostFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)

		creds := Login{}
		err := dec.Decode(&creds)
		if err != nil {
			encodeResponseErr(w)
			return
		}

		if creds.Email == "" || creds.Passphrase == "" {
			w.WriteHeader(http.StatusBadRequest)
			enc := json.NewEncoder(w)
			enc.Encode(&Error{Err: "email and passphrase required"})
			return
		}

		b := &bytes.Buffer{}
		enc := json.NewEncoder(b)
		err = enc.Encode(&LoginTokenRequest{Type: TokenTypeLogin, Email: creds.Email})
		if err != nil {
			log.Printf("Error encoding login token request: %s", err)
			encodeResponseErr(w)
			return
		}

		req, err := newReq(c, "", "POST", "/tokens", b)
		if err != nil {
			log.Printf("Error building http request: %s", err)
			encodeResponseErr(w)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error making api request: %s", err)
			encodeResponseErr(w)
			return
		}

		defer resp.Body.Close()

		salt := LoginTokenResponse{}
		dec = json.NewDecoder(resp.Body)
		err = dec.Decode(&salt)
		if err != nil {
			log.Printf("Error decoding api response: %s", err)
			encodeResponseErr(w)
			return
		}

		if resp.StatusCode != 201 {
			log.Printf("Failed to get login token from server: %s", err)
			encodeResponseErr(w)
			return
		}

		hmac, err := crypto.DeriveLoginHMAC(creds.Passphrase, salt.Salt,
			salt.Token)
		if err != nil {
			log.Printf("Error generating login token hmac: %s", err)
			encodeResponseErr(w)
			return
		}

		b.Reset()
		enc = json.NewEncoder(b)
		enc.Encode(&AuthTokenRequest{Type: "auth", TokenHMAC: hmac})
		if err != nil {
			log.Printf("Error encoding auth token request: %s", err)
			encodeResponseErr(w)
			return
		}

		req, err = newReq(c, salt.Token, "POST", "/tokens", b)
		if err != nil {
			log.Printf("Error building http request: %s", err)
			encodeResponseErr(w)
			return
		}

		resp, err = client.Do(req)
		if err != nil {
			log.Printf("Error making api request: %s", err)
			encodeResponseErr(w)
			return
		}

		defer resp.Body.Close()

		auth := AuthTokenResponse{}
		dec = json.NewDecoder(resp.Body)
		err = dec.Decode(&auth)
		if err != nil {
			log.Printf("Error decoding api response: %s", err)
			encodeResponseErr(w)
			return
		}

		s.Set(creds.Passphrase, auth.Token)

		w.WriteHeader(http.StatusNoContent)
	})

	mux.PostFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		tok := s.GetToken()

		if tok == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		req, err := newReq(c, tok, "DELETE", "/tokens/"+tok, nil)
		if err != nil {
			log.Printf("Error building http request: %s", err)
			encodeResponseErr(w)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error making api request: %s", err)
			encodeResponseErr(w)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode >= 200 || resp.StatusCode < 300 {
			s.Logout()
			w.WriteHeader(http.StatusNoContent)
		} else if resp.StatusCode >= 500 {
			// On a 5XX response, we don't know for sure that the server has
			// successfully removed the auth token. Keep the copy in the daemon,
			// so the user may try again.
			encodeResponseErr(w)
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
				encodeResponseErr(w)
			}
			return
		}

		err := enc.Encode(&Status{
			Token:      s.HasToken(),
			Passphrase: s.HasPassphrase(),
		})

		if err != nil {
			encodeResponseErr(w)
		}
	})

	mux.GetFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		err := enc.Encode(&Version{Version: c.Version})
		if err != nil {
			encodeResponseErr(w)
		}
	})

	return mux
}

// if encoding has errored, our struct is either bad, or our writer
// is broken. Try writing an error back to the client, but ignore any
// problems (ie the writer is broken).
func encodeResponseErr(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	enc := json.NewEncoder(w)
	enc.Encode(&Error{Err: "Internal server error"})
}

func newReq(c *config.Config, t, m, p string, b io.Reader) (*http.Request,
	error) {
	req, err := http.NewRequest(m, c.API+p, b)
	if err != nil {
		return nil, err
	}

	if t != "" {
		req.Header.Set("Authorization", "bearer "+t)
	}
	req.Header.Set("Content-type", "application/json")

	return req, nil
}
