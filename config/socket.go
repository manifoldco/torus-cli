// +build !windows

package config

import "path"

func setTransportAddress(cfg *Config) {
	cfg.TransportAddress = path.Join(cfg.TorusRoot, "daemon.socket")
}
