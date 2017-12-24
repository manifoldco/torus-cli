package apitypes

import (
	"encoding/json"
	"strings"

	"github.com/manifoldco/go-base64"

	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
)

// ErrorType represents the string error types that the daemon and registry can
// return.
type ErrorType string

// These are the possible error types.
const (
	BadRequestError     = "bad_request"
	UnauthorizedError   = "unauthorized"
	NotFoundError       = "not_found"
	InternalServerError = "internal_server"
	NotImplementedError = "not_implemented"
)

// Error represents standard formatted API errors from the daemon or registry.
type Error struct {
	StatusCode int

	Type string   `json:"type"`
	Err  []string `json:"error"`
}

// Error implements the error interface for formatted API errors.
func (e *Error) Error() string {
	segments := strings.Split(e.Type, "_")
	errType := strings.Join(segments, " ")
	return strings.Title(errType) + ": " + strings.Join(e.Err, " ")
}

// FormatError updates an error to contain more context
func FormatError(err error) error {
	if err == nil {
		return nil
	}

	if apiErr, ok := err.(*Error); ok {
		if apiErr.Type == UnauthorizedError {
			for _, m := range apiErr.Err {
				if strings.Contains(m, "wrong identity state: unverified") {
					return NewUnverifiedError()
				}
			}

			return &Error{
				StatusCode: 401,
				Type:       UnauthorizedError,
				Err:        []string{"You are unauthorized to perform this action."},
			}
		}
	}

	return err
}

// NewUnverifiedError returns a message telling the user to verify their account before continuing
func NewUnverifiedError() *Error {
	return &Error{
		StatusCode: 401,
		Type:       UnauthorizedError,
		Err: []string{"Your account has not yet been verified.\n\n" +
			"Please check your email for your verification code and follow the enclosed instructions.\n" +
			"Once you have verified your account you may retry this operation."},
	}
}

// IsNotFoundError returns whether or not an error is a 404 result from the api.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	if apiErr, ok := err.(*Error); ok {
		return apiErr.Type == NotFoundError
	}

	return false
}

// IsUnauthorizedError returns whether or not an error is a 401 result from the api.
func IsUnauthorizedError(err error) bool {
	if err == nil {
		return false
	}

	if apiErr, ok := err.(*Error); ok {
		return apiErr.Type == UnauthorizedError
	}

	return false
}

// SessionType is the enumerated string type of sessions.
type SessionType string

// A session can represent either a machine or a user
const (
	MachineSession SessionType = "machine"
	UserSession    SessionType = "user"
	NotLoggedIn    SessionType = "no_session"
)

// Self represents the current identity and auth combination for this session
type Self struct {
	Type SessionType `json:"type"`

	// XXX: create an ident/auth interface
	Identity envelope.Envelope `json:"identity"`
	Auth     envelope.Envelope `json:"auth"`
}

// Version contains the release version of the daemon.
type Version struct {
	Version string `json:"version"`
}

// SessionStatus contains details about the user's daemon session.
type SessionStatus struct {
	Token      bool `json:"token"`
	Passphrase bool `json:"passphrase"`
}

// Login is a wrapper around a login request from the CLI to the Daemon
type Login struct {
	Type        SessionType     `json:"type"`
	Credentials json.RawMessage `json:"credentials"`
}

// LoginCredential represents an login credentials for a user or machine
type LoginCredential interface {
	Type() SessionType
	Valid() bool
	Passphrase() []byte
	Identifier() string
}

// UserLogin contains the required details for logging in to the api and daemon
// as a user.
type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"passphrase"`
}

// Type returns the type of login request
func (UserLogin) Type() SessionType {
	return UserSession
}

// Valid returns whether or not this is a valid login request
func (u *UserLogin) Valid() bool {
	return u.Email != "" && u.Password != ""
}

// Passphrase returns the "secret" or "password" component of the request
func (u *UserLogin) Passphrase() []byte {
	return []byte(u.Password)
}

// Identifier returns the identifying piece of information of the request
func (u *UserLogin) Identifier() string {
	return u.Email
}

// MachineLogin contains the required details for logging into the api and
// daemon as a machine.
type MachineLogin struct {
	TokenID *identity.ID  `json:"token_id"`
	Secret  *base64.Value `json:"secret"`
}

// Type returns the type of the login request
func (MachineLogin) Type() SessionType {
	return MachineSession
}

// Valid returns whether or not this is a valid machine login request
func (m *MachineLogin) Valid() bool {
	return m.TokenID != nil && m.Secret != nil && m.Secret.String() != ""
}

// Passphrase returns the "secret" component of the request
func (m *MachineLogin) Passphrase() []byte {
	return *m.Secret
}

// Identifier returns the identifying piece of information of the request
func (m *MachineLogin) Identifier() string {
	return m.TokenID.String()
}

// Profile contains the fields in the response for the profiles endpoint
type Profile struct {
	ID   *identity.ID `json:"id"`
	Body *struct {
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"body"`
}

// Signup contains information required for registering an account
type Signup struct {
	Name       string
	Username   string
	Email      string
	Passphrase string
	InviteCode string
	OrgName    string
	OrgInvite  bool
}

// ProfileUpdate contains the fields a user can change on their user object
type ProfileUpdate struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// InviteAccept contains data required to accept org invite
type InviteAccept struct {
	Org   string `json:"org"`
	Email string `json:"email"`
	Code  string `json:"code"`
}

// VerifyEmail contains email verification code
type VerifyEmail struct {
	Code string `json:"code"`
}

// UpdateInfo contains info about the latest version of Torus available
// for download and if that version is higher than the local running one.
type UpdateInfo struct {
	NeedsUpdate bool   `json:"needs_update"`
	Version     string `json:"version"`
}
