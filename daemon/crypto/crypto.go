package crypto

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"

	base64url "github.com/arigatomachine/cli/base64"
	"github.com/arigatomachine/cli/daemon/ctxutil"
	"github.com/arigatomachine/cli/primitive"

	"golang.org/x/crypto/scrypt"
)

// Crypto Algorithm name constants.
const (
	Triplesec  = "triplesec-v3"
	EdDSA      = "eddsa"
	Curve25519 = "curve25519"
	EasyBox    = "easybox"
	SecretBox  = "secretbox"
	Scrypt     = "scrypt"
)

// scrypt parameter constants
const (
	n              = 32768 // 2^15
	r              = 8
	p              = 1
	keyLen         = 224
	saltBytes      = 16
	masterKeyBytes = 256
)

// DeriveLoginHMAC HMACs the provided token with a key derived from password
// and the provided base64 encoded salt.
func DeriveLoginHMAC(ctx context.Context, password, salt, token string) (string, error) {
	key, err := derivePassword(ctx, password, salt)
	if err != nil {
		return "", err
	}

	pwh := make([]byte, base64.RawURLEncoding.EncodedLen(32))
	base64.RawURLEncoding.Encode(pwh, key)

	mac := hmac.New(sha512.New, pwh)
	mac.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func derivePassword(ctx context.Context, password, salt string) ([]byte, error) {
	var pwh []byte
	s := make([]byte, base64.RawURLEncoding.DecodedLen(len(salt)))
	l, err := base64.RawURLEncoding.Decode(s, []byte(salt))
	if err != nil {
		return pwh, err
	}

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return pwh, err
	}

	key, err := scrypt.Key([]byte(password), s[:l], n, r, p, keyLen)
	pwh = key[keyLen-32:]
	if err != nil {
		return pwh, err
	}

	err = ctxutil.ErrIfDone(ctx)
	return pwh, err
}

// EncryptPasswordObject derives the master key and password hash from password
// and salt, returning the master and password objects
func EncryptPasswordObject(ctx context.Context, password string) (primitive.UserPassword, primitive.UserMaster, error) {
	pw := primitive.UserPassword{
		Alg: Scrypt,
	}
	m := primitive.UserMaster{
		Alg: Triplesec,
	}

	// Generate 128 bit (16 byte) salt for password
	salt := make([]byte, saltBytes) // 16
	_, err := rand.Read(salt)
	if err != nil {
		return pw, m, err
	}

	// Encode salt bytes to base64url
	pw.Salt = base64.RawURLEncoding.EncodeToString(salt)

	// Create password hash bytes
	pwh, err := derivePassword(ctx, password, pw.Salt)
	if err != nil {
		return pw, m, err
	}

	// Encode password value to base64url
	pw.Value = base64url.NewValue(pwh)

	// Generate the master key of 1024 bit (256 byte)
	key := make([]byte, masterKeyBytes) // 256
	_, err = rand.Read(key)
	if err != nil {
		return pw, m, err
	}

	// Encrypt master key with password bytes
	ts, err := newTriplesec(ctx, []byte(password))
	if err != nil {
		return pw, m, err
	}
	ct, err := ts.Encrypt(key)
	if err != nil {
		return pw, m, err
	}

	// Encode the master key value to base64url
	m.Value = base64url.NewValue(ct)

	return pw, m, nil
}
