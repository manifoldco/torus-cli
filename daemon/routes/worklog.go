package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-zoo/bone"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/identity"

	"github.com/manifoldco/torus-cli/daemon/logic"
	"github.com/manifoldco/torus-cli/daemon/observer"
)

func worklogListRoute(engine *logic.Engine, o *observer.Observer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		orgID, err := identity.DecodeFromString(r.URL.Query().Get("org_id"))
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		items, err := engine.Worklog.List(ctx, &orgID)
		if err != nil {
			log.Printf("error getting worklog list: %s", err)
			encodeResponseErr(w, err)
			return
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(items)
		if err != nil {
			log.Printf("error encoding worklog list resp: %s", err)
			encodeResponseErr(w, err)
			return
		}
	}
}

func worklogGetRoute(engine *logic.Engine, o *observer.Observer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		orgID, err := identity.DecodeFromString(r.URL.Query().Get("org_id"))
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		ident, err := apitypes.DecodeWorklogIDFromString(bone.GetValue(r, "id"))
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		item, err := engine.Worklog.Get(ctx, &orgID, &ident)
		if err != nil {
			log.Printf("error getting worklog item: %s", err)
			encodeResponseErr(w, err)
			return
		}

		if item == nil {
			encodeResponseErr(w, notFoundError)
			return
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(item)
		if err != nil {
			log.Printf("error encoding worklog get resp: %s", err)
			encodeResponseErr(w, err)
			return
		}
	}
}

func worklogResolveRoute(engine *logic.Engine, o *observer.Observer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		orgID, err := identity.DecodeFromString(r.URL.Query().Get("org_id"))
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		ident, err := apitypes.DecodeWorklogIDFromString(bone.GetValue(r, "id"))
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		n, err := o.Notifier(ctx, 1)
		if err != nil {
			log.Printf("Error creating Notifier: %s", err)
			encodeResponseErr(w, err)
			return
		}

		res, err := engine.Worklog.Resolve(ctx, n, &orgID, &ident)
		if err != nil {
			log.Printf("error resolving worklog item: %s", err)
			encodeResponseErr(w, err)
			return
		}

		if res == nil {
			encodeResponseErr(w, notFoundError)
			return
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(res)
		if err != nil {
			log.Printf("error encoding worklog resolve resp: %s", err)
			encodeResponseErr(w, err)
			return
		}
	}
}
