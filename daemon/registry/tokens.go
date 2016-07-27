package registry

import "log"

type Tokens struct {
	client *Client
}

func (t *Tokens) PostLogin(email string) (*LoginTokenResponse, error) {
	req, err := t.client.NewRequest("POST", "/tokens",
		&LoginTokenRequest{
			Type:  "login",
			Email: email,
		})
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, err
	}

	salt := LoginTokenResponse{}
	resp, err := t.client.Do(req, &salt)
	if err != nil && resp != nil && resp.StatusCode != 201 {
		log.Printf("Failed to get login token from server: %s", err)
		return nil, err
	} else if err != nil {
		log.Printf("Error making api request: %s", err)
		return nil, err
	}

	return &salt, nil
}

func (t *Tokens) PostAuth(token, hmac string) (*AuthTokenResponse, error) {
	req, err := t.client.NewTokenRequest(token, "POST", "/tokens",
		&AuthTokenRequest{Type: "auth", TokenHMAC: hmac})
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, err
	}

	auth := AuthTokenResponse{}
	_, err = t.client.Do(req, &auth)
	if err != nil {
		log.Printf("Error making api request: %s", err)
		return nil, err
	}

	return &auth, nil
}

func (t *Tokens) Delete(token string) error {
	req, err := t.client.NewTokenRequest(token, "DELETE", "/tokens/"+token, nil)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return err
	}

	_, err = t.client.Do(req, nil)
	if err != nil {
		log.Printf("Error making api request: %s", err)
		return err
	}

	return nil
}
