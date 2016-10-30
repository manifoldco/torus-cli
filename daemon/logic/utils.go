package logic

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/ed25519"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"

	"github.com/manifoldco/torus-cli/daemon/crypto"
	"github.com/manifoldco/torus-cli/daemon/observer"
	"github.com/manifoldco/torus-cli/daemon/registry"
	"github.com/manifoldco/torus-cli/daemon/session"
)

// public key types
const (
	encryptionKeyType = "encryption"
	signingKeyType    = "signing"
)

func packageSigningKeypair(ctx context.Context, c *crypto.Engine, authID, orgID *identity.ID,
	kp *crypto.KeyPairs) (*envelope.Signed, *envelope.Signed, error) {

	pubsig, err := packagePublicKey(ctx, c, authID, orgID, signingKeyType, kp.Signature.Public, nil, &kp.Signature)
	if err != nil {
		return nil, nil, err
	}

	privsig, err := packagePrivateKey(ctx, c, authID, orgID, kp.Signature.PNonce,
		kp.Signature.Private, pubsig.ID, pubsig.ID, &kp.Signature)
	if err != nil {
		return nil, nil, err
	}

	return pubsig, privsig, nil
}

func packageEncryptionKeypair(ctx context.Context, c *crypto.Engine, authID, orgID *identity.ID,
	kp *crypto.KeyPairs, pubsig *envelope.Signed) (*envelope.Signed, *envelope.Signed, error) {

	pubenc, err := packagePublicKey(ctx, c, authID, orgID, encryptionKeyType,
		kp.Encryption.Public[:], pubsig.ID, &kp.Signature)
	if err != nil {
		return nil, nil, err
	}

	privenc, err := packagePrivateKey(ctx, c, authID, orgID, kp.Encryption.PNonce,
		kp.Encryption.Private, pubenc.ID, pubsig.ID, &kp.Signature)
	if err != nil {
		return nil, nil, err
	}

	return pubenc, privenc, nil
}

