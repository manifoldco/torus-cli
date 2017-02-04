package api

import (
	"net"
	"net/http"

	"github.com/manifoldco/torus-cli/config"
)

func newTransport(cfg *config.Config) *http.Transport {
	return &http.Transport{
		Dial: func(network, address string) (net.Conn, error) {
			return net.Dial("tcp", "127.0.0.1:50")
		},
	}
}
