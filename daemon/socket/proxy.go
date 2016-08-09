package socket

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/facebookgo/httpdown"
	"github.com/go-zoo/bone"

	"github.com/arigatomachine/cli/daemon/config"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/routes"
	"github.com/arigatomachine/cli/daemon/session"
)

// AuthProxy exposes an HTTP interface over a domain socket.
// It handles adding auth headers to requests on the `/proxy` endpoint to
// directly proxy requests from the cli to the registry, and exposes an
// interface over `/v1` for secure and composite operations.
type AuthProxy struct {
	u    *url.URL
	l    net.Listener
	s    httpdown.Server
	c    *config.Config
	db   *db.DB
	sess session.Session
}

// NewAuthProxy returns a new AuthProxy. It will return an error if creation
// of the domain socket fails, or the upstream registry URL is misconfigured.
func NewAuthProxy(c *config.Config, sess session.Session,
	db *db.DB) (*AuthProxy, error) {

	l, err := makeSocket(c.SocketPath)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(c.API)
	if err != nil {
		return nil, err
	}

	return &AuthProxy{u: u, l: l, c: c, sess: sess, db: db}, nil
}

// Listen starts the main loop of the AuthProxy. It returns on error, or when
// the AuthProxy is closed.
func (p *AuthProxy) Listen() error {
	mux := bone.New()
	// XXX: We must validate certs, and figure something out for local dev
	// see https://github.com/arigatomachine/cli/issues/432
	t := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	proxy := &httputil.ReverseProxy{
		Transport: t,
		Director: func(r *http.Request) {
			r.URL.Scheme = p.u.Scheme
			r.URL.Host = p.u.Host
			r.Host = p.u.Host
			r.URL.Path = r.URL.Path[6:]

			tok := p.sess.Token()
			if tok != "" {
				r.Header["Authorization"] = []string{"Bearer " + tok}
			}

			r.Header["User-Agent"] = []string{"Ag-Daemon/" + p.c.Version}
			r.Header["X-Registry-Version"] = []string{p.c.APIVersion}
		},
	}

	mux.Handle("/proxy/", proxy)
	mux.SubRoute("/v1", routes.NewRouteMux(p.c, p.sess, p.db, t))

	h := httpdown.HTTP{}
	p.s = h.Serve(&http.Server{Handler: loggingHandler(mux)}, p.l)

	return p.s.Wait()
}

// Close gracefully closes the socket, ensuring all requests are finished
// within the timeout.
func (p *AuthProxy) Close() error {
	return p.s.Stop()
}

// Addr returns the domain socket this proxy is listening on.
func (p *AuthProxy) Addr() string {
	return p.l.Addr().String()
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		next.ServeHTTP(w, r)
		log.Printf("%s %s", r.Method, p)
	})
}

func makeSocket(socketPath string) (net.Listener, error) {
	absPath, err := filepath.Abs(socketPath)
	if err != nil {
		return nil, err
	}

	// Attempt to remove an existing socket at this path if it exists.
	// Guarding against a server already running is outside the scope of this
	// module.
	err = os.Remove(absPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	l, err := net.Listen("unix", absPath)
	if err != nil {
		return nil, err
	}

	// Does not guarantee security; BSD ignores file permissions for sockets
	// see https://github.com/arigatomachine/cli/issues/76 for details
	if err = os.Chmod(socketPath, 0700); err != nil {
		return nil, err
	}

	return l, nil
}
