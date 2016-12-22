package routes

// This file contains routes related to keypairs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
	"github.com/manifoldco/torus-cli/registry"

	"github.com/manifoldco/torus-cli/daemon/logic"
	"github.com/manifoldco/torus-cli/daemon/observer"
	"github.com/manifoldco/torus-cli/daemon/session"
)

func machinesCreateRoute(client *registry.Client, session session.Session,
	engine *logic.Engine, o *observer.Observer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		dec := json.NewDecoder(r.Body)
		req := apitypes.MachinesCreateRequest{}
		err := dec.Decode(&req)
		if err != nil {
			log.Printf("Error decoding request: %s", err)
			encodeResponseErr(w, err)
			return
		}

		n, err := o.Notifier(ctx, 3)
		if err != nil {
			log.Printf("Error creating Notifier: %s", err)
			encodeResponseErr(w, err)
			return
		}

		msg := fmt.Sprintf("Creating machine \"%s\"", req.Name)
		n.Notify(observer.Progress, msg, true)

		machine, memberships, err := createMachine(req.OrgID, req.TeamID, session.ID(), req.Name)
		if err != nil {
			log.Printf("Error creating machine %s: %s", req.Name, err)
			encodeResponseErr(w, err)
			return
		}

		token, err := engine.Machine.CreateToken(ctx, n, machine, req.Secret)
		if err != nil {
			log.Printf("Error creating machine token: %s", err)
			encodeResponseErr(w, err)
			return
		}

		n.Notify(observer.Progress, "Uploading token keypairs", true)

		segment, err := client.Machines.Create(ctx, machine, memberships, token)
		if err != nil {
			log.Printf("Error creating machine with registry: %s", err)
			encodeResponseErr(w, err)
			return
		}

		err = engine.Machine.EncodeToken(ctx, n, token.Token)
		if err != nil {
			log.Printf("Error encoding token into keyrings: %s", err)
			encodeResponseErr(w, err)
			return
		}

		n.Notify(observer.Finished, "Machine created", true)

		enc := json.NewEncoder(w)
		err = enc.Encode(segment)
		if err != nil {
			log.Printf("Error encoding MachineSegment: %s", err)
			encodeResponseErr(w, err)
			return
		}
	}
}

// createMachine generates a Machine object and associated Membership objects
// to be uploaded to the registry in the future.
func createMachine(orgID, teamID, creatorID *identity.ID, name string) (
	*envelope.Machine, []envelope.Membership, error) {

	machineBody := &primitive.Machine{
		Name:        name,
		OrgID:       orgID,
		CreatedBy:   creatorID,
		Created:     time.Now().UTC(),
		DestroyedBy: nil,
		Destroyed:   nil,
		State:       primitive.MachineActiveState,
	}
	machineID, err := identity.NewMutable(machineBody)
	if err != nil {
		return nil, nil, err
	}
	machine := envelope.Machine{
		ID:      &machineID,
		Version: 1,
		Body:    machineBody,
	}

	machineTeamID := identity.DeriveMutable(&primitive.Team{}, orgID, primitive.DerivableMachineTeamSymbol)
	teamIDList := []*identity.ID{&machineTeamID, teamID}
	var memberships []envelope.Membership
	for _, curTeamID := range teamIDList {
		body := primitive.Membership{
			OwnerID: &machineID,
			OrgID:   orgID,
			TeamID:  curTeamID,
		}

		ID, err := identity.NewMutable(&body)
		if err != nil {
			return nil, nil, err
		}

		membership := envelope.Membership{
			ID:      &ID,
			Version: 1,
			Body:    &body,
		}
		memberships = append(memberships, membership)
	}

	return &machine, memberships, nil
}
