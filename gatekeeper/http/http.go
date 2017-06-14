package http

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/facebookgo/httpdown"
	"github.com/go-zoo/bone"

	"github.com/manifoldco/torus-cli/api"
	"github.com/manifoldco/torus-cli/config"
	"github.com/manifoldco/torus-cli/gatekeeper/routes"
)

type gatekeeperDefaults struct {
	Org  string
	Team string
}

// Gatekeeper exposes an HTTP interface over a normal TCP socket. It handles
// machine creation, for machines bootstrapped with `torus bootstrap`
type Gatekeeper struct {
	defaults gatekeeperDefaults
	s        *http.Server
	hd       httpdown.Server
	c        *config.Config
	api      *api.Client
}

// NewGatekeeper returns a new Gatekeeper.
func NewGatekeeper(org, team, certpath, keypath string, cfg *config.Config, api *api.Client) (*Gatekeeper, error) {
	server := &http.Server{
		Addr: cfg.GatekeeperAddress,
	}

	keypair, err := tlsKeypair(certpath, keypath)
	if err != nil {
		log.Printf("Starting Gatekeeper without SSL: %s", err)
	} else {
		if err != nil {
			return nil, err
		}

		tlsConfig := &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{*keypair},
		}

		server.TLSConfig = tlsConfig
	}

	g := &Gatekeeper{
		defaults: gatekeeperDefaults{
			Org:  org,
			Team: team,
		},
		s:   server,
		c:   cfg,
		api: api,
	}

	return g, nil
}

// Listen listens on a TCP port for HTTP machine requests
func (g *Gatekeeper) Listen() error {
	mux := bone.New()

	mux.Post("/v0/machine/aws", routes.AWSBootstrapRoute(g.defaults.Org, g.defaults.Team, g.api))

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

func tlsKeypair(certpath, keypath string) (*tls.Certificate, error) {
	if certpath == "" {
		return nil, fmt.Errorf("no certificate provided")
	}

	certBytes, err := ioutil.ReadFile(certpath)
	if err != nil {
		return nil, fmt.Errorf("unable to read certificate: %s", err)
	}

	if keypath == "" {
		return nil, fmt.Errorf("no certificate key provided")
	}

	keyBytes, err := ioutil.ReadFile(keypath)
	if err != nil {
		return nil, fmt.Errorf("unable to read keyfile: %s", err)
	}

	cert, err := tls.X509KeyPair(certBytes, keyBytes)
	return &cert, err
}
