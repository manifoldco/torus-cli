package routes

// The file contains routes related to org invitations

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-zoo/bone"

	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/identity"
	"github.com/arigatomachine/cli/daemon/primitive"
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

		// Get the organization claim tree; but only the keys for the invitee
		inviteBody := invite.Body.(*primitive.OrgInvite)
		claimTrees, err := client.ClaimTree.List(inviteBody.OrgID, inviteBody.InviteeID)
		if err != nil {
			log.Printf("could not retrieve claim tree for invite approval: %s", err)
			encodeResponseErr(w, err)
			return
		}

		if len(claimTrees) != 1 {
			log.Printf("incorrect number of claim trees returned: %d", len(claimTrees))
			encodeResponseErr(w, fmt.Errorf(
				"No claim tree found for org: %s", inviteBody.OrgID))
			return
		}

		// Get all the keyrings and memberships for the current user. This way we
		// can decrypt the MEK for each and then create a new KeyringMember for
		// our wonderful and new org member!
		keyringSections, err := client.Keyring.List(inviteBody.OrgID, s.ID())
		if err != nil {
			log.Printf("could not retrieve keyring sections for user: %s", err)
			encodeResponseErr(w, err)
			return
		}

		if len(keyringSections) != 0 {
			log.Printf("we need to do some real wokr here")
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