// createCredentialGraph generates, signs, and posts a new CredentialGraph
// to the registry.
func createCredentialGraph(ctx context.Context, credBody *PlaintextCredential,
	parent registry.CredentialGraph, sigID *identity.ID, encID *identity.ID, kp *crypto.KeyPairs,
	client *registry.Client, engine *crypto.Engine) (*registry.CredentialGraphV2, error) {

	pathExp, err := credBody.PathExp.WithInstance("*")
	if err != nil {
		return nil, err
	}

	keyringBody := primitive.NewKeyring(credBody.OrgID, credBody.ProjectID, pathExp)
	if parent != nil {
		keyringBody.Previous = parent.GetKeyring().ID
		keyringBody.KeyringVersion = parent.KeyringVersion() + 1
	}

	keyring, err := engine.SignedEnvelope(ctx, keyringBody, sigID, &kp.Signature)
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
		return nil, &apitypes.Error{
			Type: apitypes.NotFoundError,
			Err: []string{
				fmt.Sprintf("Claim tree not found for org: %s", credBody.OrgID),
			},
		}
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

func createKeyringMemberships(ctx context.Context, c *crypto.Engine, notifier *observer.Notifier,
	client *registry.Client, s session.Session, orgID, ownerID *identity.ID) (
	[]envelope.Signed, []registry.KeyringMember, error) {

	n := notifier.Notifier(3)

	// Get this users keypairs
	sigID, encID, kp, err := fetchKeyPairs(ctx, client, orgID)
	if err != nil {
		log.Printf("could not fetch keypairs for org: %s", err)
		return nil, nil, err
	}

	n.Notify(observer.Progress, "Keypairs retrieved", true)

	claimTrees, err := client.ClaimTree.List(ctx, orgID, nil)
	if err != nil {
		log.Printf("could not retrieve claim tree for invite approval: %s", err)
		return nil, nil, err
	}

	if len(claimTrees) != 1 {
		log.Printf("incorrect number of claim trees returned: %d", len(claimTrees))
		return nil, nil, &apitypes.Error{
			Type: apitypes.NotFoundError,
			Err: []string{
				fmt.Sprintf("Claim tree not found for org: %s", orgID),
			},
		}
	}

	n.Notify(observer.Progress, "Claims retrieved", true)

	// Get all the keyrings and memberships for the current user. This way we
	// can decrypt the MEK for each and then create a new KeyringMember for
	// our wonderful new org member!
	org, err := client.Orgs.Get(ctx, orgID)
	if err != nil {
		return nil, nil, err
	}

	projects, err := client.Projects.List(ctx, org.ID)
	if err != nil {
		return nil, nil, err
	}

	var graphs []registry.CredentialGraph
	orgName := org.Body.(*primitive.Org).Name
	for _, project := range projects {
		projName := project.Body.(*primitive.Project).Name
		projGraphs, err := client.CredentialGraph.Search(ctx,
			"/"+orgName+"/"+projName+"/*/*/*/*", s.AuthID())
		if err != nil {
			log.Printf("Error retrieving credential graphs: %s", err)
			return nil, nil, err
		}

		graphs = append(graphs, projGraphs...)
	}

	// Find encryption keys for user
	targetPubKey, err := findEncryptionPublicKey(claimTrees, orgID, ownerID)
	if err != nil {
		log.Printf("could not find encryption key for owner id: %s", ownerID.String())
		return nil, nil, err
	}

	cgs := newCredentialGraphSet()
	err = cgs.Add(graphs...)
	if err != nil {
		return nil, nil, err
	}

	activeGraphs, err := cgs.Active()
	if err != nil {
		return nil, nil, err
	}

	n.Notify(observer.Progress, "Keyrings retrieved", true)

	v1members := []envelope.Signed{}
	v2members := []registry.KeyringMember{}
	for _, graph := range activeGraphs {
		krm, mekshare, err := graph.FindMember(s.AuthID())
		if err != nil {
			log.Printf("could not find keyring membership: %s", err)
			return nil, nil, &apitypes.Error{
				Type: apitypes.NotFoundError,
				Err:  []string{"Keyring membership not found."},
			}
		}

		encPubKey, err := findEncryptionPublicKeyByID(claimTrees, orgID, krm.EncryptingKeyID)
		if err != nil {
			log.Printf("could not find encypting public key for membership: %s", err)
			return nil, nil, err
		}

		encPKBody := encPubKey.Body.(*primitive.PublicKey)
		targetPKBody := targetPubKey.Body.(*primitive.PublicKey)

		encMek, nonce, err := c.CloneMembership(ctx, *mekshare.Key.Value,
			*mekshare.Key.Nonce, &kp.Encryption, *encPKBody.Key.Value, *targetPKBody.Key.Value)
		if err != nil {
			log.Printf("could not clone keyring membership: %s", err)
			return nil, nil, err
		}

		key := &primitive.KeyringMemberKey{
			Algorithm: crypto.EasyBox,
			Nonce:     base64.NewValue(nonce),
			Value:     base64.NewValue(encMek),
		}

		switch graph.GetKeyring().Version {
		case 1:
			projectID := graph.GetKeyring().Body.(*primitive.KeyringV1).ProjectID
			member, err := newV1KeyringMember(ctx, c, krm.OrgID, projectID,
				krm.KeyringID, ownerID, targetPubKey.ID, encID, sigID, key, kp)
			if err != nil {
				return nil, nil, err
			}
			v1members = append(v1members, *member)
		case 2:
			member, err := newV2KeyringMember(ctx, c, krm.OrgID, krm.KeyringID,
				ownerID, targetPubKey.ID, encID, sigID, key, kp)
			if err != nil {
				return nil, nil, err
			}
			v2members = append(v2members, *member)
		default:
			return nil, nil, &apitypes.Error{
				Type: apitypes.InternalServerError,
				Err:  []string{"Unknown keyring schema version"},
			}
		}
	}

	return v1members, v2members, nil
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
			return nil, nil, nil, &apitypes.Error{
				Type: apitypes.InternalServerError,
				Err:  []string{fmt.Sprintf("Unknown key type: %s", pubKey.KeyType)},
			}
		}
	}

	if sigClaimed.PublicKey == nil || encClaimed.PublicKey == nil {
		return nil, nil, nil, &apitypes.Error{
			Type: apitypes.NotFoundError,
			Err:  []string{"Missing encryption or signing keypairs"},
		}
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
		return nil, &apitypes.Error{
			Type: apitypes.NotFoundError,
			Err: []string{
				fmt.Sprintf("Claim tree not found for org: %s", orgID),
			},
		}
	}

	var encryptingKey *primitive.PublicKey
	for _, segment := range claimTrees[0].PublicKeys {
		if *segment.Key.ID == *encryptingKeyID {
			encryptingKey = segment.Key.Body.(*primitive.PublicKey)
			break
		}
	}
	if encryptingKey == nil {
		return nil, &apitypes.Error{
			Type: apitypes.NotFoundError,
			Err: []string{
				fmt.Sprintf("Encrypting key not found: %s", encryptingKeyID),
			},
		}
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
		return nil, &apitypes.Error{
			Err:  []string{"Member team not found."},
			Type: apitypes.NotFoundError,
		}
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
		return nil, &apitypes.Error{
			Type: apitypes.NotFoundError,
			Err: []string{
				fmt.Sprintf("Encryption pubkey not found for: %s", ID.String()),
			},
		}
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

var errCredVersionMistmach = errors.New("Mismatched credential version and body")

func baseCredential(cred *envelope.Signed) (*primitive.BaseCredential, error) {
	switch v := cred.Version; v {
	case 1:
		if b, ok := cred.Body.(*primitive.CredentialV1); ok {
			return &b.BaseCredential, nil
		}
		return nil, errCredVersionMistmach
	case 2:
		if b, ok := cred.Body.(*primitive.Credential); ok {
			return &b.BaseCredential, nil
		}
		return nil, errCredVersionMistmach
	default:
		return nil, fmt.Errorf("Unknown credential version %d", v)
	}
}
