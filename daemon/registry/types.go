package registry

type Error struct {
	StatusCode int

	Type string   `json:"type"`
	Err  []string `json:"error"`
}

func (e *Error) Error() string {
	return e.Type
}

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
