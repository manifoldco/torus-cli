package crypto

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"

	"golang.org/x/crypto/scrypt"
)

// Crypto Algorithm name constants.
const (
	Triplesec  = "triplesec-v3"
	EdDSA      = "eddsa"
	Curve25519 = "curve25519"
	EasyBox    = "easybox"
	SecretBox  = "secretbox"
)

// scrypt parameter constants
const (
	n      = 32768 // 2^15
	r      = 8
	p      = 1
	keyLen = 224
)

// DeriveLoginHMAC HMACs the provided token with a key derived from password
// and the provided base64 encoded salt.
func DeriveLoginHMAC(password, salt, token string) (string, error) {
	s := make([]byte, base64.RawURLEncoding.DecodedLen(len(salt)))
	l, err := base64.RawURLEncoding.Decode(s, []byte(salt))
	if err != nil {
		return "", err
	}

	k, err := scrypt.Key([]byte(password), s[:l], n, r, p, keyLen)
	if err != nil {
		return "", err
	}

	pwh := make([]byte, base64.RawURLEncoding.EncodedLen(32))
	base64.RawURLEncoding.Encode(pwh, k[keyLen-32:])

	mac := hmac.New(sha512.New, pwh)
	mac.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
