package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-zoo/bone"

	"github.com/arigatomachine/cli/daemon/config"
	"github.com/arigatomachine/cli/daemon/session"
)

func login(w http.ResponseWriter, r *http.Request) {

}

type Version struct {
	Version string `json:"version"`
}

type Status struct {
	Token      bool `json:"token"`
	Passphrase bool `json:"passphrase"`
}

type Error struct {
	Message string `json:"message"`
}

func NewRouteMux(c *config.Config, s session.Session,
	t *http.Transport) *bone.Mux {

	client := http.Client{Transport: t}
	mux := bone.New()
	mux.PostFunc("/login", login)
	mux.PostFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		tok := s.GetToken()

		if tok == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		req, err := http.NewRequest(
			"DELETE",
			c.API+"/tokens/"+tok,
			nil,
		)
		if err != nil {
			log.Printf("Error building http request: %s", err)
			encodeResponseErr(w)
			return
		}

		req.Header.Set("Authorization", "bearer "+tok)
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

	mux.GetFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
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
	enc.Encode(&Error{Message: "Internal server error"})
}
