package api

import (
	"net"
	"net/http"

	"github.com/manifoldco/torus-cli/config"
	"github.com/natefinch/npipe"
)

func newTransport(cfg *config.Config) *http.Transport {
	return &http.Transport{
		Dial: func(network, address string) (net.Conn, error) {
			return npipe.DialTimeout(cfg.TransportAddress, dialTimeout)
		},
	}
}
