package routes

// The file contains routes related to org invitations

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-zoo/bone"

	"github.com/manifoldco/torus-cli/identity"

	"github.com/manifoldco/torus-cli/daemon/logic"
	"github.com/manifoldco/torus-cli/daemon/observer"
)

func orgInvitesApproveRoute(engine *logic.Engine, o *observer.Observer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		n, err := o.Notifier(ctx, 1)
		if err != nil {
			log.Printf("Error creating Notififer: %s", err)
			encodeResponseErr(w, err)
			return
		}

		inviteID, err := identity.DecodeFromString(bone.GetValue(r, "id"))
		if err != nil {
			log.Printf("Could not approve org invite; invalid id: %s", err)
			encodeResponseErr(w, err)
			return
		}

		invite, err := engine.ApproveInvite(ctx, n, &inviteID)
		if err != nil {
			// Allow engine to log debugs
			encodeResponseErr(w, err)
			return
		}

		n.Notify(observer.Finished, "Completed", true)
		enc := json.NewEncoder(w)
		err = enc.Encode(invite)
		if err != nil {
			log.Printf("error encoding invite approve resp: %s", err)
			encodeResponseErr(w, err)
			return
		}
	}
}
