package routes

import (
	"net/http"

	"github.com/manifoldco/torus-cli/api"
)

// awsMachineBoot
func awsBootstrapRoute(api *api.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}
}
