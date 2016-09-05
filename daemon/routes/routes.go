package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-zoo/bone"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/config"

	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/logic"
	"github.com/arigatomachine/cli/daemon/observer"
	"github.com/arigatomachine/cli/daemon/registry"
	"github.com/arigatomachine/cli/daemon/session"
)

// NewRouteMux returns a *bone.Mux responsible for handling the cli to daemon
// http api.
func NewRouteMux(c *config.Config, s session.Session, db *db.DB,
	t *http.Transport, o *observer.Observer) *bone.Mux {

	cryptoEngine := crypto.NewEngine(s, db)
	client := registry.NewClient(c.RegistryURI.String(), c.APIVersion,
		c.Version, s, t)
	lEngine := logic.NewEngine(c, s, db, cryptoEngine, client)

	mux := bone.New()

	mux.Get("/observe", o)

	mux.PostFunc("/signup", signupRoute(client, s, db))
	mux.PostFunc("/login", loginRoute(client, s, db))
	mux.PostFunc("/logout", logoutRoute(client, s))
	mux.GetFunc("/session", sessionRoute(s))

	mux.PostFunc("/keypairs/generate", keypairsGenerateRoute(lEngine, o))

	mux.GetFunc("/credentials", credentialsGetRoute(lEngine, o))
	mux.PostFunc("/credentials", credentialsPostRoute(lEngine, o))

	mux.PostFunc("/org-invites/:id/approve",
		orgInvitesApproveRoute(lEngine, o))

	mux.GetFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		err := enc.Encode(&apitypes.Version{Version: c.Version})
		if err != nil {
			encodeResponseErr(w, err)
		}
	})

	return mux
}

// if encoding has errored, our struct is either bad, or our writer
// is broken. Try writing an error back to the client, but ignore any
// problems (ie the writer is broken).
func encodeResponseErr(w http.ResponseWriter, err interface{}) {
	enc := json.NewEncoder(w)

	rErr, ok := err.(*apitypes.Error)
	if ok {
		w.WriteHeader(rErr.StatusCode)
		enc.Encode(rErr)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(&errorMsg{
			Type:  apitypes.InternalServerError,
			Error: []string{"Internal server error"},
		})
	}
}
