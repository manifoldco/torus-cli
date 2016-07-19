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

type LoginTokenRequest struct {
	Type  string `json:"type"`
	Email string `json:"email"`
}

type LoginTokenResponse struct {
	Salt  string `json:"salt"`
	Token string `json:"login_token"`
}

type AuthTokenRequest struct {
	Type      string `json:"type"`
	TokenHMAC string `json:"login_token_hmac"`
}

type AuthTokenResponse struct {
	Token string `json:"auth_token"`
}
