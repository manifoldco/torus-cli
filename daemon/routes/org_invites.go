package routes

// The file contains routes related to org invitations

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-zoo/bone"

	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/identity"
	"github.com/arigatomachine/cli/daemon/registry"
	"github.com/arigatomachine/cli/daemon/session"
)

func orgInvitesApproveRoute(client *registry.Client, s session.Session,
	db *db.DB, engine *crypto.Engine) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		inviteID, err := identity.DecodeFromString(bone.GetValue(r, "id"))
		if err != nil {
			log.Printf("Could not approve org invite; invalid id: %s", err)
			encodeResponseErr(w, err)
			return
		}

		invite, err := client.OrgInvite.Approve(&inviteID)
		if err != nil {
			log.Printf("Could not approve org invite: %s", err)
			encodeResponseErr(w, err)
			return
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(invite)
		if err != nil {
			log.Printf("error encoding invite approve resp: %s", err)
			encodeResponseErr(w, err)
			return
		}
	}
}
