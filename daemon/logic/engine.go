// Package logic exposes the core logic engine used for working with keyrings,
// keys, claims, teams, memberships, orgs, and other primitive objects core
// to the cryptography architecture
package logic

import (
	"context"
	"fmt"
	"log"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/base64"
	"github.com/arigatomachine/cli/config"
	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/primitive"

	"github.com/arigatomachine/cli/daemon/crypto"
	"github.com/arigatomachine/cli/daemon/db"
	"github.com/arigatomachine/cli/daemon/observer"
	"github.com/arigatomachine/cli/daemon/registry"
	"github.com/arigatomachine/cli/daemon/session"
)

// Engine exposes methods for performing actions that will affect the keys,
// keyrings, keyring memberships, or credential objects.
//
// All data passing in and out of the engine is unencrypted for the currently
// logged in user.
type Engine struct {
	config  *config.Config
	session session.Session
	db      *db.DB
	crypto  *crypto.Engine
	client  *registry.Client

	Worklog Worklog
}

// NewEngine returns a new Engine
func NewEngine(c *config.Config, s session.Session, db *db.DB, e *crypto.Engine,
	client *registry.Client) *Engine {
	engine := &Engine{
		config:  c,
		session: s,
		db:      db,
		crypto:  e,
		client:  client,
	}
	engine.Worklog = Worklog{engine: engine}
	return engine
}

// AppendCredential attempts to append a plain-text Credential object to the
// Credential Graph.
func (e *Engine) AppendCredential(ctx context.Context, notifier *observer.Notifier,
	cred *PlaintextCredentialEnvelope) (*PlaintextCredentialEnvelope, error) {

	n := notifier.Notifier(4)

	// Ensure we have an existing keyring for this credential's pathexp
	graphs, err := e.client.CredentialGraph.List(ctx, "", cred.Body.PathExp,
		e.session.ID())
	if err != nil {
		log.Printf("Error retrieving credential graphs: %s", err)
		return nil, err
	}

	n.Notify(observer.Progress, "Credentials retrieved", true)

	sigID, encID, kp, err := fetchKeyPairs(ctx, e.client, cred.Body.OrgID)
	if err != nil {
		log.Printf("Error fetching keypairs: %s", err)
		return nil, err
	}

	n.Notify(observer.Progress, "Keypairs retrieved", true)

	cgs := newCredentialGraphSet()
	err = cgs.Add(graphs...)
	if err != nil {
		return nil, err
	}

	// Find the credentialgraph/keyring that we should store our credential in
	graph, err := cgs.Head(cred.Body.PathExp)
	if err != nil {
		return nil, err
	}

	// Find the  most recent version of this credential to act as our previous.
	previousCred, err := cgs.HeadCredential(cred.Body.PathExp, cred.Body.Name)
	if err != nil {
		log.Printf("error finding credentials to match: %s", err)
		return nil, err
	}

	var newGraph *registry.CredentialGraphV2
	// No matching CredentialGraph/KeyRing for this credential.
	// We'll make a new one now.
	if graph == nil || graph.HasRevocations() {
		newGraph, err = createCredentialGraph(ctx, cred.Body, graph, sigID,
			encID, kp, e.client, e.crypto)
		if err != nil {
			log.Printf("error creating credential graph: %s", err)
			return nil, err
		}
		cgs.Add(newGraph)
		graph = newGraph
	}

	// Construct an encrypted and signed version of the credential
	credBody := primitive.Credential{
		State: cred.Body.State,
		BaseCredential: primitive.BaseCredential{
			Name:      cred.Body.Name,
			PathExp:   cred.Body.PathExp,
			KeyringID: graph.GetKeyring().ID,
			ProjectID: cred.Body.ProjectID,
			OrgID:     cred.Body.OrgID,
			Credential: &primitive.CredentialValue{
				Algorithm: crypto.SecretBox,
			},
		},
	}

	if previousCred == nil {
		log.Printf("no previous")
		credBody.Previous = nil
		credBody.CredentialVersion = 1
	} else {
		base, err := baseCredential(previousCred)
		if err != nil {
			return nil, err
		}

		credBody.Previous = previousCred.ID
		credBody.CredentialVersion = base.CredentialVersion + 1
	}

	krm, mekshare, err := graph.FindMember(e.session.ID())
	if err != nil {
		log.Printf("Error finding keyring membership: %s", err)
		return nil, err
	}

	encryptingKey, err := findEncryptingKey(ctx, e.client, credBody.OrgID,
		krm.EncryptingKeyID)
	if err != nil {
		log.Printf("Error finding encrypting key: %s", err)
		return nil, err
	}

	n.Notify(observer.Progress, "Encrypting key retrieved", true)

	// Derive a key for the credential using the keyring master key
	// and use the derived key to encrypt the credential
	cekNonce, ctNonce, ct, err := e.crypto.BoxCredential(
		ctx, []byte(cred.Body.Value), *mekshare.Key.Value, *mekshare.Key.Nonce,
		&kp.Encryption, *encryptingKey.Key.Value)

	credBody.Nonce = base64.NewValue(cekNonce)

	credBody.Credential.Nonce = base64.NewValue(ctNonce)
	credBody.Credential.Value = base64.NewValue(ct)

	signed, err := e.crypto.SignedEnvelope(ctx, &credBody, sigID, &kp.Signature)
	if err != nil {
		log.Printf("Error signing credential body: %s", err)
		return nil, err
	}

	n.Notify(observer.Progress, "Credential encrypted", true)

	if newGraph != nil {
		newGraph.Credentials = []envelope.Signed{*signed}
		_, err = e.client.CredentialGraph.Post(ctx, &graph)
	} else {
		_, err = e.client.Credentials.Create(ctx, signed)
	}

	if err != nil {
		log.Printf("error creating credential: %s", err)
		return nil, err
	}

	return cred, nil
}

