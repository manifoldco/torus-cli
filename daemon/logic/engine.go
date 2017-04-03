// Package logic exposes the core logic engine used for working with keyrings,
// keys, claims, teams, memberships, orgs, and other primitive objects core
// to the cryptography architecture
package logic

import (
	"context"
	"log"

	"github.com/manifoldco/go-base64"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
	"github.com/manifoldco/torus-cli/registry"

	"github.com/manifoldco/torus-cli/daemon/crypto"
	"github.com/manifoldco/torus-cli/daemon/observer"
	"github.com/manifoldco/torus-cli/daemon/session"
)

// Engine exposes methods for performing actions that will affect the keys,
// keyrings, keyring memberships, or credential objects.
//
// All data passing in and out of the engine is unencrypted for the currently
// logged in user.
type Engine struct {
	session session.Session
	db      Database
	crypto  *crypto.Engine
	client  *registry.Client

	Worklog Worklog
	Machine Machine
	Session Session
}

// Database interface for logic engine
type Database interface {
	Set(envs ...envelope.Envelope) error
}

// NewEngine returns a new Engine
func NewEngine(s session.Session, db Database, e *crypto.Engine,
	client *registry.Client) *Engine {
	engine := &Engine{
		session: s,
		db:      db,
		crypto:  e,
		client:  client,
	}
	engine.Worklog = newWorklog(engine)
	engine.Machine = Machine{engine: engine}
	engine.Session = Session{engine: engine}
	return engine
}

