package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/manifoldco/torus-cli/api"
	baseapitypes "github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/gatekeeper/apitypes"
	"github.com/manifoldco/torus-cli/gatekeeper/bootstrap/aws"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// AWSBootstrapRoute is the http.HandlerFunc for handling Bootstrap requests
// from AWS
func AWSBootstrapRoute(orgName, teamName string, api *api.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		dec := json.NewDecoder(r.Body)
		req := apitypes.AWSBootstrapRequest{}
		err := dec.Decode(&req)
		if err != nil {
			log.Printf("Error decoding request: %s", err)
			writeError(w, http.StatusBadRequest, err)
			return
		}

		v := aws.Verifier{
			Identity:  req.Identity,
			Signature: req.Signature,
		}

		if err := v.Verify(); err != nil {
			log.Printf("Instance verification failed: %s", err)
			writeError(w, http.StatusBadRequest, fmt.Errorf("instance verification failed: %s", err))
			return
		}

		if req.Machine.Org != "" {
			orgName = req.Machine.Org
		}
		if orgName == "" {
			log.Printf("No organization provided to bootstrap")
			writeError(w, http.StatusBadRequest, fmt.Errorf("no organization provided to bootstrap"))
			return
		}

		org, newOrg, err := selectOrg(ctx, api, orgName)
		if !newOrg {
			if org == nil {
				log.Print("No organization found")
				writeError(w, http.StatusNotFound, err)
				return
			}
		}

		if req.Machine.Team != "" {
			teamName = req.Machine.Team
		}
		if teamName == "" {
			log.Print("No team provided to bootstrap")
			writeError(w, http.StatusBadRequest, fmt.Errorf("no team provided by bootstrap"))
			return
		}

		team, newTeam, err := selectTeam(ctx, api, org.ID, teamName)
		if !newTeam {
			if team == nil {
				log.Printf("No team found")
				writeError(w, http.StatusNotFound, err)
				return
			}
		}

		if newOrg {
			var err error
			org, err = api.Orgs.Create(ctx, orgName)
			if err != nil {
				log.Print("Could not create org")
				writeError(w, http.StatusInternalServerError, err)
				return
			}

			err = api.KeyPairs.Create(ctx, org.ID, nil)
			if err != nil {
				log.Printf("Unable to generate org keypairs: %s", err)
				writeError(w, http.StatusInternalServerError, err)
				return
			}

			log.Printf("Org %s created", orgName)
		}

		if newTeam {
			var err error
			team, err = api.Teams.Create(ctx, org.ID, teamName, primitive.MachineTeamType)
			if err != nil {
				log.Printf("Could not create team")
				writeError(w, http.StatusInternalServerError, err)
				return
			}

			log.Printf("Team %s created", teamName)
		}

		machine, tokenSecret, err := api.Machines.Create(ctx, org.ID, team.ID, req.Machine.Name, nil)
		if err != nil {
			log.Printf("Unable to create machine: %s", err)
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		if len(machine.Tokens) < 1 {
			log.Printf("Error generating machine credentials")
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		enc := json.NewEncoder(w)

		w.WriteHeader(http.StatusCreated)

		resp := apitypes.BootstrapResponse{
			Token:  machine.Tokens[0].Token.ID,
			Secret: tokenSecret,
		}

		enc.Encode(resp)
	}
}

// writeError returns a bootstrapping error
func writeError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	resp := baseapitypes.Error{
		Type: baseapitypes.InternalServerError,
		Err:  []string{err.Error()},
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
func selectTeam(ctx context.Context, api *api.Client, orgID *identity.ID, name string) (*envelope.Team, bool, error) {
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
