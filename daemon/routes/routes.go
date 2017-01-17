package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-zoo/bone"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/config"

	"github.com/manifoldco/torus-cli/daemon/db"
	"github.com/manifoldco/torus-cli/daemon/logic"
	"github.com/manifoldco/torus-cli/daemon/observer"
	"github.com/manifoldco/torus-cli/daemon/registry"
	"github.com/manifoldco/torus-cli/daemon/session"
	"github.com/manifoldco/torus-cli/daemon/updates"
)

// NewRouteMux returns a *bone.Mux responsible for handling the cli to daemon
// http api.
func NewRouteMux(c *config.Config, s session.Session, db *db.DB,
	t *http.Transport, o *observer.Observer, client *registry.Client, lEngine *logic.Engine, uEngine *updates.Engine) *bone.Mux {

	mux := bone.New()

	mux.Get("/observe", o)

	mux.PostFunc("/signup", signupRoute(client, s, db))
	mux.PostFunc("/login", loginRoute(lEngine))
	mux.PostFunc("/logout", logoutRoute(lEngine))
	mux.GetFunc("/session", sessionRoute(s))
	mux.GetFunc("/self", selfRoute(s))
	mux.PatchFunc("/self", updateSelfRoute(client, s, lEngine))

	mux.PostFunc("/machines", machinesCreateRoute(client, s, lEngine, o))

	mux.PostFunc("/keypairs/generate", keypairsGenerateRoute(lEngine, o))
	mux.PostFunc("/keypairs/revoke", keypairsRevokeRoute(lEngine, o))

	mux.GetFunc("/credentials", credentialsGetRoute(lEngine, o))
	mux.PostFunc("/credentials", credentialsPostRoute(lEngine, o))

	mux.PostFunc("/org-invites/:id/approve",
		orgInvitesApproveRoute(lEngine, o))

	mux.GetFunc("/worklog", worklogListRoute(lEngine, o))
	mux.GetFunc("/worklog/:id", worklogGetRoute(lEngine, o))
	mux.PostFunc("/worklog/:id", worklogResolveRoute(lEngine, o))

	mux.GetFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		err := enc.Encode(&apitypes.Version{Version: c.Version})
		if err != nil {
			encodeResponseErr(w, err)
		}
	})

	mux.GetFunc("/updates", func(w http.ResponseWriter, r *http.Request) {
		needsUpdate, version := uEngine.VersionInfo()
		payload := &apitypes.UpdateInfo{
			NeedsUpdate: needsUpdate,
			Version:     version,
		}
		enc := json.NewEncoder(w)
		err := enc.Encode(payload)
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
