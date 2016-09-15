package logic

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"golang.org/x/crypto/ed25519"

	"github.com/arigatomachine/cli/base64"
	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"

	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/registry"
)

// public key types
const (
	encryptionKeyType = "encryption"
	signingKeyType    = "signing"
)

// createCredentialGraph generates, signs, and posts a new CredentialGraph
// to the registry.
func createCredentialGraph(ctx context.Context, credBody *PlaintextCredential,
	sigID *identity.ID, encID *identity.ID, kp *crypto.KeyPairs,
	client *registry.Client, engine *crypto.Engine) (*registry.CredentialGraphV2, error) {

	pathExp, err := credBody.PathExp.WithInstance("*")
	if err != nil {
		return nil, err
	}

	keyring, err := engine.SignedEnvelope(
		ctx, primitive.NewKeyring(credBody.OrgID, credBody.ProjectID, pathExp),
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
	members := []registry.KeyringMember{}
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

		key := &primitive.KeyringMemberKey{
			Algorithm: crypto.EasyBox,
			Nonce:     base64.NewValue(nonce),
			Value:     base64.NewValue(encmek),
		}

		member, err := newV2KeyringMember(ctx, engine, credBody.OrgID, keyring.ID,
			pubKey.OwnerID, encPubKey.ID, encID, sigID, key, kp)
		if err != nil {
			return nil, err
		}

		members = append(members, *member)
	}

	graph := registry.CredentialGraphV2{
		KeyringSectionV2: registry.KeyringSectionV2{
			Keyring: keyring,
			Claims:  []envelope.Signed{},
			Members: members,
		},
	}

	return &graph, nil
}

func newV1KeyringMember(ctx context.Context, engine *crypto.Engine,
	orgID, projectID, keyringID, ownerID, pubKeyID, encKeyID, sigID *identity.ID,
	key *primitive.KeyringMemberKey, kp *crypto.KeyPairs) (*envelope.Signed, error) {

	now := time.Now().UTC()
	return engine.SignedEnvelope(ctx, &primitive.KeyringMemberV1{
		Created:         now,
		OrgID:           orgID,
		ProjectID:       projectID,
		KeyringID:       keyringID,
		OwnerID:         ownerID,
		PublicKeyID:     pubKeyID,
		EncryptingKeyID: encKeyID,
		Key:             key,
	}, sigID, &kp.Signature)
}

func newV2KeyringMember(ctx context.Context, engine *crypto.Engine,
	orgID, keyringID, ownerID, pubKeyID, encKeyID, sigID *identity.ID,
	key *primitive.KeyringMemberKey, kp *crypto.KeyPairs) (*registry.KeyringMember, error) {

	now := time.Now().UTC()
	member, err := engine.SignedEnvelope(ctx, &primitive.KeyringMember{
		Created:         now,
		OrgID:           orgID,
		KeyringID:       keyringID,
		OwnerID:         ownerID,
		PublicKeyID:     pubKeyID,
		EncryptingKeyID: encKeyID,
	}, sigID, &kp.Signature)

	if err != nil {
		return nil, err
	}

	mekshare, err := engine.SignedEnvelope(ctx, &primitive.MEKShare{
		Created:         now,
		OrgID:           orgID,
		OwnerID:         ownerID,
		KeyringID:       keyringID,
		KeyringMemberID: member.ID,
		Key:             key,
	}, sigID, &kp.Signature)
	if err != nil {
		return nil, err
	}

	return &registry.KeyringMember{
		Member:   member,
		MEKShare: mekshare,
	}, nil
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

func findEncryptionPublicKey(trees []registry.ClaimTree, orgID *identity.ID,
	userID *identity.ID) (*envelope.Signed, error) {

	// Loop over claimtree looking for the users encryption key
	var encKey *envelope.Signed
	for _, tree := range trees {
		if *tree.Org.ID != *orgID {
			continue
		}

		for _, segment := range tree.PublicKeys {
			key := segment.Key
			keyBody := key.Body.(*primitive.PublicKey)
			if *keyBody.OwnerID != *userID {
				continue
			}

			if keyBody.KeyType != encryptionKeyType {
				continue
			}

			encKey = key
		}
	}

	if encKey == nil {
		err := fmt.Errorf("No encryption pubkey found for: %s", userID.String())
		return nil, err
	}

	return encKey, nil
}

func findEncryptionPublicKeyByID(trees []registry.ClaimTree, orgID *identity.ID,
	ID *identity.ID) (*envelope.Signed, error) {

	// Loop over claimtree looking for the users encryption key
	var encKey *envelope.Signed
	for _, tree := range trees {
		if *tree.Org.ID != *orgID {
			continue
		}

		for _, segment := range tree.PublicKeys {
			key := segment.Key
			keyBody := key.Body.(*primitive.PublicKey)
			if *key.ID != *ID {
				continue
			}

			if keyBody.KeyType != encryptionKeyType {
				continue
			}

			encKey = key
		}
	}

	if encKey == nil {
		err := fmt.Errorf("No encryption pubkey found for: %s", ID.String())
		return nil, err
	}

	return encKey, nil
}
func packagePublicKey(ctx context.Context, engine *crypto.Engine, ownerID,
	orgID *identity.ID, keyType string, public []byte, sigID *identity.ID,
	sigKP *crypto.SignatureKeyPair) (*envelope.Signed, error) {

	alg := crypto.Curve25519
	if keyType == signingKeyType {
		alg = crypto.EdDSA
	}

	now := time.Now().UTC()

	body := primitive.PublicKey{
		OrgID:     orgID,
		OwnerID:   ownerID,
		KeyType:   keyType,
		Algorithm: alg,

		Key: primitive.PublicKeyValue{
			Value: base64.NewValue(public),
		},

		Created: now,
		Expires: now.Add(time.Hour * 8760), // one year
	}

	return engine.SignedEnvelope(ctx, &body, sigID, sigKP)
}

func packagePrivateKey(ctx context.Context, engine *crypto.Engine, ownerID,
	orgID *identity.ID, pnonce, private []byte, pubID, sigID *identity.ID,
	sigKP *crypto.SignatureKeyPair) (*envelope.Signed, error) {

	body := primitive.PrivateKey{
		OrgID:       orgID,
		OwnerID:     ownerID,
		PNonce:      base64.NewValue(pnonce),
		PublicKeyID: pubID,

		Key: primitive.PrivateKeyValue{
			Algorithm: crypto.Triplesec,
			Value:     base64.NewValue(private),
		},
	}

	return engine.SignedEnvelope(ctx, &body, sigID, sigKP)
}
