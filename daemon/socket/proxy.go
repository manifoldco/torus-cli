package socket

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/facebookgo/httpdown"
)

type AuthProxy struct {
	u *url.URL
	l net.Listener
	s httpdown.Server
	t TokenReader
}

type TokenReader interface {
	GetToken() string
}

func NewAuthProxy(upstream string, socketPath string, t TokenReader) (*AuthProxy, error) {
	l, err := MakeSocket(socketPath)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(upstream)
	if err != nil {
		return nil, err
	}

	return &AuthProxy{u: u, l: l, t: t}, nil
}

func (p *AuthProxy) Listen() {
	mux := http.NewServeMux()
	proxy := &httputil.ReverseProxy{
		// XXX: We must validate certs, and figure something out for local dev
		// see https://github.com/arigatomachine/cli/issues/432
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Director: func(r *http.Request) {
			r.URL.Scheme = p.u.Scheme
			r.URL.Host = p.u.Host
			r.Host = p.u.Host
			r.URL.Path = r.URL.Path[6:]

			tok := p.t.GetToken()
			if tok != "" {
				r.Header["Authorization"] = []string{"Bearer " + tok}
			}
		},
	}

	mux.Handle("/proxy/", proxy)

	h := httpdown.HTTP{}
	p.s = h.Serve(&http.Server{Handler: loggingHandler(mux)}, p.l)
}

func (p *AuthProxy) Close() error {
	return p.s.Stop()
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		log.Printf("%s %s", r.Method, r.URL.Path)
	})
}
