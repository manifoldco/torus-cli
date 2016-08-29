package routes

// This file contains routes related to credentials/secrets

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arigatomachine/cli/daemon/logic"
	"github.com/arigatomachine/cli/daemon/observer"
)

func credentialsGetRoute(engine *logic.Engine, o *observer.Observer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		q := r.URL.Query()
		n, err := o.Notifier(ctx, 1)
		if err != nil {
			log.Printf("Error creating parent Notifier: %s", err)
			encodeResponseErr(w, err)
			return
		}

		creds, err := engine.RetrieveCredentials(ctx, n, q.Get("path"))
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
		cred := &logic.PlaintextCredentialEnvelope{}

		dec := json.NewDecoder(r.Body)
		err := dec.Decode(cred)
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

		cred, err = engine.AppendCredential(ctx, n, cred)
		if err != nil {
			// Rely on logs inside engine for debugging
			encodeResponseErr(w, err)
			return
		}

		n.Notify(observer.Finished, "Completed Operation", true)

		enc := json.NewEncoder(w)
		err = enc.Encode(cred)
		if err != nil {
			log.Printf("error encoding credential create resp: %s", err)
			encodeResponseErr(w, err)
			return
		}
	}
}
