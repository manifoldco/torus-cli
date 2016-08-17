package routes

// This file contains routes related to credentials/secrets

import (
	"context"
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
	s session.Session, engine *crypto.Engine) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		q := r.URL.Query()
		trees, err := client.CredentialTree.List(ctx, q.Get("Name"),
			q.Get("path"), q.Get("pathexp"), s.ID())
		if err != nil {
			log.Printf("error retrieving credential trees: %s", err)
			encodeResponseErr(w, err)
			return
		}

		// Loop over the trees and unpack the credentials; later on we will
		// actually do real work and decrypt each of these credentials but for
		// now we just need ot return a list of them!
		creds := []plainTextEnvelope{}
		for _, tree := range trees {
			orgID := tree.Keyring.Body.(*primitive.Keyring).OrgID
			_, _, kp, err := fetchKeyPairs(ctx, client, orgID)
			if err != nil {
				encodeResponseErr(w, err)
				return
			}

			krm, err := findKeyringMember(s.ID(), &tree)
			if err != nil {
				log.Printf("Error finding keyring membership")
				encodeResponseErr(w, err)
				return
			}

			encryptingKey, err := findEncryptingKey(ctx, client, orgID,
				krm.EncryptingKeyID)
			if err != nil {
				log.Printf("Error finding encrypting key: %s", err)
				encodeResponseErr(w, err)
				return
			}

			for _, cred := range tree.Credentials {
				credBody := cred.Body.(*primitive.Credential)
				pt, err := engine.UnboxCredential(ctx,
					*credBody.Credential.Value, *krm.Key.Value, *krm.Key.Nonce,
					*credBody.Nonce, *credBody.Credential.Nonce, &kp.Encryption,
					*encryptingKey.Key.Value)
				if err != nil {
					log.Printf("Error decrypting credential: %s", err)
					encodeResponseErr(w, err)
					return
				}

				plainCred := plainTextEnvelope{
					ID:      cred.ID,
					Version: cred.Version,
					Body: &plainTextCredential{
						Name:      credBody.Name,
						PathExp:   credBody.PathExp,
						ProjectID: credBody.ProjectID,
						OrgID:     credBody.OrgID,
						Value:     string(pt),
					},
				}
				creds = append(creds, plainCred)
			}
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

func credentialsPostRoute(client *registry.Client, s session.Session,
	engine *crypto.Engine) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dec := json.NewDecoder(r.Body)
		cred := plainTextEnvelope{}

		err := dec.Decode(&cred)
		if err != nil {
			log.Printf("error decoding credential: %s", err)
			encodeResponseErr(w, err)
			return
		}

		// Ensure we have an existing keyring for this credential's pathexp
		trees, err := client.CredentialTree.List(ctx, cred.Body.Name, "",
			cred.Body.PathExp, s.ID())
		if err != nil {
			log.Printf("error retrieving credential trees: %s", err)
			encodeResponseErr(w, err)
			return
		}

		sigID, encID, kp, err := fetchKeyPairs(ctx, client, cred.Body.OrgID)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		newTree := false
		// No matching CredentialTree/KeyRing for this credential.
		// We'll make a new one now.
		if len(trees) == 0 {
			tree, err := createCredentialTree(ctx, cred.Body, sigID, encID, kp,
				client, engine)
			if err != nil {
				log.Printf("error creating credential tree: %s", err)
				encodeResponseErr(w, err)
				return
			}
			trees = []registry.CredentialTree{*tree}
			newTree = true
		}

		tree := trees[0]
		creds := tree.Credentials

		// Construct an encrypted and signed version of the credential
		credBody := primitive.Credential{
			Name:      cred.Body.Name,
			PathExp:   cred.Body.PathExp,
			KeyringID: tree.Keyring.ID,
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

		// Find our keyring membership entry, so we can access the keyring
		// master encryption key.
		krm, err := findKeyringMember(s.ID(), &tree)
		if err != nil {
			log.Printf("Error finding keyring membership")
			encodeResponseErr(w, err)
			return
		}

		encryptingKey, err := findEncryptingKey(ctx, client, credBody.OrgID,
			krm.EncryptingKeyID)
		if err != nil {
			log.Printf("Error finding encrypting key: %s", err)
			encodeResponseErr(w, err)
			return
		}

		// Derive a key for the credential using the keyring master key
		// and use the derived key to encrypt the credential
		cekNonce, ctNonce, ct, err := engine.BoxCredential(
			ctx, []byte(cred.Body.Value), *krm.Key.Value, *krm.Key.Nonce,
			&kp.Encryption, *encryptingKey.Key.Value)

		credBody.Nonce = base64.NewValue(cekNonce)

		credBody.Credential.Nonce = base64.NewValue(ctNonce)
		credBody.Credential.Value = base64.NewValue(ct)

		signed, err := engine.SignedEnvelope(ctx, &credBody, sigID, &kp.Signature)
		if err != nil {
			encodeResponseErr(w, err)
			return
		}

		if newTree {
			tree.Credentials = []envelope.Signed{*signed}
			_, err = client.CredentialTree.Post(ctx, &tree)
		} else {
			_, err = client.Credentials.Create(ctx, signed)
		}
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
func createCredentialTree(ctx context.Context, credBody *plainTextCredential,
	sigID *identity.ID, encID *identity.ID, kp *crypto.KeyPairs,
	client *registry.Client, engine *crypto.Engine) (*registry.CredentialTree, error) {

	parts := strings.Split(credBody.PathExp, "/")
	if len(parts) != 7 { // first part is empty
		return nil, fmt.Errorf("Invalid path expression: %s", credBody.PathExp)
	}

	pathExp := strings.Join(parts[:6], "/") + "/*"

	keyring, err := engine.SignedEnvelope(
		ctx, &primitive.Keyring{
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
	mek := make([]byte, 64)
	_, err = rand.Read(mek)
	if err != nil {
		return nil, err
	}

	teams, err := client.Teams.List(ctx, credBody.OrgID)
	if err != nil {
		return nil, err
	}

	team, err := findMembersTeam(teams)
	if err != nil {
		return nil, err
	}

	memberships, err := client.Memberships.List(ctx, credBody.OrgID, team.ID, nil)
	if err != nil {
		return nil, err
	}

	claimTrees, err := client.ClaimTree.List(ctx, credBody.OrgID, nil)
	if err != nil {
		return nil, err
	}

	if len(claimTrees) != 1 {
		return nil, fmt.Errorf("No claim tree found for org: %s", credBody.OrgID)
	}

	// get users in the members group of this org.
	// use their public key to encrypt the mek with a random nonce.
	// XXX: we need to filter this down
	members := []envelope.Signed{}
	for _, membership := range memberships {
		mBody := membership.Body.(*primitive.Membership)

		// For this user, find their public encryption key
		encPubKey, err := findEncryptionPublicKey(claimTrees, credBody.OrgID, mBody.OwnerID)
		if err != nil {
			return nil, err
		}

		pubKey := encPubKey.Body.(*primitive.PublicKey)
		encmek, nonce, err := engine.Box(ctx, mek, &kp.Encryption, []byte(*pubKey.Key.Value))
		if err != nil {
			return nil, err
		}

		member, err := engine.SignedEnvelope(
			ctx, &primitive.KeyringMember{
				Created:         time.Now().UTC(),
				OrgID:           credBody.OrgID,
				ProjectID:       credBody.ProjectID,
				KeyringID:       keyring.ID,
				OwnerID:         pubKey.OwnerID,
				PublicKeyID:     encPubKey.ID,
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

	tree := registry.CredentialTree{
		Keyring: keyring,
		Members: members,
	}

	return &tree, nil
}

// fetchKeyPairs fetches the user's signing and encryption keypairs from the
// registry for the given org id.
func fetchKeyPairs(ctx context.Context, client *registry.Client,
	orgID *identity.ID) (*identity.ID, *identity.ID, *crypto.KeyPairs, error) {

	keyPairs, err := client.KeyPairs.List(ctx, orgID)
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

// findKeyringMember finds the keyring member with the given id
func findKeyringMember(id *identity.ID,
	tree *registry.CredentialTree) (*primitive.KeyringMember, error) {
	var krm *primitive.KeyringMember
	for _, m := range tree.Members {
		mBody := m.Body.(*primitive.KeyringMember)
		if *mBody.OwnerID == *id {
			krm = mBody
			break
		}
	}

	if krm == nil {
		err := fmt.Errorf("No keyring membership found")
		return nil, err
	}

	return krm, nil
}

// findEncryptingKey queries the registry for public keys in the given org, to
// find the matching one
func findEncryptingKey(ctx context.Context, client *registry.Client, orgID *identity.ID,
	encryptingKeyID *identity.ID) (*primitive.PublicKey, error) {

	claimTrees, err := client.ClaimTree.List(ctx, orgID, nil)
	if err != nil {
		return nil, err
	}

	if len(claimTrees) != 1 {
		err = fmt.Errorf("No claim tree found for org: %s", orgID)
		return nil, err
	}

	var encryptingKey *primitive.PublicKey
	for _, segment := range claimTrees[0].PublicKeys {
		if *segment.Key.ID == *encryptingKeyID {
			encryptingKey = segment.Key.Body.(*primitive.PublicKey)
			break
		}
	}
	if encryptingKey == nil {
		err = fmt.Errorf("Couldn't find encrypting key %s", encryptingKeyID)
		return nil, err
	}

	return encryptingKey, nil
}

// findMembersTeam takes in a list of team objects and returns the members team.
func findMembersTeam(teams []envelope.Unsigned) (*envelope.Unsigned, error) {
	var team *envelope.Unsigned
	for _, t := range teams {
		tBody := t.Body.(*primitive.Team)

		if tBody.Name == "member" && tBody.TeamType == primitive.SystemTeam {
			team = &t
			break
		}
	}

	if team == nil {
		return nil, fmt.Errorf("couldn't find members team")
	}

	return team, nil
}
