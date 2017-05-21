package routes

import (
	"net/http"

	"github.com/manifoldco/torus-cli/api"
)

func machineCreateRoute(api *api.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}
}
