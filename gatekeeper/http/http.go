package http

import (
	"log"
	"net/http"

	"github.com/facebookgo/httpdown"
	"github.com/go-zoo/bone"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/gatekeeper/routes"
)

// Gatekeeper exposes an HTTP interface over a normal TCP socket. It handles
// machine creation, for machines bootstrapped with `torus bootstrap`
type Gatekeeper struct {
	s   *http.Server
	hd  httpdown.Server
	c   *config.Config
	api *api.Client
}

// NewGatekeeper returns a new Gatekeeper.
func NewGatekeeper(cfg *config.Config, api *api.Client) *Gatekeeper {
	server := &http.Server{
		Addr: cfg.GatekeeperAddress,
	}

	g := &Gatekeeper{
		s:   server,
		c:   cfg,
		api: api,
	}

	return g
}

// Listen listens on a TCP port for HTTP machine requests
func (g *Gatekeeper) Listen() error {
	mux := bone.New()

	mux.SubRoute("/v0", routes.NewRoutesMux(g.c, g.api))

	g.s.Handler = loggingHandler(mux)
	h := httpdown.HTTP{}

	var err error
	g.hd, err = h.ListenAndServe(g.s)
	if err != nil {
		return err
	}

	return g.hd.Wait()
}

// Close gracefully stops the HTTP server
func (g *Gatekeeper) Close() error {
	return g.hd.Stop()
}

// Addr returns the address of the running Gatekeeper service
func (g *Gatekeeper) Addr() string {
	return g.s.Addr
}

// loggingHandler logs requests in the format: METHOD URL
func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		next.ServeHTTP(w, r)
		log.Printf("%s %s", r.Method, p)
	})
}
