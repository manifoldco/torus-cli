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
	"github.com/arigatomachine/cli/daemon/primitive"
	"github.com/arigatomachine/cli/daemon/registry"
	"github.com/arigatomachine/cli/daemon/session"
)

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
		cred := &envelope.Unsigned{}

		err := dec.Decode(&cred)
		if err != nil {
			log.Printf("error decoding credential: %s", err)
			encodeResponseErr(w, err)
			return
		}

		credBody := cred.Body.(*primitive.Credential)

		// Ensure we have an existing keyring for this credential's pathexp
		trees, err := client.CredentialTree.List(credBody.Name, "", credBody.PathExp,
			s.ID())
		if err != nil {
			log.Printf("error retrieving credential trees: %s", err)
			encodeResponseErr(w, err)
			return
		}

		// No matching CredentialTree/KeyRing for this credential.
		// We'll make a new one now.
		if len(trees) == 0 {
			tree, err := createCredentialTree(credBody, client, s, db, engine)
			if err != nil {
				log.Printf("error creating credential tree: %s", err)
				encodeResponseErr(w, err)
				return
			}
			trees = []registry.CredentialTree{*tree}
		}

		tree := trees[0]
		creds := tree.Credentials

		if len(creds) == 0 {
			log.Printf("no previous")
			credBody.Previous = nil
			credBody.CredentialVersion = 1
		} else {
			previousCred := creds[len(creds)-1]
			previousCredBody := previousCred.Body.(*primitive.Credential)

			if previousCredBody.Name != credBody.Name || previousCredBody.PathExp != credBody.PathExp {
				err = fmt.Errorf("Non-matching credential returned in tree")
				log.Printf("Error finding previous credential version: %s", err)
				encodeResponseErr(w, err)
				return
			}

			credBody.Previous = previousCred.ID
			credBody.CredentialVersion = previousCredBody.CredentialVersion + 1
		}

		cred, err = client.Credentials.Create(cred)
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
func createCredentialTree(credBody *primitive.Credential,
	client *registry.Client, s session.Session, db *db.DB,
	engine *crypto.Engine) (*registry.CredentialTree, error) {

	keyPairs, err := client.KeyPairs.List(credBody.OrgID)
	if err != nil {
		return nil, err
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
			return nil, fmt.Errorf("Unknown key type: %s", pubKey.KeyType)
		}
	}

	if sigClaimed.PublicKey == nil || encClaimed.PublicKey == nil {
		return nil, fmt.Errorf("Missing encryption or signing keypairs")
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
		sigClaimed.PublicKey.ID, &sigKP)
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

		encmek, nonce, err := engine.Box(mek, &encKP, []byte(*pubKey.Key.Value))
		if err != nil {
			return nil, err
		}

		noncev := base64.Value(nonce)
		encmekv := base64.Value(encmek)

		member, err := engine.SignedEnvelope(
			&primitive.KeyringMember{
				Created:         time.Now().UTC(),
				OrgID:           credBody.OrgID,
				ProjectID:       credBody.ProjectID,
				KeyringID:       keyring.ID,
				OwnerID:         pubKey.OwnerID,
				PublicKeyID:     segment.Key.ID,
				EncryptingKeyID: encClaimed.PublicKey.ID,

				Key: &primitive.KeyringMemberKey{
					Algorithm: crypto.EasyBox,
					Nonce:     &noncev,
					Value:     &encmekv,
				},
			},
			sigClaimed.PublicKey.ID, &sigKP)
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
