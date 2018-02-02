package routes

// This file contains routes related to credentials/secrets

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/manifoldco/torus-cli/daemon/logic"
	"github.com/manifoldco/torus-cli/daemon/observer"
	"github.com/manifoldco/torus-cli/identity"
)

func credentialsGetRoute(engine *logic.Engine, o *observer.Observer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		ctx := r.Context()
		q := r.URL.Query()
		n, err := o.Notifier(ctx, 1)
		if err != nil {
			log.Printf("Error creating parent Notifier: %s", err)
			encodeResponseErr(w, err)
			return
		}

		path := q.Get("path")
		pathexp := q.Get("pathexp")
		skip := q.Get("skip-decryption") == "true"
		if path == "" && pathexp == "" {
			err = errors.New("missing path or pathexp")
			log.Printf("Error constructing request: %s", err)
			encodeResponseErr(w, err)
			return
		}

		teamStrIDs := q["team_id"]

		var teamIDs []identity.ID
		for _, i := range teamStrIDs {
			id, err := identity.DecodeFromString(i)
			if err != nil {
				err = errors.New("Failed to decode ID " + i)
				log.Printf("Failed to decode ID "+i, err)
				encodeResponseErr(w, err)
				return
			}
			teamIDs = append(teamIDs, id)
		}

		var creds []logic.PlaintextCredentialEnvelope
		if path != "" {
			creds, err = engine.RetrieveCredentials(ctx, n, &path, nil, teamIDs, skip)
		} else {
			creds, err = engine.RetrieveCredentials(ctx, n, nil, &pathexp, teamIDs, skip)
		}
		if err != nil {
			// Rely on logs inside engine for debugging
			encodeResponseErr(w, err)
			return
		}

		n.Notify(observer.Finished, "Completed Operation", true)

		enc := json.NewEncoder(w)
		err = enc.Encode(creds)
		if err != nil {
			log.Printf("error encoding credentials: %s", err)
			encodeResponseErr(w, err)
			return
		}
	}
}

func credentialsPostRoute(engine *logic.Engine, o *observer.Observer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		creds := []*logic.PlaintextCredentialEnvelope{}

		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&creds)
		if err != nil {
			log.Printf("error decoding credential: %s", err)
			encodeResponseErr(w, err)
			return
		}

		n, err := o.Notifier(ctx, 1)
		if err != nil {
			log.Printf("error constructing Notifier: %s", err)
			encodeResponseErr(w, err)
			return
		}

		creds, err = engine.AppendCredentials(ctx, n, creds)
		if err != nil {
			// Rely on logs inside engine for debugging
			encodeResponseErr(w, err)
			return
		}

		n.Notify(observer.Finished, "Completed Operation", true)

		enc := json.NewEncoder(w)
		err = enc.Encode(creds)
		if err != nil {
			log.Printf("error encoding credential create resp: %s", err)
			encodeResponseErr(w, err)
			return
		}
	}
}
