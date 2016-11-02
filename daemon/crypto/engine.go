// Package crypto provides access to secure encryption and signing methods
package crypto

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/dchest/blake2b"
	"github.com/keybase/go-triplesec"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"

	"github.com/manifoldco/torus-cli/daemon/ctxutil"
	"github.com/manifoldco/torus-cli/daemon/session"
)

const (
	nonceSize = 16
	blakeSize = 16
)

// SignatureKeyPair is an ed25519/eddsa digital signature keypair.
// The private portion of the keypair is encrypted with triplesec.
//
// PNonce contains the nonce used when deriving the password used to encrypt
// the private portion.
type SignatureKeyPair struct {
	Public  ed25519.PublicKey
	Private []byte
	PNonce  []byte
}

// EncryptionKeyPair is a curve25519 encryption keypair.
// The private portion of the keypair is encrypted with triplesec.
//
// PNonce contains the nonce used when deriving the password used to encrypt
// the private portion.
type EncryptionKeyPair struct {
	Public  [32]byte
	Private []byte
	PNonce  []byte
}

// KeyPairs contains a signature and an encryption keypair for a user.
type KeyPairs struct {
	Signature  SignatureKeyPair
	Encryption EncryptionKeyPair
}

// Engine exposes methods to encrypt, unencrypt and sign values, using
// the logged in user's credentials.
type Engine struct {
	sess session.Session
}

// NewEngine returns a new Engine
func NewEngine(sess session.Session) *Engine {
	return &Engine{sess: sess}
}

// Seal encrypts the plaintext pt bytes with triplesec-v3 using a key derived
// via blake2b from the user's master key and a nonce (returned).
func (e *Engine) Seal(ctx context.Context, pt []byte) ([]byte, []byte, error) {
	mk, err := e.unsealMasterKey(ctx)
	if err != nil {
		return nil, nil, err
	}

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, nonceSize)
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, nil, err
	}

	dk, err := deriveKey(ctx, mk, nonce, blakeSize)
	if err != nil {
		return nil, nil, err
	}

	ts, err := newTriplesec(ctx, dk)
	if err != nil {
		return nil, nil, err
	}

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, nil, err
	}

	ct, err := ts.Encrypt(pt)

	return ct, nonce, err
}

// Unseal decrypts the ciphertext ct, encrypted with triplesec-v3, using the
// a key derived via blake2b from the user's master key and the provided nonce.
func (e *Engine) Unseal(ctx context.Context, ct, nonce []byte) ([]byte, error) {
	mk, err := e.unsealMasterKey(ctx)
	if err != nil {
		return nil, err
	}

	dk, err := deriveKey(ctx, mk, nonce, blakeSize)
	if err != nil {
		return nil, err
	}

	ts, err := newTriplesec(ctx, dk)
	if err != nil {
		return nil, err
	}

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	return ts.Decrypt(ct)
}

// Box encrypts the plaintext pt bytes with Box, using the private key found in
// privKP, first decrypted with the user's master key, and encrypted for the
// public key pubKey.
//
// It returns the ciphertext, the nonce used for encrypting the plaintext,
// and an optional error.
func (e *Engine) Box(ctx context.Context, pt []byte, privKP *EncryptionKeyPair,
	pubKey []byte) ([]byte, []byte, error) {

	err := ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, nil, err
	}

	nonce := [24]byte{}
	_, err = rand.Read(nonce[:])
	if err != nil {
		return nil, nil, err
	}

	privKey, err := e.Unseal(ctx, privKP.Private, privKP.PNonce)
	if err != nil {
		return nil, nil, err
	}

	privkb := [32]byte{}
	copy(privkb[:], privKey)

	pubkb := [32]byte{}
	copy(pubkb[:], pubKey)

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, nil, err
	}

	return box.Seal([]byte{}, pt, &nonce, &pubkb, &privkb), nonce[:], nil
}

