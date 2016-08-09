package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-zoo/bone"

	"github.com/arigatomachine/cli/daemon/config"
	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/registry"
	"github.com/arigatomachine/cli/daemon/session"
)

// NewRouteMux returns a *bone.Mux responsible for handling the cli to daemon
// http api.
func NewRouteMux(c *config.Config, s session.Session, db *db.DB,
	t *http.Transport) *bone.Mux {

	engine := crypto.NewEngine(s, db)
	client := registry.NewClient(c.API, c.APIVersion, c.Version, s, t)
	mux := bone.New()

	mux.PostFunc("/login", loginRoute(client, s, db))
	mux.PostFunc("/logout", logoutRoute(client, s))
	mux.GetFunc("/session", sessionRoute(s))

	mux.PostFunc("/keypairs/generate",
		keypairsGenerateRoute(client, s, db, engine))

	mux.GetFunc("/credentials", credentialsGetRoute(client, s, engine))
	mux.PostFunc("/credentials", credentialsPostRoute(client, s, engine))

	mux.PostFunc("/org-invites/:id/approve", orgInvitesApproveRoute(client, s, db, engine))

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
