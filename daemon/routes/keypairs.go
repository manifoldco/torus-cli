package routes

// This file contains routes related to keypairs

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/identity"

	"github.com/manifoldco/torus-cli/daemon/logic"
	"github.com/manifoldco/torus-cli/daemon/observer"
)

type keyPairRequest struct {
	OrgID *identity.ID `json:"org_id"`
}

func keypairsGenerateRoute(engine *logic.Engine, o *observer.Observer) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		dec := json.NewDecoder(r.Body)
		genReq := keyPairRequest{}
		err := dec.Decode(&genReq)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		if genReq.OrgID == nil {
			encodeResponseErr(w, &apitypes.Error{
				Type: apitypes.BadRequestError,
				Err:  []string{"missing or invalid OrgID provided"},
			})
			return
		}

		n, err := o.Notifier(ctx, 1)
		if err != nil {
			log.Printf("Error creating Notifier: %s", err)
			encodeResponseErr(w, err)
			return
		}

		err = engine.GenerateKeypairs(ctx, n, genReq.OrgID)
		if err != nil {
			// Rely on engine for debug logging
			encodeResponseErr(w, err)
			return
		}

		n.Notify(observer.Progress, "Encryption keys uploaded", true)
		w.WriteHeader(http.StatusNoContent)
	}
}

func keypairsRevokeRoute(engine *logic.Engine, o *observer.Observer) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		dec := json.NewDecoder(r.Body)
		revReq := keyPairRequest{}
		err := dec.Decode(&revReq)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		if revReq.OrgID == nil {
			encodeResponseErr(w, &apitypes.Error{
				Type: apitypes.BadRequestError,
				Err:  []string{"missing or invalid OrgID provided"},
			})
			return
		}

		n, err := o.Notifier(ctx, 0)
		if err != nil {
			log.Printf("Error creating Notifier: %s", err)
			encodeResponseErr(w, err)
			return
		}

		err = engine.RevokeKeypairs(ctx, n, revReq.OrgID)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
