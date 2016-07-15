package crypto

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"

	"golang.org/x/crypto/scrypt"
)

const (
	n      = 32768 // 2^15
	r      = 8
	p      = 1
	keyLen = 224
)

func DeriveLoginHMAC(password, salt, token string) (string, error) {
	// TODO: handle err
	s := make([]byte, base64.RawURLEncoding.DecodedLen(len(salt)))
	l, err := base64.RawURLEncoding.Decode(s, []byte(salt))
	if err != nil {
		return "", err
	}

	// XXX: deal wiht error
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
