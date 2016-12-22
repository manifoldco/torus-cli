package registry

import "net/http"

func replaceAuthToken(req *http.Request, token string) {
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		req.Header.Del("Authorization")
	}
}
