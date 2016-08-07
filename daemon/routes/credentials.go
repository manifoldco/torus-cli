package routes

// This file contains routes related to credentials/secrets

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arigatomachine/cli/daemon/envelope"
	"github.com/arigatomachine/cli/daemon/primitive"
	"github.com/arigatomachine/cli/daemon/registry"
	"github.com/arigatomachine/cli/daemon/session"
)

func credentialsGetRoute(client *registry.Client,
	s session.Session) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		trees, err := client.CredentialTree.List(
			q.Get("Name"), q.Get("path"), q.Get("pathexp"), s.ID())
		if err != nil {
			log.Printf("error retrieving credential trees: %s", err)
			encodeResponseErr(w, err)
			return
		}

		// Loop over the trees and unpack the credentials; later on we will
		// actually do real work and decrypt each of these credentials but for
		// now we just need ot return a list of them!
		creds := []envelope.Unsigned{}
		for _, tree := range trees {
			creds = append(creds, tree.Credentials...)
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(creds)
		if err != nil {
			log.Printf("error encoding credentials: %s", err)
			encodeResponseErr(w, err)
			return
		}
	}
}

func credentialsPostRoute(client *registry.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		cred := &envelope.Unsigned{}

		err := dec.Decode(&cred)
		if err != nil {
			log.Printf("error decoding credential: %s", err)
			encodeResponseErr(w, err)
			return
		}

		// Get matching credentials
		credBody := cred.Body.(*primitive.Credential)
		creds, err := client.Credentials.List(credBody.Name, "", credBody.PathExp)
		if err != nil {
			log.Printf("error retrieving previous cred: %s", err)
			encodeResponseErr(w, err)
			return
		}

		if len(creds) == 0 {
			credBody.Previous = nil
			credBody.CredentialVersion = 1
		} else {
			previousCred := creds[len(creds)-1]
			previousCredBody := previousCred.Body.(*primitive.Credential)

			credBody.Previous = previousCred.ID
			credBody.CredentialVersion = previousCredBody.CredentialVersion + 1
		}

		cred, err = client.Credentials.Create(cred)
		if err != nil {
			log.Printf("error creating credential: %s", err)
			encodeResponseErr(w, err)
			return
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(cred)
		if err != nil {
			log.Printf("error encoding credential create resp: %s", err)
			encodeResponseErr(w, err)
			return
		}
	}
}
