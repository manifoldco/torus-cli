package api

import (
	"net"
	"net/http"

	"github.com/Microsoft/go-winio"
	"github.com/manifoldco/torus-cli/config"
)

func newTransport(cfg *config.Config) *http.Transport {
	return &http.Transport{
		Dial: func(network, address string) (net.Conn, error) {
			return winio.DialPipe(cfg.TransportAddress, nil)
		},
	}
}