// Unbox Decrypts and verifies ciphertext ct that was previously encrypted using
// the provided nonce, and the inverse parts of the provided keypairs.
func (e *Engine) Unbox(ctx context.Context, ct, nonce []byte,
	privKP *EncryptionKeyPair, pubKey []byte) ([]byte, error) {

	privKey, err := e.Unseal(ctx, privKP.Private, privKP.PNonce)
	if err != nil {
		return nil, err
	}

	nonceb := [24]byte{}
	copy(nonceb[:], nonce)

	privkb := [32]byte{}
	copy(privkb[:], privKey)

	pubkb := [32]byte{}
	copy(pubkb[:], pubKey)

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	pt, success := box.Open([]byte{}, ct, &nonceb, &pubkb, &privkb)
	if !success {
		return nil, errors.New("Failed to decrypt ciphertext")
	}

	return pt, nil
}

// BoxCredential encrypts the credential value pt via symmetric secretbox
// encryption.
//
// Doing so is a multistep process.
// First we use the user's session data to unseal their private encryption key.
// With their encryption key and the public encryption key provided, we can
// decrypt the keyring master key (mek).
// Using mek and a generated nonce, we derive the credential encryption key
// (cek) via blake2b.
// Finally, we use the cek and a generated nonce to encrypt the credential.
//
// BoxCredential returns the nonce generated to derive the credential
// encryption key,  the nonce generated for encrypting the credential, and the
// encrypted credential.
func (e *Engine) BoxCredential(ctx context.Context, pt, encMec, mecNonce []byte,
	privKP *EncryptionKeyPair, pubKey []byte) ([]byte, []byte, []byte, error) {

	err := ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	nonces := make([]byte, 48)
	_, err = rand.Read(nonces)
	if err != nil {
		return nil, nil, nil, err
	}

	cekNonce := nonces[:24]
	nonce := [24]byte{}
	copy(nonce[:], nonces[24:])

	mek, err := e.Unbox(ctx, encMec, mecNonce, privKP, pubKey)
	if err != nil {
		return nil, nil, nil, err
	}

	cek, err := deriveKey(ctx, mek, cekNonce, 32)
	if err != nil {
		return nil, nil, nil, err
	}

	cekb := [32]byte{}
	copy(cekb[:], cek)

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	ct := secretbox.Seal([]byte{}, pt, &nonce, &cekb)
	return cekNonce, nonce[:], ct, err
}

// UnboxCredential does the inverse of BoxCredential to retrieve the plaintext
// version of a credential.
func (e *Engine) UnboxCredential(ctx context.Context, ct, encMec, mecNonce,
	cekNonce, ctNonce []byte, privKP *EncryptionKeyPair, pubKey []byte) ([]byte, error) {

	mek, err := e.Unbox(ctx, encMec, mecNonce, privKP, pubKey)
	if err != nil {
		return nil, err
	}

	cek, err := deriveKey(ctx, mek, cekNonce, 32)
	if err != nil {
		return nil, err
	}

	cekb := [32]byte{}
	copy(cekb[:], cek)

	ctNonceb := [24]byte{}
	copy(ctNonceb[:], ctNonce)

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	pt, success := secretbox.Open([]byte{}, ct, &ctNonceb, &cekb)
	if !success {
		return nil, errors.New("Failed to decrypt ciphertext")
	}

	return pt, nil
}

// Unboxer provides an interface to unbox credentials, within the context
type Unboxer interface {
	Unbox(context.Context, []byte, []byte, []byte) ([]byte, error)
}

type unboxerImpl struct {
	mek []byte
}

func (u *unboxerImpl) Unbox(ctx context.Context, ct, cekNonce, ctNonce []byte) ([]byte, error) {
	cek, err := deriveKey(ctx, u.mek, cekNonce, 32)
	if err != nil {
		return nil, err
	}

	cekb := [32]byte{}
	copy(cekb[:], cek)

	ctNonceb := [24]byte{}
	copy(ctNonceb[:], ctNonce)

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	pt, success := secretbox.Open([]byte{}, ct, &ctNonceb, &cekb)
	if !success {
		return nil, errors.New("Failed to decrypt ciphertext")
	}

	return pt, nil
}

// WithUnboxer returns an Unboxer for unboxing credentials within the context
// of the provided keypairs.
func (e *Engine) WithUnboxer(ctx context.Context, encMec, mecNonce []byte,
	privKP *EncryptionKeyPair, pubKey []byte, fn func(Unboxer) error) error {

	mek, err := e.Unbox(ctx, encMec, mecNonce, privKP, pubKey)
	if err != nil {
		return err
	}

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return err
	}

	u := unboxerImpl{mek: mek}

	return fn(&u)
}

