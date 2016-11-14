package main

// TorusEnv is a struct which is used as env vars
type TorusEnv struct {
	TorusRoot string `env:"TORUS_ROOT"`
}

// SignupData is the data used for a test participant
type SignupData struct {
	Name     string
	Email    string
	Username string
	Password string
}