// RetrieveCredentials returns all credentials for the given CPath string
func (e *Engine) RetrieveCredentials(ctx context.Context,
	notifier *observer.Notifier, cpath string) ([]PlaintextCredentialEnvelope, error) {

	graphs, err := e.client.CredentialGraph.List(ctx, cpath, nil, e.session.ID())
	if err != nil {
		log.Printf("error retrieving credential graphs: %s", err)
		return nil, err
	}

	var steps uint = 1
	for _, graph := range graphs {
		steps += uint(len(graph.GetCredentials()))
	}

	n := notifier.Notifier(steps)
	n.Notify(observer.Progress, "Credentials retrieved", true)

	keypairs := make(map[identity.ID]*crypto.KeyPairs)
	encryptingKeys := make(map[identity.ID]*primitive.PublicKey)

	// Loop over the trees and unpack the credentials; later on we will
	// actually do real work and decrypt each of these credentials but for
	// now we just need ot return a list of them!
	creds := []PlaintextCredentialEnvelope{}
	for _, graph := range graphs {
		var orgID *identity.ID
		switch b := graph.GetKeyring().Body.(type) {
		case *primitive.Keyring:
			orgID = b.OrgID
		case *primitive.KeyringV1:
			orgID = b.OrgID
		default:
			return nil, fmt.Errorf("Malformed keyring body")
		}
		kp, ok := keypairs[*orgID]
		if !ok {
			_, _, kp, err = fetchKeyPairs(ctx, e.client, orgID)
			if err != nil {
				log.Printf("Error fetching keypairs: %s", err)
				return nil, err
			}
			keypairs[*orgID] = kp
		}

		krm, mekshare, err := graph.FindMember(e.session.ID())
		if err != nil {
			log.Printf("Error finding keyring membership: %s", err)
			return nil, err
		}

		encryptingKey, ok := encryptingKeys[*krm.EncryptingKeyID]
		if !ok {
			encryptingKey, err = findEncryptingKey(ctx, e.client, orgID,
				krm.EncryptingKeyID)
			if err != nil {
				log.Printf("Error finding encrypting key for user: %s", err)
				return nil, err
			}
			encryptingKeys[*krm.EncryptingKeyID] = encryptingKey
		}

		err = e.crypto.WithUnboxer(ctx, *mekshare.Key.Value, *mekshare.Key.Nonce, &kp.Encryption, *encryptingKey.Key.Value, func(u crypto.Unboxer) error {
			for _, cred := range graph.GetCredentials() {
				var state *string

				base, err := baseCredential(&cred)
				if err != nil {
					return err
				}

				if c, ok := cred.Body.(*primitive.Credential); ok {
					state = c.State
				}

				pt, err := u.Unbox(ctx, *base.Credential.Value, *base.Nonce, *base.Credential.Nonce)
				if err != nil {
					log.Printf("Error decrypting credential: %s", err)
					return err
				}

				plainCred := PlaintextCredentialEnvelope{
					ID:      cred.ID,
					Version: cred.Version,
					Body: &PlaintextCredential{
						Name:      base.Name,
						PathExp:   base.PathExp,
						ProjectID: base.ProjectID,
						OrgID:     base.OrgID,
						Value:     string(pt),
						State:     state,
					},
				}
				creds = append(creds, plainCred)

				n.Notify(observer.Progress, "Credential decrypted", true)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return creds, nil
}

// ApproveInvite approves an invitation of a user into an organzation by
// encoding them into a Keyring.
func (e *Engine) ApproveInvite(ctx context.Context, notifier *observer.Notifier,
	InviteID *identity.ID) (*envelope.Unsigned, error) {

	n := notifier.Notifier(6)

	invite, err := e.client.OrgInvite.Get(ctx, InviteID)
	if err != nil {
		log.Printf("could not fetch org invitation: %s", err)
		return nil, err
	}

	inviteBody := invite.Body.(*primitive.OrgInvite)

	if inviteBody.State != primitive.OrgInviteAcceptedState {
		log.Printf("invitation not in accepted state: %s", inviteBody.State)
		return nil, &apitypes.Error{
			Type: apitypes.BadRequestError,
			Err:  []string{"Invite must be accepted before it can be approved"},
		}
	}

	n.Notify(observer.Progress, "Invite retrieved", true)

	// Get this users keypairs
	sigID, encID, kp, err := fetchKeyPairs(ctx, e.client, inviteBody.OrgID)
	if err != nil {
		log.Printf("could not fetch keypairs for org: %s", err)
		return nil, err
	}

	n.Notify(observer.Progress, "Keypairs retrieved", true)

	claimTrees, err := e.client.ClaimTree.List(ctx, inviteBody.OrgID, nil)
	if err != nil {
		log.Printf("could not retrieve claim tree for invite approval: %s", err)
		return nil, err
	}

	if len(claimTrees) != 1 {
		log.Printf("incorrect number of claim trees returned: %d", len(claimTrees))
		return nil, fmt.Errorf(
			"No claim tree found for org: %s", inviteBody.OrgID)
	}

	n.Notify(observer.Progress, "Claims retrieved", true)

	// Get all the keyrings and memberships for the current user. This way we
	// can decrypt the MEK for each and then create a new KeyringMember for
	// our wonderful new org member!
	org, err := e.client.Orgs.Get(ctx, inviteBody.OrgID)
	if err != nil {
		return nil, err
	}

	projects, err := e.client.Projects.List(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	var graphs []registry.CredentialGraph
	orgName := org.Body.(*primitive.Org).Name
	for _, project := range projects {
		projName := project.Body.(*primitive.Project).Name
		projGraphs, err := e.client.CredentialGraph.Search(ctx,
			"/"+orgName+"/"+projName+"/*/*/*/*", e.session.ID())
		if err != nil {
			log.Printf("Error retrieving credential graphs: %s", err)
			return nil, err
		}

		graphs = append(graphs, projGraphs...)
	}

	// Find encryption keys for user
	targetPubKey, err := findEncryptionPublicKey(claimTrees,
		inviteBody.OrgID, inviteBody.InviteeID)
	if err != nil {
		log.Printf("could not find encryption key for invitee: %s",
			inviteBody.InviteeID.String())
		return nil, err
	}

	cgs := newCredentialGraphSet()
	err = cgs.Add(graphs...)
	if err != nil {
		return nil, err
	}

	activeGraphs, err := cgs.Active()
	if err != nil {
		return nil, err
	}

	n.Notify(observer.Progress, "Keyrings retrieved", true)

	v1members := []envelope.Signed{}
	v2members := []registry.KeyringMember{}
	for _, graph := range activeGraphs {
		krm, mekshare, err := graph.FindMember(e.session.ID())
		if err != nil {
			log.Printf("could not find keyring membership: %s", err)
			return nil, fmt.Errorf("could not find keyring membership")
		}

		encPubKey, err := findEncryptionPublicKeyByID(claimTrees, inviteBody.OrgID, krm.EncryptingKeyID)
		if err != nil {
			log.Printf("could not find encypting public key for membership: %s", err)
			return nil, err
		}

		encPKBody := encPubKey.Body.(*primitive.PublicKey)
		targetPKBody := targetPubKey.Body.(*primitive.PublicKey)

		encMek, nonce, err := e.crypto.CloneMembership(ctx, *mekshare.Key.Value,
			*mekshare.Key.Nonce, &kp.Encryption, *encPKBody.Key.Value, *targetPKBody.Key.Value)
		if err != nil {
			log.Printf("could not clone keyring membership: %s", err)
			return nil, err
		}

		key := &primitive.KeyringMemberKey{
			Algorithm: crypto.EasyBox,
			Nonce:     base64.NewValue(nonce),
			Value:     base64.NewValue(encMek),
		}

		switch graph.GetKeyring().Version {
		case 1:
			projectID := graph.GetKeyring().Body.(*primitive.KeyringV1).ProjectID
			member, err := newV1KeyringMember(ctx, e.crypto, krm.OrgID, projectID,
				krm.KeyringID, inviteBody.InviteeID, targetPubKey.ID, encID, sigID, key, kp)
			if err != nil {
				return nil, err
			}
			v1members = append(v1members, *member)
		case 2:
			member, err := newV2KeyringMember(ctx, e.crypto, krm.OrgID, krm.KeyringID,
				inviteBody.InviteeID, targetPubKey.ID, encID, sigID, key, kp)
			if err != nil {
				return nil, err
			}
			v2members = append(v2members, *member)
		default:
			return nil, fmt.Errorf("Unknown keyring schema version")
		}
	}

	n.Notify(observer.Progress, "Keyring memberships created", true)

	invite, err = e.client.OrgInvite.Approve(ctx, InviteID)
	if err != nil {
		log.Printf("could not approve org invite: %s", err)
		return nil, err
	}

	n.Notify(observer.Progress, "Invite approved", true)

	if len(v1members) != 0 {
		_, err = e.client.KeyringMember.Post(ctx, v1members)
		if err != nil {
			log.Printf("error uploading memberships: %s", err)
			return nil, err
		}
	}

	for _, member := range v2members {
		err = e.client.Keyring.Members.Post(ctx, member)
		if err != nil {
			log.Printf("error uploading memberships: %s", err)
			return nil, err
		}
	}

	return invite, nil
}

// GenerateKeypair creates a signing and encrypting keypair for the current
// user for the given organization.
func (e *Engine) GenerateKeypair(ctx context.Context, notifier *observer.Notifier,
	OrgID *identity.ID) error {

	n := notifier.Notifier(4)

	kp, err := e.crypto.GenerateKeyPairs(ctx)
	if err != nil {
		log.Printf("Error generating keypairs: %s", err)
		return err
	}

	n.Notify(observer.Progress, "Keypairs generated", true)

	pubsig, err := packagePublicKey(ctx, e.crypto, e.session.ID(), OrgID,
		signingKeyType, kp.Signature.Public, nil, &kp.Signature)
	if err != nil {
		log.Printf("Error packaging signature public key: %s", err)
		return err
	}

	privsig, err := packagePrivateKey(ctx, e.crypto, e.session.ID(), OrgID,
		kp.Signature.PNonce, kp.Signature.Private, pubsig.ID, pubsig.ID,
		&kp.Signature)
	if err != nil {
		log.Printf("Error packaging signing private key: %s", err)
		return err
	}

	sigclaim, err := e.crypto.SignedEnvelope(
		ctx, primitive.NewClaim(OrgID, e.session.ID(), pubsig.ID, pubsig.ID,
			primitive.SignatureClaimType),
		pubsig.ID, &kp.Signature)
	if err != nil {
		log.Printf("Error creating signature claim: %s", err)
		return err
	}

	n.Notify(observer.Progress, "Signing keys signed", true)

	pubsig, privsig, claims, err := e.client.KeyPairs.Post(ctx, pubsig,
		privsig, sigclaim)
	if err != nil {
		log.Printf("Error uploading signature keypair: %s", err)
		return err
	}

	objs := make([]envelope.Envelope, len(claims)+2)
	objs[0] = pubsig
	objs[1] = privsig
	for i, claim := range claims {
		objs[i+2] = &claim
	}
	err = e.db.Set(objs...)
	if err != nil {
		log.Printf("Error storing signing keys in local db: %s", err)
		return err
	}

	n.Notify(observer.Progress, "Signing keys uploaded", true)

	pubenc, err := packagePublicKey(ctx, e.crypto, e.session.ID(), OrgID,
		encryptionKeyType, kp.Encryption.Public[:], pubsig.ID,
		&kp.Signature)
	if err != nil {
		log.Printf("Error packaging encryption public key: %s", err)
		return err
	}

	privenc, err := packagePrivateKey(ctx, e.crypto, e.session.ID(), OrgID,
		kp.Encryption.PNonce, kp.Encryption.Private, pubenc.ID, pubsig.ID,
		&kp.Signature)
	if err != nil {
		log.Printf("Error packaging encryption private key: %s", err)
		return err
	}

	encclaim, err := e.crypto.SignedEnvelope(
		ctx, primitive.NewClaim(OrgID, e.session.ID(), pubenc.ID, pubenc.ID,
			primitive.SignatureClaimType),
		pubsig.ID, &kp.Signature)
	if err != nil {
		log.Printf("Error creating signature claim for encryption key: %s", err)
		return err
	}

	n.Notify(observer.Progress, "Encryption keys signed", true)

	pubenc, privenc, claims, err = e.client.KeyPairs.Post(ctx, pubenc,
		privenc, encclaim)
	if err != nil {
		log.Printf("Error uploading encryption keypair: %s", err)
		return err
	}

	objs = make([]envelope.Envelope, len(claims)+2)
	objs[0] = pubenc
	objs[1] = privenc
	for i, claim := range claims {
		objs[i+2] = &claim
	}
	err = e.db.Set(objs...)
	if err != nil {
		log.Printf("Error storing encryption keys in local db: %s", err)
		return err
	}

	return nil
}
