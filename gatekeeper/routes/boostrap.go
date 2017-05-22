package routes

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"fmt"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/gatekeeper/apitypes"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// awsMachineBoot
func awsBootstrapRoute(orgName, teamName string, api *api.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		dec := json.NewDecoder(r.Body)
		req := apitypes.AWSBootstrapRequest{}
		err := dec.Decode(&req)
		if err != nil {
			log.Printf("Error decoding request: %s", err)
			writeError(w, err)
			return
		}

		if req.Machine.Org != "" {
			orgName = req.Machine.Org
		}
		if orgName == "" {
			log.Printf("No organization provided to bootstrap: %s", err)
			writeError(w, fmt.Errorf("no organization provided to bootstrap"))
			return
		}

		org, newOrg, err := selectOrg(ctx, api, orgName)
		if !newOrg {
			if org == nil {
				log.Print("No organization found")
				writeError(w, err)
				return
			}
		}

		if req.Machine.Team != "" {
			teamName = req.Machine.Team
		}
		if teamName == "" {
			log.Print("No team provided to bootstrap")
			writeError(w, fmt.Errorf("no team provided by bootstra"))
			return
		}

		team, newTeam, err := selectTeam(ctx, api, org.ID, teamName)
		if !newTeam {
			if team == nil {
				log.Printf("No team found")
				writeError(w, err)
				return
			}
		}

		if newOrg {
			var err error
			org, err = api.Orgs.Create(ctx, orgName)
			if err != nil {
				log.Print("Could not create org")
				writeError(w, err)
				return
			}

			err = api.KeyPairs.Create(ctx, org.ID, nil)
			if err != nil {
				log.Printf("Unable to generate org keypairs: %s", err)
				writeError(w, err)
				return
			}

			log.Printf("Org %s created", orgName)
		}

		if newTeam {
			var err error
			team, err = api.Teams.Create(ctx, org.ID, teamName, primitive.MachineTeamType)
			if err != nil {
				log.Printf("Could not create team")
				writeError(w, err)
				return
			}

			log.Printf("Team %s created", teamName)
		}

		machine, tokenSecret, err := api.Machines.Create(ctx, org.ID, team.ID, req.Machine.Name, nil)
		if err != nil {
			log.Printf("Unable to create machine: %s", err)
			writeError(w, err)
			return
		}

		enc := json.NewEncoder(w)

		w.WriteHeader(http.StatusOK)

		resp := apitypes.BootstrapResponse{
			Token:  machine.Machine.ID,
			Secret: tokenSecret,
		}

		enc.Encode(resp)
	}
}

// writeError returns a bootstrapping error
func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	enc := json.NewEncoder(w)
	resp := apitypes.BootstrapResponse{
		Error: err.Error(),
	}

	enc.Encode(resp)
}

// selectOrg selects (or marks for creation) the org used to create the Machine
func selectOrg(ctx context.Context, api *api.Client, name string) (*envelope.Org, bool, error) {
	orgs, err := api.Orgs.List(ctx)
	if err != nil {
		return nil, false, err
	}

	for _, o := range orgs {
		if o.Body.Name == name {
			return &o, false, nil
		}
	}

	return nil, true, nil
}

// selectTeam selects (or marks for creation) the team used to create the Machine
func selectTeam(
	ctx context.Context,
	api *api.Client,
	orgID *identity.ID,
	name string,
) (*envelope.Team, bool, error) {
	teams, err := api.Teams.List(ctx, orgID, "", primitive.MachineTeamName)
	if err != nil {
		return nil, false, err
	}

	for _, t := range teams {
		if t.Body.Name == name {
			return &t, false, nil
		}
	}

	return nil, true, nil
}
