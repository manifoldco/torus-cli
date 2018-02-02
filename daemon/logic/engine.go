// Package logic exposes the core logic engine used for working with keyrings,
// keys, claims, teams, memberships, orgs, and other primitive objects core
// to the cryptography architecture
package logic

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"

	"github.com/manifoldco/go-base64"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
	"github.com/manifoldco/torus-cli/registry"

	"github.com/manifoldco/torus-cli/daemon/crypto"
	"github.com/manifoldco/torus-cli/daemon/crypto/secure"
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
	guard   *secure.Guard

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
	client *registry.Client, guard *secure.Guard) *Engine {
	engine := &Engine{
		session: s,
		db:      db,
		crypto:  e,
		client:  client,
		guard:   guard,
	}
	engine.Worklog = newWorklog(engine)
	engine.Machine = Machine{engine: engine}
	engine.Session = Session{engine: engine}
	return engine
}

// AppendCredentials attempts to append plain-text Credential objects to the
// Credential Graph.
func (e *Engine) AppendCredentials(ctx context.Context, notifier *observer.Notifier,
	creds []*PlaintextCredentialEnvelope) ([]*PlaintextCredentialEnvelope, error) {

	if len(creds) == 0 {
		return creds, nil
	}

	n := notifier.Notifier(3 + uint(len(creds)))
	cred := creds[0]

	// Ensure all creds have matching pathexp
	for _, c := range creds {
		if !cred.Body.PathExp.Equal(c.Body.PathExp) {
			return nil, &apitypes.Error{
				Type: apitypes.BadRequestError,
				Err:  []string{"All credential PathExp must match"},
			}
		}
	}

	// Ensure we have an existing keyring for this credential's pathexp
	graphs, err := e.client.CredentialGraph.List(ctx, "", cred.Body.PathExp,
		e.session.AuthID(), nil)
	if err != nil {
		log.Printf("Error retrieving credential graphs: %s", err)
		return nil, err
	}

	n.Notify(observer.Progress, "Credentials retrieved", true)

	keypairs, err := e.client.KeyPairs.List(ctx, cred.Body.OrgID)
	if err != nil {
		log.Printf("Error fetching keypairs: %s", err)
		return nil, err
	}

	claimtree, err := e.client.ClaimTree.Get(ctx, cred.Body.OrgID, nil)
	if err != nil {
		log.Printf("Error fetching claimtree for org[%s]: %s", cred.Body.OrgID, err)
		return nil, err
	}

	sigID, encID, kp, err := fetchKeyPairs(keypairs, cred.Body.OrgID)
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

	var newGraph *registry.CredentialGraphV2
	// No matching CredentialGraph/KeyRing for this credential.
	// We'll make a new one now.
	if graph == nil || graph.HasRevocations() {
		newGraph, err = createCredentialGraph(ctx, cred.Body, graph,
			sigID, encID, kp, claimtree, e.client, e.crypto, e.guard)
		if err != nil {
			log.Printf("error creating credential graph: %s", err)
			return nil, err
		}
		cgs.Add(newGraph)
		graph = newGraph
	}

	krm, mekshare, err := graph.FindMember(e.session.AuthID())
	if err != nil {
		log.Printf("Error finding keyring membership: %s", err)
		return nil, err
	}

	encKeySegment, err := claimtree.Find(krm.EncryptingKeyID, true)
	if err != nil {
		log.Printf("Error finding encrypting key[%s]: %s", krm.EncryptingKeyID, err)
		return nil, err
	}
	encryptingKey := encKeySegment.PublicKey.Body

	n.Notify(observer.Progress, "Encrypting key retrieved", true)

	toCreate := []envelope.CredentialInf{}
	for _, c := range creds {
		// Find the  most recent version of this credential to act as our previous.
		previousCred, err := cgs.HeadCredential(c.Body.PathExp, c.Body.Name)
		if err != nil {
			log.Printf("error finding credentials to match: %s", err)
			return nil, err
		}

		// Construct an encrypted and signed version of the credential
		credBody := primitive.Credential{
			State: c.Body.State,
			BaseCredential: primitive.BaseCredential{
				Name:      c.Body.Name,
				PathExp:   c.Body.PathExp,
				KeyringID: graph.GetKeyring().GetID(),
				ProjectID: c.Body.ProjectID,
				OrgID:     c.Body.OrgID,
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

		// Derive a key for the credential using the keyring master key
		// and use the derived key to encrypt the credential
		cekNonce, ctNonce, ct, err := e.crypto.BoxCredential(
			ctx, []byte(c.Body.Value), *mekshare.Key.Value, *mekshare.Key.Nonce,
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

		toCreate = append(toCreate, signed)
		n.Notify(observer.Progress, "Credential encrypted", true)
	}

	// XXX: At the end do this stuff (support multi!)
	if newGraph != nil {
		newGraph.Credentials = toCreate
		_, err = e.client.CredentialGraph.Post(ctx, &graph)
	} else {
		_, err = e.client.Credentials.Create(ctx, toCreate)
	}

	if err != nil {
		log.Printf("error creating credential: %s", err)
		return nil, err
	}

	return creds, nil
}

// RetrieveCredentials returns all credentials for the given CPath string
func (e *Engine) RetrieveCredentials(ctx context.Context,
	notifier *observer.Notifier, cpath, cpathexp *string, teamIDs []identity.ID, skipDecryption bool) ([]PlaintextCredentialEnvelope, error) {
	if cpath != nil && cpathexp != nil {
		panic("cannot use both cpath and cpathexp")
	}
	if cpath == nil && cpathexp == nil {
		panic("cpath or cpathexp required")
	}

	var graphs []registry.CredentialGraph
	var err error
	if cpath != nil {
		graphs, err = e.client.CredentialGraph.List(ctx, *cpath, nil, e.session.AuthID(), teamIDs)
	} else if cpathexp != nil {
		graphs, err = e.client.CredentialGraph.Search(ctx, *cpathexp, e.session.AuthID(), teamIDs)
	}

	if err != nil {
		log.Printf("error retrieving credential graph: %s", err)
		return nil, err
	}

	cgs := newCredentialGraphSet()
	err = cgs.Add(graphs...)
	if err != nil {
		log.Printf("error creating credential graph set: %s", err)
		return nil, err
	}

	// Prune removes all unactive graphs (those without a head credential) and
	// unset credentials.
	activeGraphs, err := cgs.Prune()
	if err != nil {
		log.Printf("error encountered while pruning graph: %s", err)
		return nil, err
	}

	creds := []PlaintextCredentialEnvelope{}
	if len(activeGraphs) == 0 {
		log.Printf("no active graphs found")
		return creds, nil
	}

	var steps uint = 1
	for _, graph := range activeGraphs {
		steps += uint(len(graph.GetCredentials()))
	}

	n := notifier.Notifier(steps)
	n.Notify(observer.Progress, "Credentials retrieved", true)

	if skipDecryption {
		encrypted := []PlaintextCredentialEnvelope{}
		log.Printf("skipping decryption of credentials")
		for _, graph := range activeGraphs {
			for _, cred := range graph.GetCredentials() {
				// If we encounter a v1 credential then we have to decrypt it
				// to get the credential value.
				//
				// Very few v1 credentials exist so we can just decrypt
				// everything in those cases.
				if cred.GetVersion() == 1 {
					log.Printf("encountered a v1 credential; forcing decryption")
					goto Decryption
				}

				cValue := apitypes.NewUndecryptedCredentialValue()
				bv, err := json.Marshal(cValue)
				if err != nil {
					log.Printf("could not marshal undecrypted cvalue: %s", err)
					return nil, err
				}

				cv, err := strconv.Unquote(string(bv))
				if err != nil {
					return nil, err
				}
				encrypted = append(encrypted, packagePlaintextCred(cred, cv))
			}
		}

		return encrypted, nil
	}

Decryption:
	// Loop over the trees and unpack the credentials; later on we will
	// actually do real work and decrypt each of these credentials but for
	// now we just need ot return a list of them!
	idx := newCredentialGraphKeyIndex(*(e.session.AuthID()))
	idx.Add(activeGraphs...)

	// All graphs will belong to the same org
	orgID := activeGraphs[0].GetKeyring().OrgID()

	var fetchKeys sync.WaitGroup
	var kps *registry.Keypairs
	var claimtree *registry.ClaimTree
	var kpsErr, ctErr error
	fetchKeys.Add(2)

	// Fetch the user's keypairs for this specific organization
	go func() {
		kps, kpsErr = e.client.KeyPairs.List(ctx, orgID)
		fetchKeys.Done()
	}()

	// Fetch the org's claimtree which will include all public keys and their
	// claims for all users and machines inside the org
	go func() {
		claimtree, ctErr = e.client.ClaimTree.Get(ctx, orgID, nil)
		fetchKeys.Done()
	}()

	fetchKeys.Wait()
	if kpsErr != nil {
		log.Printf("Cannot fetch keypairs for org[%s]: %s", orgID, kpsErr)
		return nil, kpsErr
	}
	if ctErr != nil {
		log.Printf("Could not fetch claimtree for org[%s]: %s", orgID, ctErr)
		return nil, ctErr
	}

	// Cache the bundled crypto keypairs for reuse
	keypairs := make(map[identity.ID]*crypto.KeyPairs)
	for encryptingKeyID, graphs := range idx.GetIndex() {
		if len(graphs) == 0 {
			continue
		}

		kp, ok := keypairs[*orgID]
		if !ok {
			_, _, kp, err = fetchKeyPairs(kps, orgID)
			if err != nil {
				log.Printf("Error fetching keypairs: %s", err)
				return nil, err
			}
			keypairs[*orgID] = kp
		}

		encryptingKeySegment, err := claimtree.Find(&encryptingKeyID, false)
		if err != nil {
			log.Printf("Could not find encrypting key[%s]: %s", encryptingKeyID, err)
			return nil, err
		}

		encryptingKey := encryptingKeySegment.PublicKey.Body
		err = e.crypto.WithUnsealer(ctx, &kp.Encryption, *encryptingKey.Key.Value, func(unsealer crypto.Unsealer) error {
			for _, graph := range graphs {
				mekshare, err := graph.FindMEKByKeyID(&encryptingKeyID)
				if err != nil {
					log.Printf("Error finding keyring membership: %s %s", encryptingKeyID, err)
					return err
				}

				err = unsealer.WithUnboxer(ctx, *mekshare.Key.Value, *mekshare.Key.Nonce, func(u crypto.Unboxer) error {
					for _, cred := range graph.GetCredentials() {
						pt, err := u.Unbox(ctx, *cred.Credential().Value, *cred.Nonce(), *cred.Credential().Nonce)
						if err != nil {
							log.Printf("Error decrypting credential: %s", err)
							return err
						}

						// If this is a v1 credential, then we need to unmarshal the
						// plain text value to check whether or not we should return
						// the credentials.
						if cred.GetVersion() == 1 {
							cValue, err := extractCredentialValue(pt)
							if err != nil {
								log.Printf("could not unmarshal credential value from v1 cred: %s", err)
								return err
							}

							if cValue.IsUnset() {
								continue
							}
						}

						creds = append(creds, packagePlaintextCred(cred, string(pt)))
						n.Notify(observer.Progress, "Credential decrypted", true)
					}
					return nil
				})
				if err != nil {
					log.Printf("encountered an error while unboxing: %s", err)
					return err
				}
			}

			return nil
		})
		if err != nil {
			log.Printf("encountered an error while unsealing: %s", err)
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

	keypairs, err := e.client.KeyPairs.List(ctx, orgID)
	if err != nil {
		log.Printf("Error retrieving keypairs: %s", err)
		return err
	}

	encKP, err := keypairs.Select(orgID, primitive.EncryptionKeyType)
	if err == registry.ErrMissingValidKeypair {
		log.Printf("No keys to revoke, can't find encryption keypair")
		return nil
	}
	if err != nil {
		log.Printf("Could not find encryption keypair: %s", err)
		return err
	}

	sigKP, err := keypairs.Select(orgID, primitive.SigningKeyType)
	if err == registry.ErrMissingValidKeypair {
		log.Printf("No keys to revoke, can't find signing keypair")
		return nil
	}
	if err != nil {
		log.Printf("Could not find signing keypair: %s", err)
		return err
	}

	n.Notify(observer.Progress, "Keypairs retrieved", true)

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
