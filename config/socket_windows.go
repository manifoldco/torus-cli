package config

import "os"

func setTransportAddress(cfg *Config) {
	transportAddress := os.Getenv("TORUS_TCP_ADDRESS")
	if transportAddress == "" {
		transportAddress = "127.0.0.1:50"
	}

	cfg.TransportAddress = transportAddress
}
