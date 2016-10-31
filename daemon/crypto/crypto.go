package crypto

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/scrypt"

	base64url "github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/daemon/ctxutil"
	"github.com/manifoldco/torus-cli/primitive"
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
	keyLen         = 256
	saltBytes      = 16
	masterKeyBytes = 256
)

// LoginKeypair represents an Ed25519 Keypair used for generating a login token
// signature for Passphrase-Dervied Public Key Authentication.
type LoginKeypair struct {
	public  ed25519.PublicKey
	private ed25519.PrivateKey
	salt    *base64url.Value
}

// PublicKey returns the base64 value of the public key
func (k *LoginKeypair) PublicKey() *base64url.Value {
	return base64url.NewValue(k.public)
}

// Salt returns the base64 representation of the salt used to derive the
// LoginKeypair
func (k *LoginKeypair) Salt() *base64url.Value {
	return k.salt
}

// Sign returns a signature of the given token as a base64 string
func (k *LoginKeypair) Sign(token []byte) *base64url.Value {
	sig := ed25519.Sign(k.private, token)
	return base64url.NewValue(sig)
}

// DeriveLoginHMAC HMACs the provided token with a key derived from password
// and the provided base64 encoded salt.
func DeriveLoginHMAC(ctx context.Context, password []byte, salt, token string) (string, error) {
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

func deriveHash(ctx context.Context, password []byte, salt string) ([]byte, error) {
	s := make([]byte, base64.RawURLEncoding.DecodedLen(len(salt)))
	l, err := base64.RawURLEncoding.Decode(s, []byte(salt))
	if err != nil {
		return nil, err
	}

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	key, err := scrypt.Key(password, s[:l], n, r, p, keyLen)
	return key, err
}

func derivePassword(ctx context.Context, password []byte, salt string) ([]byte, error) {
	key, err := deriveHash(ctx, password, salt)
	if err != nil {
		return nil, err
	}

	pwh := key[192:224]
	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	return pwh, err
}

// GenerateSalt returns a 16-byte (128 bit) salt used in password and secret
// key derivation.
func GenerateSalt(ctx context.Context) (*base64url.Value, error) {
	err := ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	salt := make([]byte, saltBytes) // 16
	_, err = rand.Read(salt)
	if err != nil {
		return nil, err
	}

	return base64url.NewValue(salt), nil
}

// EncryptPasswordObject derives the master key and password hash from password
// and salt, returning the master and password objects
func EncryptPasswordObject(ctx context.Context, password string) (*primitive.UserPassword, *primitive.MasterKey, error) {
	pw := &primitive.UserPassword{
		Alg: Scrypt,
	}

	// Generate 128 bit (16 byte) salt for password
	salt := make([]byte, saltBytes) // 16
	_, err := rand.Read(salt)
	if err != nil {
		return nil, nil, err
	}

	// Encode salt bytes to base64url
	pw.Salt = base64.RawURLEncoding.EncodeToString(salt)

	// Create password hash bytes
	pwh, err := derivePassword(ctx, []byte(password), pw.Salt)
	if err != nil {
		return nil, nil, err
	}

	// Encode password value to base64url
	pw.Value = base64url.NewValue(pwh)

	m, err := CreateMasterKeyObject(ctx, []byte(password))
	if err != nil {
		return nil, nil, err
	}

	return pw, m, nil
}

// CreateMasterKeyObject generates a 256 byte master key which is then
// encrypted using TripleSec-v3 using the given password.
func CreateMasterKeyObject(ctx context.Context, password []byte) (*primitive.MasterKey, error) {
	m := &primitive.MasterKey{
		Alg: Triplesec,
	}

	// Generate a master key of 256 bytes
	key := make([]byte, masterKeyBytes)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	ts, err := newTriplesec(ctx, password)
	if err != nil {
		return nil, err
	}

	ct, err := ts.Encrypt(key)
	if err != nil {
		return nil, err
	}

	// Encode the master key value to base64url
	m.Value = base64url.NewValue(ct)
	return m, nil
}

// DeriveLoginKeypair dervies the ed25519 login keypair used for machine
// authentication from the given salt and secret values.
func DeriveLoginKeypair(ctx context.Context, secret, salt *base64url.Value) (
	*LoginKeypair, error) {

	key, err := deriveHash(ctx, *secret, salt.String())
	if err != nil {
		return nil, err
	}

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(key[224:]) // Use last 32 bytes of 256 to derive key
	pubKey, privKey, err := ed25519.GenerateKey(r)
	if err != nil {
		return nil, err
	}

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	keypair := &LoginKeypair{
		public:  pubKey,
		private: privKey,
		salt:    salt,
	}
	return keypair, nil
}
