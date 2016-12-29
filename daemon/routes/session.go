package routes

// This file contains routes related to the user's session

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
	"github.com/manifoldco/torus-cli/registry"

	"github.com/manifoldco/torus-cli/daemon/crypto"
	"github.com/manifoldco/torus-cli/daemon/db"
	"github.com/manifoldco/torus-cli/daemon/logic"
	"github.com/manifoldco/torus-cli/daemon/session"
)

func loginRoute(engine *logic.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dec := json.NewDecoder(r.Body)

		req := apitypes.Login{}
		err := dec.Decode(&req)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		var creds apitypes.LoginCredential
		switch req.Type {
		case apitypes.UserSession:
			creds = &apitypes.UserLogin{}
		case apitypes.MachineSession:
			creds = &apitypes.MachineLogin{}
		default:
			encodeResponseErr(w, &apitypes.Error{
				Type: apitypes.BadRequestError,
				Err:  []string{"unrecognized login request type"},
			})
			return
		}

		err = json.Unmarshal(req.Credentials, creds)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		err = engine.Session.Login(ctx, creds)
		if err != nil {
			log.Printf("Could not complete login: %s", err)
			encodeResponseErr(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func logoutRoute(engine *logic.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		err := engine.Session.Logout(ctx)
		if err != nil {
			log.Printf("Could not complete logout: %s", err)
			encodeResponseErr(w, err)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func sessionRoute(s session.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		if !(s.HasToken() && s.HasPassphrase()) {
			w.WriteHeader(http.StatusNotFound)
			err := enc.Encode(&errorMsg{
				Type:  apitypes.UnauthorizedError,
				Error: []string{"Not logged in"},
			})
			if err != nil {
				encodeResponseErr(w, err)
			}
			return
		}

		err := enc.Encode(&apitypes.SessionStatus{
			Token:      s.HasToken(),
			Passphrase: s.HasPassphrase(),
		})

		if err != nil {
			encodeResponseErr(w, err)
		}
	}
}

type updateName struct {
	Name string `json:"name"`
}

type updateEmail struct {
	Email string `json:"email"`
}

type updatePassword struct {
	Password *primitive.UserPassword `json:"password"`
	Master   *primitive.MasterKey    `json:"master"`
}

func updateSelfRoute(client *registry.Client, s session.Session, e *logic.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := r.Context()
		dec := json.NewDecoder(r.Body)

		if s.Type() == apitypes.NotLoggedIn {
			encodeResponseErr(w, &apitypes.Error{
				Type: apitypes.UnauthorizedError,
				Err:  []string{"invalid login"},
			})
			return
		}

		req := apitypes.ProfileUpdate{}
		err := dec.Decode(&req)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}
		if req.Name == "" && req.Email == "" && req.Password == "" {
			encodeResponseErr(w, &apitypes.Error{
				Type: apitypes.BadRequestError,
				Err:  []string{"missing profile fields"},
			})
			return
		}

		// The fresh user object
		var result envelope.UserInf

		if req.Email != "" {
			envelope, err := client.Users.Update(c, updateEmail{Email: req.Email})
			if err != nil {
				encodeResponseErr(w, err)
				return
			}
			result = envelope
		}

		if req.Name != "" {
			envelope, err := client.Users.Update(c, updateName{Name: req.Name})
			if err != nil {
				encodeResponseErr(w, err)
				return
			}
			result = envelope
		}

		if req.Password != "" {
			// Encrypt the new password and re-encrypt the original master key
			passwordObj, masterObj, err := e.ChangePassword(c, req.Password)
			if err != nil {
				log.Printf("Error generating password object: %s", err)
				encodeResponseErr(w, err)
				return
			}
			envelope, err := client.Users.Update(c, updatePassword{Password: passwordObj, Master: masterObj})
			if err != nil {
				encodeResponseErr(w, err)
				return
			}
			result = envelope
		}

		// Update the local session to have the new user details
		s.SetIdentity(apitypes.UserSession, result, result)

		enc := json.NewEncoder(w)
		err = enc.Encode(result)
		if err != nil {
			encodeResponseErr(w, err)
		}
	}
}

func selfRoute(s session.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)

		if s.Type() == apitypes.NotLoggedIn {
			encodeResponseErr(w, &apitypes.Error{
				Type: apitypes.UnauthorizedError,
				Err:  []string{"invalid login"},
			})
			return
		}

		resp := s.Self()
		err := enc.Encode(resp)
		if err != nil {
			encodeResponseErr(w, err)
		}
	}
}

func signupRoute(client *registry.Client, s session.Session, db *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dec := json.NewDecoder(r.Body)

		signup := apitypes.Signup{}
		err := dec.Decode(&signup)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		passwordObj, masterObj, err := crypto.EncryptPasswordObject(ctx, signup.Passphrase, nil)
		if err != nil {
			log.Printf("Error generating password object: %s", err)
			encodeResponseErr(w, err)
			return
		}

		b64Salt, err := base64.NewValueFromString(passwordObj.Salt)
		if err != nil {
			log.Printf("Error casting Salt into Base64: %s", err)
			encodeResponseErr(w, err)
			return
		}

		bPassphrase := []byte(signup.Passphrase)
		keypair, err := crypto.DeriveLoginKeypair(ctx, bPassphrase, b64Salt)
		if err != nil {
			log.Printf("Error deriving login keypair: %s", err)
			encodeResponseErr(w, err)
			return
		}

		userBody := primitive.User{
			BaseUser: primitive.BaseUser{
				Username: signup.Username,
				Name:     signup.Name,
				Email:    signup.Email,
				Password: passwordObj,
				Master:   masterObj,
				State:    "unverified",
			},
			PublicKey: &primitive.LoginPublicKey{
				Salt:  keypair.Salt(),
				Value: keypair.PublicKey(),
				Alg:   crypto.EdDSA,
			},
		}

		ID, err := identity.NewMutable(&userBody)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		userObj := envelope.User{
			ID:      &ID,
			Version: 2,
			Body:    &userBody,
		}

		user, err := client.Users.Create(ctx, &userObj, signup)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		w.WriteHeader(201)
		enc := json.NewEncoder(w)
		err = enc.Encode(user)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}
	}
}
