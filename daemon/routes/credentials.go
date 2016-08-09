package routes

// This file contains routes related to credentials/secrets

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/ed25519"

	"github.com/arigatomachine/cli/daemon/base64"
	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/envelope"
	"github.com/arigatomachine/cli/daemon/identity"
	"github.com/arigatomachine/cli/daemon/primitive"
	"github.com/arigatomachine/cli/daemon/registry"
	"github.com/arigatomachine/cli/daemon/session"
)

// plainTextCredential is an unencrypted credential. This is what we expect to
// send and recieve to/from the CLI.
type plainTextCredential struct {
	Name      string       `json:"name"`
	OrgID     *identity.ID `json:"org_id"`
	PathExp   string       `json:"pathexp"`
	ProjectID *identity.ID `json:"project_id"`
	Value     string       `json:"value"`
}

// plainEnvelope holds our plainTextCredential
type plainTextEnvelope struct {
	ID      *identity.ID         `json:"id"`
	Version uint8                `json:"version"`
	Body    *plainTextCredential `json:"body"`
}

func credentialsGetRoute(client *registry.Client,
	s session.Session) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		trees, err := client.CredentialTree.List(
			q.Get("Name"), q.Get("path"), q.Get("pathexp"), s.ID())
		if err != nil {
			log.Printf("error retrieving credential trees: %s", err)
			encodeResponseErr(w, err)
			return
		}

		// Loop over the trees and unpack the credentials; later on we will
		// actually do real work and decrypt each of these credentials but for
		// now we just need ot return a list of them!
		creds := []envelope.Unsigned{}
		for _, tree := range trees {
			creds = append(creds, tree.Credentials...)
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(creds)
		if err != nil {
			log.Printf("error encoding credentials: %s", err)
			encodeResponseErr(w, err)
			return
		}
	}
}

func credentialsPostRoute(client *registry.Client,
	s session.Session, db *db.DB, engine *crypto.Engine) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		cred := plainTextEnvelope{}

		err := dec.Decode(&cred)
		if err != nil {
			log.Printf("error decoding credential: %s", err)
			encodeResponseErr(w, err)
			return
		}

		// Ensure we have an existing keyring for this credential's pathexp
		trees, err := client.CredentialTree.List(cred.Body.Name, "",
			cred.Body.PathExp, s.ID())
		if err != nil {
			log.Printf("error retrieving credential trees: %s", err)
			encodeResponseErr(w, err)
			return
		}

		sigID, encID, kp, err := fetchKeyPairs(client, cred.Body.OrgID)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		// No matching CredentialTree/KeyRing for this credential.
		// We'll make a new one now.
		if len(trees) == 0 {
			tree, err := createCredentialTree(cred.Body, sigID, encID, kp,
				client, engine)
			if err != nil {
				log.Printf("error creating credential tree: %s", err)
				encodeResponseErr(w, err)
				return
			}
			trees = []registry.CredentialTree{*tree}
		}

		tree := trees[0]
		creds := tree.Credentials

		// Construct an encrypted and signed version of the credential
		credBody := primitive.Credential{
			Name:      cred.Body.Name,
			PathExp:   cred.Body.PathExp,
			ProjectID: cred.Body.ProjectID,
			OrgID:     cred.Body.OrgID,
			Credential: &primitive.CredentialValue{
				Algorithm: crypto.SecretBox,
			},
		}

		if len(creds) == 0 {
			log.Printf("no previous")
			credBody.Previous = nil
			credBody.CredentialVersion = 1
		} else {
			previousCred := creds[len(creds)-1]
			previousCredBody := previousCred.Body.(*primitive.Credential)

			if previousCredBody.Name != credBody.Name ||
				previousCredBody.PathExp != credBody.PathExp {

				err = fmt.Errorf("Non-matching credential returned in tree")
				log.Printf("Error finding previous credential version: %s", err)
				encodeResponseErr(w, err)
				return
			}

			credBody.Previous = previousCred.ID
			credBody.CredentialVersion = previousCredBody.CredentialVersion + 1
		}

		myID := s.ID()

		// Find our keyring membership entry, so we can access the keyring
		// master encryption key.
		var krm *primitive.KeyringMember
		for _, m := range tree.Members {
			mBody := m.Body.(*primitive.KeyringMember)
			if mBody.OwnerID == myID {
				krm = mBody
				break
			}
		}

		if krm == nil {
			err = fmt.Errorf("No keyring membership found")
			log.Printf("Error finding keyring membership")
			encodeResponseErr(w, err)
			return
		}

		// Lookup the key that encrypted the mek for us, so we can decrypt it.
		claimTrees, err := client.ClaimTree.List(credBody.OrgID, nil)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		if len(claimTrees) != 1 {
			err = fmt.Errorf("No claim tree found for org: %s", credBody.OrgID)
			log.Printf("Error looking up claim tree: %s", err)
			encodeResponseErr(w, err)
			return
		}

		var encryptingKey *primitive.PublicKey
		for _, segment := range claimTrees[0].PublicKeys {
			if segment.Key.ID == krm.EncryptingKeyID {
				encryptingKey = segment.Key.Body.(*primitive.PublicKey)
				break
			}
		}
		if encryptingKey == nil {
			err = fmt.Errorf("Couldn't find encrypting key %s", krm.EncryptingKeyID)
			log.Printf("Error finding keyring membership: %s", err)
			encodeResponseErr(w, err)
			return
		}

		// Derive a key for the credential using the keyring master key
		// and use the derived key to encrypt the credential
		cekNonce, ctNonce, ct, err := engine.BoxCredential(
			[]byte(cred.Body.Value), *krm.Key.Value, *krm.Key.Nonce,
			&kp.Encryption, *encryptingKey.Key.Value)

		credBody.Nonce = base64.NewValue(cekNonce)

		credBody.Credential.Nonce = base64.NewValue(ctNonce)
		credBody.Credential.Value = base64.NewValue(ct)

		signed, err := engine.SignedEnvelope(&credBody, sigID, &kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		_, err = client.Credentials.Create(signed)
		if err != nil {
			log.Printf("error creating credential: %s", err)
			encodeResponseErr(w, err)
			return
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(cred)
		if err != nil {
			log.Printf("error encoding credential create resp: %s", err)
			encodeResponseErr(w, err)
			return
		}
	}
}