// CloneMembership decrypts the given KeyringMember object, and creates another
// for the targeted user.
func (e *Engine) CloneMembership(ctx context.Context, encMec, mecNonce []byte, privKP *EncryptionKeyPair, encPubKey, targetPubKey []byte) ([]byte, []byte, error) {
	mek, err := e.Unbox(ctx, encMec, mecNonce, privKP, encPubKey)
	if err != nil {
		return nil, nil, err
	}

	return e.Box(ctx, mek, privKP, targetPubKey)
}

// GenerateKeyPairs generates and ed25519 signing key pair, and a curve25519
// encryption key pair for the user, encrypting the private keys in
// triplesec-v3 with the user's master key.
func (e *Engine) GenerateKeyPairs(ctx context.Context) (*KeyPairs, error) {
	kp := &KeyPairs{}

	err := ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	pubSig, privSig, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	sealedSig, nonceSig, err := e.Seal(ctx, privSig)
	if err != nil {
		return nil, err
	}

	kp.Signature.Private = sealedSig
	kp.Signature.Public = pubSig
	kp.Signature.PNonce = nonceSig

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	pubEnc, privEnc, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	sealedEnc, nonceEnc, err := e.Seal(ctx, (*privEnc)[:])
	if err != nil {
		return nil, err
	}

	kp.Encryption.Private = sealedEnc
	kp.Encryption.Public = *pubEnc
	kp.Encryption.PNonce = nonceEnc

	return kp, nil
}

// Sign signs b bytes using the provided Sealed ed25519 keypair.
func (e *Engine) Sign(ctx context.Context, s SignatureKeyPair, b []byte) ([]byte, error) {
	pk, err := e.Unseal(ctx, s.Private, s.PNonce)
	if err != nil {
		return nil, err
	}

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	return ed25519.Sign(pk, b), nil
}

// Verify verifies that sig is the correct signature for b given
// SignatureKeyPair s.
func (e *Engine) Verify(ctx context.Context, s SignatureKeyPair, b, sig []byte) (bool, error) {
	err := ctxutil.ErrIfDone(ctx)
	if err != nil {
		return false, err
	}

	return ed25519.Verify(s.Public, b, sig), nil
}

// SignedEnvelope returns a new SignedEnvelope containing body
func (e *Engine) SignedEnvelope(ctx context.Context, body identity.Immutable,
	sigID *identity.ID, sigKP *SignatureKeyPair) (*envelope.Signed, error) {

	err := ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}

	s, err := e.Sign(ctx, *sigKP, append([]byte(strconv.Itoa(body.Version())), b...))
	if err != nil {
		return nil, err
	}

	sig := primitive.Signature{
		PublicKeyID: sigID,
		Algorithm:   EdDSA,
		Value:       base64.NewValue(s),
	}

	id, err := identity.NewImmutable(body, &sig)
	if err != nil {
		return nil, err
	}

	return &envelope.Signed{
		ID:        &id,
		Version:   uint8(body.Version()),
		Body:      body,
		Signature: sig,
	}, nil
}

// unsealMasterKey uses the scrypt stretched password to decrypt the master
// password, which is encrypted with triplesec-v3
func (e *Engine) unsealMasterKey(ctx context.Context) ([]byte, error) {
	ts, err := newTriplesec(ctx, []byte(e.sess.Passphrase()))
	if err != nil {
		return nil, err
	}

	masterKey, err := e.sess.MasterKey()
	if err != nil {
		return nil, err
	}

	err = ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	mk, err := ts.Decrypt(*masterKey)
	return mk, err
}

func newTriplesec(ctx context.Context, k []byte) (*triplesec.Cipher, error) {
	err := ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	ts, err := triplesec.NewCipher(k, nil)
	if err != nil {
		return nil, err
	}

	return ts, nil
}

// deriveKey Derives a single use key from the given master key via blake2b
// and a nonce.
func deriveKey(ctx context.Context, mk, nonce []byte, size uint8) ([]byte, error) {
	err := ctxutil.ErrIfDone(ctx)
	if err != nil {
		return nil, err
	}

	h := blake2b.NewMAC(size, nonce) // NewMAC can panic if size is too big.
	h.Sum(mk)
	return h.Sum(nil), nil
}
