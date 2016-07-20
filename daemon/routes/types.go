package routes

type Login struct {
	Email      string `json:"email"`
	Passphrase string `json:"passphrase"`
}

type Version struct {
	Version string `json:"version"`
}

type Status struct {
	Token      bool `json:"token"`
	Passphrase bool `json:"passphrase"`
}

type Error struct {
	Err     string `json:"error"`
	Message string `json:"message"`
}

const (
	TokenTypeLogin = "login"
	TokenTypeAuth  = "auth"
)