// createCredentialTree generates, signs, and posts a new CredentialTree
// to the registry.
func createCredentialTree(credBody *plainTextCredential, sigID *identity.ID,
	encID *identity.ID, kp *crypto.KeyPairs, client *registry.Client,
	engine *crypto.Engine) (*registry.CredentialTree, error) {

	parts := strings.Split(credBody.PathExp, "/")
	if len(parts) != 7 { // first part is empty
		return nil, fmt.Errorf("Invalid path expression: %s", credBody.PathExp)
	}

	pathExp := strings.Join(parts[:6], "/") + "/*"

	keyring, err := engine.SignedEnvelope(
		&primitive.Keyring{
			Created:   time.Now().UTC(),
			OrgID:     credBody.OrgID,
			PathExp:   pathExp,
			ProjectID: credBody.ProjectID,

			// This is the first instance of the keyring, so version is 1,
			// and there is no previous instance.
			Previous:       nil,
			KeyringVersion: 1,
		},
		sigID, &kp.Signature)
	if err != nil {
		return nil, err
	}

	// XXX: sensitive value. protect with OS things.
	mek := make([]byte, 256)
	_, err = rand.Read(mek)
	if err != nil {
		return nil, err
	}

	// get users in the members group of this org.
	// use their public key to encrypt the mek with a random nonce.
	// XXX: we need to filter this down
	claimTrees, err := client.ClaimTree.List(credBody.OrgID, nil)
	if err != nil {
		return nil, err
	}

	if len(claimTrees) != 1 {
		return nil, fmt.Errorf("No claim tree found for org: %s", credBody.OrgID)
	}

	members := []envelope.Signed{}
	for _, segment := range claimTrees[0].PublicKeys {
		// For user in members group, generate membership object
		pubKey := segment.Key.Body.(*primitive.PublicKey)

		if pubKey.KeyType != encryptionKeyType {
			continue
		}

		encmek, nonce, err := engine.Box(mek, &kp.Encryption, []byte(*pubKey.Key.Value))
		if err != nil {
			return nil, err
		}

		member, err := engine.SignedEnvelope(
			&primitive.KeyringMember{
				Created:         time.Now().UTC(),
				OrgID:           credBody.OrgID,
				ProjectID:       credBody.ProjectID,
				KeyringID:       keyring.ID,
				OwnerID:         pubKey.OwnerID,
				PublicKeyID:     segment.Key.ID,
				EncryptingKeyID: encID,

				Key: &primitive.KeyringMemberKey{
					Algorithm: crypto.EasyBox,
					Nonce:     base64.NewValue(nonce),
					Value:     base64.NewValue(encmek),
				},
			},
			sigID, &kp.Signature)
		if err != nil {
			return nil, err
		}

		members = append(members, *member)
	}

	return client.CredentialTree.Post(&registry.CredentialTree{
		Keyring:     keyring,
		Members:     members,
		Credentials: []envelope.Unsigned{},
	})
}

// fetchKeyPairs fetches the user's signing and encryption keypairs from the
// registry for the given org id.
func fetchKeyPairs(client *registry.Client, orgID *identity.ID) (*identity.ID,
	*identity.ID, *crypto.KeyPairs, error) {

	keyPairs, err := client.KeyPairs.List(orgID)
	if err != nil {
		return nil, nil, nil, err
	}

	var sigClaimed registry.ClaimedKeyPair
	var encClaimed registry.ClaimedKeyPair
	for _, keyPair := range keyPairs {
		pubKey := keyPair.PublicKey.Body.(*primitive.PublicKey)
		switch pubKey.KeyType {
		case signingKeyType:
			sigClaimed = keyPair
		case encryptionKeyType:
			encClaimed = keyPair
		default:
			err = fmt.Errorf("Unknown key type: %s", pubKey.KeyType)
			return nil, nil, nil, err
		}
	}

	if sigClaimed.PublicKey == nil || encClaimed.PublicKey == nil {
		err = fmt.Errorf("Missing encryption or signing keypairs")
		return nil, nil, nil, err
	}

	sigPub := sigClaimed.PublicKey.Body.(*primitive.PublicKey).Key.Value
	sigKP := crypto.SignatureKeyPair{
		Public:  ed25519.PublicKey(*sigPub),
		Private: *sigClaimed.PrivateKey.Body.(*primitive.PrivateKey).Key.Value,
		PNonce:  *sigClaimed.PrivateKey.Body.(*primitive.PrivateKey).PNonce,
	}

	encPub := *encClaimed.PublicKey.Body.(*primitive.PublicKey).Key.Value
	encPubB := [32]byte{}
	copy(encPubB[:], encPub)
	encKP := crypto.EncryptionKeyPair{
		Public:  encPubB,
		Private: *encClaimed.PrivateKey.Body.(*primitive.PrivateKey).Key.Value,
		PNonce:  *encClaimed.PrivateKey.Body.(*primitive.PrivateKey).PNonce,
	}

	kp := crypto.KeyPairs{
		Signature:  sigKP,
		Encryption: encKP,
	}

	return sigClaimed.PublicKey.ID, encClaimed.PublicKey.ID, &kp, nil
}
