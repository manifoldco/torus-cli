package logic

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/daemon/crypto"
	"github.com/manifoldco/torus-cli/daemon/observer"
	"github.com/manifoldco/torus-cli/daemon/registry"
	"github.com/manifoldco/torus-cli/daemon/session"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"
)

// Machine represents the business logic for managing machines
type Machine struct {
	engine *Engine
}

// MachineTokenSegment represents a Token and it's associated Keypair
type MachineTokenSegment struct {
	Token   *envelope.Unsigned       `json:"token"`
	Keypair *registry.ClaimedKeyPair `json:"keypair"`
}

// CreateToken generates a new machine token given a machine and a secret value.
func (m *Machine) CreateToken(ctx context.Context, notifier *observer.Notifier,
	machine *envelope.Unsigned, secret *base64.Value) (*registry.MachineTokenCreationSegment, error) {
	n := notifier.Notifier(2)

	n.Notify(observer.Progress, "Generating machine token", true)
	salt, err := crypto.GenerateSalt(ctx)
	if err != nil {
		return nil, err
	}

	keypair, err := crypto.DeriveLoginKeypair(ctx, secret, salt)
	if err != nil {
		return nil, err
	}

	machineBody, ok := machine.Body.(*primitive.Machine)
	if !ok {
		return nil, &apitypes.Error{
			Type: apitypes.InternalServerError,
			Err:  []string{"Could not cast to Machine"},
		}
	}
	orgID := machineBody.OrgID

	masterKey, err := crypto.CreateMasterKeyObject(ctx, secret.String())
	if err != nil {
		return nil, err
	}

	tokenBody := &primitive.MachineToken{
		OrgID:     orgID,
		MachineID: machine.ID,
		PublicKey: &primitive.MachineTokenPublicKey{
			Salt:  keypair.Salt(),
			Value: keypair.PublicKey(),
			Alg:   crypto.EdDSA,
		},
		Master:      masterKey,
		CreatedBy:   m.engine.session.ID(),
		Created:     time.Now().UTC(),
		DestroyedBy: nil,
		Destroyed:   nil,
		State:       primitive.MachineTokenActiveState,
	}
	tokenID, err := identity.NewMutable(tokenBody)
	if err != nil {
		return nil, err
	}

	token := &envelope.Unsigned{
		ID:      &tokenID,
		Version: 1,
		Body:    tokenBody,
	}

	// Create an "empty" machine session in order to create a Crypto engine on
	// behalf of the machine for deriving and uploading these keys.
	sess := session.NewSession()
	err = sess.Set(apitypes.MachineSession, machine, token, secret.String(), "asdfsdf")
	if err != nil {
		return nil, err
	}
	c := crypto.NewEngine(sess)

	n.Notify(observer.Progress, "Generating token keypairs", true)
	kp, err := c.GenerateKeyPairs(ctx)
	if err != nil {
		log.Printf("Error generating machine keypairs: %s", err)
		return nil, err
	}

	authID := sess.AuthID()
	keypairs, err := generateKeypairs(ctx, c, orgID, authID, kp)
	if err != nil {
		return nil, err
	}

	return &registry.MachineTokenCreationSegment{
		Token:    token,
		Keypairs: keypairs,
	}, nil
}

// EncodeToken creates KeyringMemberships for the provided Machine Token. Used
// during the machine creation process
func (m *Machine) EncodeToken(ctx context.Context, notifier *observer.Notifier,
	token *envelope.Unsigned) error {

	n := notifier.Notifier(2)

	n.Notify(observer.Progress, "Creating keyring memberships for token", true)

	tokenBody, ok := token.Body.(*primitive.MachineToken)
	if !ok {
		return errors.New("Could not cast token to MachineToken")
	}

	v1members, v2members, err := createKeyringMemberships(ctx, m.engine.crypto,
		m.engine.client, m.engine.session, tokenBody.OrgID, token.ID)
	if err != nil {
		return err
	}

	n.Notify(observer.Progress, "Uploading keyring memberships", true)

	if len(v1members) != 0 {
		_, err = m.engine.client.KeyringMember.Post(ctx, v1members)
		if err != nil {
			log.Printf("error uploading memberships: %s", err)
			return err
		}
	}

	for _, member := range v2members {
		err = m.engine.client.Keyring.Members.Post(ctx, member)
		if err != nil {
			log.Printf("error uploading memberships: %s", err)
			return err
		}
	}

	return nil
}

func generateKeypairs(ctx context.Context, c *crypto.Engine, orgID, authID *identity.ID,
	kp *crypto.KeyPairs) ([]*registry.ClaimedKeyPair, error) {

	pubsig, privsig, err := packageSigningKeypair(ctx, c, authID, orgID, kp)
	if err != nil {
		log.Printf("Error packaging machine signing keypair: %s", err)
		return nil, err
	}

	rawsigClaim := primitive.NewClaim(orgID, authID, pubsig.ID, pubsig.ID, primitive.SignatureClaimType)
	sigclaim, err := c.SignedEnvelope(ctx, rawsigClaim, pubsig.ID, &kp.Signature)
	if err != nil {
		log.Printf("Error generating signature claim: %s", err)
		return nil, err
	}

	pubenc, privenc, err := packageEncryptionKeypair(ctx, c, authID, orgID, kp, pubsig)
	if err != nil {
		log.Printf("Error packaging machine encryption keypair: %s", err)
		return nil, err
	}

	rawencClaim := primitive.NewClaim(orgID, authID, pubenc.ID, pubenc.ID, primitive.SignatureClaimType)
	encclaim, err := c.SignedEnvelope(ctx, rawencClaim, pubsig.ID, &kp.Signature)
	if err != nil {
		log.Printf("Error generating encryption claim: %s", err)
		return nil, err
	}

	return []*registry.ClaimedKeyPair{
		{
			PublicKey:  pubsig,
			PrivateKey: privsig,
			Claims:     []envelope.Signed{*sigclaim},
		},
		{
			PublicKey:  pubenc,
			PrivateKey: privenc,
			Claims:     []envelope.Signed{*encclaim},
		},
	}, nil
}