// AppendCredential attempts to append a plain-text Credential object to the
// Credential Graph.
func (e *Engine) AppendCredential(ctx context.Context, notifier *observer.Notifier,
	cred *PlaintextCredentialEnvelope) (*PlaintextCredentialEnvelope, error) {

	n := notifier.Notifier(4)

	// Ensure we have an existing keyring for this credential's pathexp
	graphs, err := e.client.CredentialGraph.List(ctx, "", cred.Body.PathExp,
		e.session.AuthID())
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
		newGraph, err = createCredentialGraph(ctx, cred.Body, graph,
			sigID, encID, kp, e.client, e.crypto)
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
			KeyringID: graph.GetKeyring().GetID(),
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
		credBody.Previous = previousCred.GetID()
		credBody.CredentialVersion = previousCred.CredentialVersion() + 1
	}

	krm, mekshare, err := graph.FindMember(e.session.AuthID())
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
	if err != nil {
		log.Printf("Error encrypting credential: %s", err)
		return nil, err
	}

	credBody.Nonce = base64.New(cekNonce)

	credBody.Credential.Nonce = base64.New(ctNonce)
	credBody.Credential.Value = base64.New(ct)

	signed, err := e.crypto.SignedCredential(ctx, &credBody, sigID, &kp.Signature)
	if err != nil {
		log.Printf("Error signing credential body: %s", err)
		return nil, err
	}

	n.Notify(observer.Progress, "Credential encrypted", true)

	if newGraph != nil {
		newGraph.Credentials = []envelope.CredentialInf{signed}
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
	notifier *observer.Notifier, cpath, cpathexp *string) ([]PlaintextCredentialEnvelope, error) {
	if cpath != nil && cpathexp != nil {
		panic("cannot use both cpath and cpathexp")
	}
	if cpath == nil && cpathexp == nil {
		panic("cpath or cpathexp required")
	}

	var err error
	var graphs []registry.CredentialGraph
	if cpath != nil {
		graphs, err = e.client.CredentialGraph.List(ctx, *cpath, nil, e.session.AuthID())
	} else if cpathexp != nil {
		graphs, err = e.client.CredentialGraph.Search(ctx, *cpathexp, e.session.AuthID())
	}
	if err != nil {
		log.Printf("error retrieving credential graphs: %s", err)
		return nil, err
	}

	cgs := newCredentialGraphSet()
	err = cgs.Add(graphs...)
	if err != nil {
		return nil, err
	}

	activeGraphs, err := cgs.Prune()
	if err != nil {
		return nil, err
	}

	var steps uint = 1
	for _, graph := range activeGraphs {
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
	for _, graph := range activeGraphs {
		orgID := graph.GetKeyring().OrgID()
		kp, ok := keypairs[*orgID]
		if !ok {
			_, _, kp, err = fetchKeyPairs(ctx, e.client, orgID)
			if err != nil {
				log.Printf("Error fetching keypairs: %s", err)
				return nil, err
			}
			keypairs[*orgID] = kp
		}

		krm, mekshare, err := graph.FindMember(e.session.AuthID())
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
				state := "set"
				if cred.Unset() {
					state = "unset"
				}

				pt, err := u.Unbox(ctx, *cred.Credential().Value, *cred.Nonce(), *cred.Credential().Nonce)
				if err != nil {
					log.Printf("Error decrypting credential: %s", err)
					return err
				}

				plainCred := PlaintextCredentialEnvelope{
					ID:      cred.GetID(),
					Version: cred.GetVersion(),
					Body: &PlaintextCredential{
						Name:      cred.Name(),
						PathExp:   cred.PathExp(),
						ProjectID: cred.ProjectID(),
						OrgID:     cred.OrgID(),
						Value:     string(pt),
						State:     &state,
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
	InviteID *identity.ID) (*envelope.OrgInvite, error) {

	n := notifier.Notifier(3)

	invite, err := e.client.OrgInvites.Get(ctx, InviteID)
	if err != nil {
		log.Printf("could not fetch org invitation: %s", err)
		return nil, err
	}

	if invite.Body.State != primitive.OrgInviteAcceptedState {
		log.Printf("invitation not in accepted state: %s", invite.Body.State)
		return nil, &apitypes.Error{
			Type: apitypes.BadRequestError,
			Err:  []string{"Invite must be accepted before it can be approved"},
		}
	}

	n.Notify(observer.Progress, "Invite retrieved", true)

	v1members, v2members, err := createKeyringMemberships(ctx, e.crypto,
		e.client, e.session, invite.Body.OrgID, invite.Body.InviteeID)
	if err != nil {
		return nil, err
	}

	n.Notify(observer.Progress, "Keyring memberships created", true)

	invite, err = e.client.OrgInvites.Approve(ctx, InviteID)
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

// GenerateKeypairs creates a signing and encrypting keypair for the current
// user for the given organization.
func (e *Engine) GenerateKeypairs(ctx context.Context, notifier *observer.Notifier,
	OrgID *identity.ID) error {

	n := notifier.Notifier(4)

	kp, err := e.crypto.GenerateKeyPairs(ctx)
	if err != nil {
		log.Printf("Error generating keypairs: %s", err)
		return err
	}

	n.Notify(observer.Progress, "Keypairs generated", true)

	pubsig, privsig, err := packageSigningKeypair(ctx, e.crypto, e.session.AuthID(),
		OrgID, kp)
	if err != nil {
		log.Printf("Error packaging signing keypair: %s", err)
		return err
	}

	sigBody := primitive.NewClaim(OrgID, e.session.AuthID(), pubsig.ID, pubsig.ID,
		primitive.SignatureClaimType)
	sigclaim, err := e.crypto.SignedClaim(ctx, sigBody, pubsig.ID, &kp.Signature)
	if err != nil {
		log.Printf("Error creating signature claim: %s", err)
		return err
	}

	n.Notify(observer.Progress, "Signing keys signed", true)

	pubsig, privsig, claims, err := e.client.KeyPairs.Create(ctx, pubsig,
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

	pubenc, privenc, err := packageEncryptionKeypair(ctx, e.crypto, e.session.AuthID(),
		OrgID, kp, pubsig)
	if err != nil {
		log.Printf("Error packaging encryption keypair: %s", err)
	}

	encBody := primitive.NewClaim(OrgID, e.session.AuthID(), pubenc.ID, pubenc.ID,
		primitive.SignatureClaimType)
	encclaim, err := e.crypto.SignedClaim(ctx, encBody, pubsig.ID, &kp.Signature)
	if err != nil {
		log.Printf("Error creating signature claim for encryption key: %s", err)
		return err
	}

	n.Notify(observer.Progress, "Encryption keys signed", true)

	pubenc, privenc, claims, err = e.client.KeyPairs.Create(ctx, pubenc,
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

// RevokeKeypairs creates revocation claims for the signing and encrypting
// keypair for the current user for the given organization.
//
// A revocation claim is a self-signed claim that effectively deletes the
// keypairs.
func (e *Engine) RevokeKeypairs(ctx context.Context, notifier *observer.Notifier,
	orgID *identity.ID) error {

	n := notifier.Notifier(5)

	encKP, sigKP, err := fetchRegistryKeyPairs(ctx, e.client, orgID)
	if err != nil {
		log.Printf("Error retrieving keypairs: %s", err)
		return err
	}

	n.Notify(observer.Progress, "Keypairs retrieved", true)

	if sigKP == nil { // no active keypairs; nothing to revoke
		log.Println("No keys to revoke")
		return nil
	}

	sigID := sigKP.PublicKey.ID

	kp := bundleKeypairs(sigKP, encKP)

	if encKP != nil { // the encryption keypair might already be revoked
		encID := encKP.PublicKey.ID

		prevEncClaim, err := encKP.HeadClaim()
		if err != nil {
			return err
		}

		encBody := primitive.NewClaim(orgID, e.session.AuthID(), prevEncClaim.ID,
			encID, primitive.RevocationClaimType)
		encclaim, err := e.crypto.SignedClaim(ctx, encBody, sigID, &kp.Signature)
		if err != nil {
			log.Printf("Error creating revocation claim for encryption key: %s", err)
			return err
		}

		n.Notify(observer.Progress, "Encryption keys revoked", true)

		_, err = e.client.Claims.Create(ctx, encclaim)
		if err != nil {
			log.Printf("Error uploading encryption keypair revocation: %s", err)
			return err
		}

		n.Notify(observer.Progress, "Encryption key revocation uploaded", true)
	}

	prevSigClaim, err := sigKP.HeadClaim()
	if err != nil {
		return err
	}

	sigBody := primitive.NewClaim(orgID, e.session.AuthID(), prevSigClaim.ID,
		sigID, primitive.RevocationClaimType)
	sigclaim, err := e.crypto.SignedClaim(ctx, sigBody, sigID, &kp.Signature)
	if err != nil {
		log.Printf("Error creating revocation claim for signing key: %s", err)
		return err
	}

	n.Notify(observer.Progress, "Signing keys revoked", true)

	_, err = e.client.Claims.Create(ctx, sigclaim)
	if err != nil {
		log.Printf("Error uploading signature keypair revocation: %s", err)
		return err
	}

	n.Notify(observer.Progress, "Signing key revocation uploaded", true)

	return nil
}

// ChangePassword returns the new password object and re-encrypted masterkey object
func (e *Engine) ChangePassword(ctx context.Context, newPassword string) (*primitive.UserPassword, *primitive.MasterKey, error) {
	return e.crypto.ChangePassword(ctx, newPassword)
}
