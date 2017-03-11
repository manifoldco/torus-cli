package config

func setTransportAddress(cfg *Config) {
	cfg.TransportAddress = `\\.\pipe\manifoldco.torusd.sock`
}
