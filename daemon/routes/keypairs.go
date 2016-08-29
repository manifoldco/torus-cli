package routes

// This file contains routes related to keypairs

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/daemon/logic"
	"github.com/arigatomachine/cli/daemon/observer"
)

func keypairsGenerateRoute(engine *logic.Engine, o *observer.Observer) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		dec := json.NewDecoder(r.Body)
		genReq := keyPairGenerate{}
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

		err = engine.GenerateKeypair(ctx, n, genReq.OrgID)
		if err != nil {
			// Rely on engine for debug logging
			encodeResponseErr(w, err)
			return
		}

		n.Notify(observer.Progress, "Encryption keys uploaded", true)
		w.WriteHeader(http.StatusNoContent)
	}
}
