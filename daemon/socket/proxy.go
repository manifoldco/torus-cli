package socket

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/facebookgo/httpdown"
	"github.com/go-zoo/bone"
	"github.com/satori/go.uuid"

	"github.com/arigatomachine/cli/daemon/config"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/observer"
	"github.com/arigatomachine/cli/daemon/registry"
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
	o    *observer.Observer
}

// NewAuthProxy returns a new AuthProxy. It will return an error if creation
// of the domain socket fails, or the upstream registry URL is misconfigured.
func NewAuthProxy(c *config.Config, sess session.Session,
	db *db.DB) (*AuthProxy, error) {

	l, err := makeSocket(c.SocketPath)
	if err != nil {
		return nil, err
	}

	return &AuthProxy{
		u:    c.RegistryURI,
		l:    l,
		c:    c,
		db:   db,
		sess: sess,
		o:    observer.New(),
	}, nil
}

// Listen starts the main loop of the AuthProxy. It returns on error, or when
// the AuthProxy is closed.
func (p *AuthProxy) Listen() error {
	mux := bone.New()
	t := &http.Transport{TLSClientConfig: &tls.Config{
		ServerName: strings.Split(p.c.RegistryURI.Host, ":")[0],
		RootCAs:    p.c.CABundle,
	}}

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

	go p.o.Start()

	mux.HandleFunc("/proxy/", proxyCanceler(proxy))
	mux.SubRoute("/v1", routes.NewRouteMux(p.c, p.sess, p.db, t, p.o))

	h := httpdown.HTTP{}
	p.s = h.Serve(&http.Server{Handler: requestIDHandler(loggingHandler(mux))}, p.l)

	return p.s.Wait()
}

// Close gracefully closes the socket, ensuring all requests are finished
// within the timeout.
func (p *AuthProxy) Close() error {
	p.o.Stop()
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

func requestIDHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.NewV4().String()
		}
		ctx := context.WithValue(r.Context(), "id", id)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
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

// proxyCanceler supports canceling proxied requests via a timeout, and
// returning a custom error response.
func proxyCanceler(proxy http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancelFunc()

		cw := cancelingProxyResponseWriter{
			redirect: false,
			written:  false,
			rw:       w,
		}

		r = r.WithContext(ctx)
		done := make(chan bool)
		go func() {
			proxy.ServeHTTP(&cw, r)
			close(done)
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			if ctx.Err() != context.DeadlineExceeded {
				return
			}

			cw.redirect = true
			if !cw.written {
				w.WriteHeader(http.StatusRequestTimeout)

				enc := json.NewEncoder(w)
				err := enc.Encode(&registry.Error{
					Type: "request_timeout",
					Err:  []string{"Request timed out"},
				})
				if err != nil {
					log.Printf("Error writing response timeout: %s", err)
				}
			}
		}
	}
}

// cancelingProxyResponseWriter Wraps a regular ResponseWriter to allow it to
// be canceled, discarding anything written to it, providing it has not yet
// been written to.
type cancelingProxyResponseWriter struct {
	redirect bool
	written  bool
	rw       http.ResponseWriter
}

func (c *cancelingProxyResponseWriter) Header() http.Header {
	return c.rw.Header()
}

func (c *cancelingProxyResponseWriter) Write(b []byte) (int, error) {
	if c.redirect && !c.written {
		return len(b), nil
	}
	c.written = true
	return c.rw.Write(b)
}

func (c *cancelingProxyResponseWriter) WriteHeader(s int) {
	if c.redirect && !c.written {
		return
	}

	c.written = true
	c.rw.WriteHeader(s)
}
