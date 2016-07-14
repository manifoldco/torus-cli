package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-zoo/bone"

	"github.com/arigatomachine/cli/daemon/config"
	"github.com/arigatomachine/cli/daemon/session"
)

func login(w http.ResponseWriter, r *http.Request) {

}

func logout(w http.ResponseWriter, r *http.Request) {

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

func NewRouteMux(c *config.Config, s session.Session) *bone.Mux {
	mux := bone.New()
	mux.PostFunc("/login", login)
	mux.PostFunc("/logout", logout)

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
