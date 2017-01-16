package logic

import (
	"context"
	"fmt"
	"sort"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/base64"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/pathexp"
	"github.com/manifoldco/torus-cli/primitive"
	"github.com/manifoldco/torus-cli/registry"

	"github.com/manifoldco/torus-cli/daemon/crypto"
	"github.com/manifoldco/torus-cli/daemon/observer"
)

// Worklog holds the logic for discovering and acting on worklog items.
// A Worklog item is some action the user should take, either for
// maintenance (this user should be in this keyring, this invite can be
// approved), or as a preventative measure (this credential value should be
// rotated).
//
// Worklog items may be automatically resolved, or require the user do manually
// perform some action.
type Worklog struct {
	engine   *Engine
	handlers map[apitypes.WorklogType]worklogTypeHandler
}

func newWorklog(e *Engine) Worklog {
	w := Worklog{
		engine: e,
		handlers: map[apitypes.WorklogType]worklogTypeHandler{
			apitypes.SecretRotateWorklogType:    &secretRotateHandler{engine: e},
			apitypes.MissingKeypairsWorklogType: &missingKeypairsHandler{engine: e},
			apitypes.InviteApproveWorklogType:   &inviteApproveHandler{engine: e},
			apitypes.KeyringMembersWorklogType:  &keyringMembersHandler{engine: e},
		},
	}

	return w
}

type worklogTypeHandler interface {
	list(context.Context, *envelope.Org) ([]apitypes.WorklogItem, error)
	resolve(context.Context, *observer.Notifier, *identity.ID,
		*apitypes.WorklogItem) (*apitypes.WorklogResult, error)
	resolveErr() string
}

// List returns the list of all outstanding worklog items for the given org
func (w *Worklog) List(ctx context.Context, orgID *identity.ID,
	itemType apitypes.WorklogType) ([]apitypes.WorklogItem, error) {

	org, err := w.engine.client.Orgs.Get(ctx, orgID)
	if err != nil {
		return nil, err
	}

	var items []apitypes.WorklogItem
	for t, h := range w.handlers {
		if t&itemType == 0 {
			continue
		}

		hItems, err := h.list(ctx, org)
		if err != nil {
			return nil, err
		}
		items = append(items, hItems...)
	}

	return items, nil
}

// Get returns a single worklog item for the given org with the given ident.
func (w *Worklog) Get(ctx context.Context, orgID *identity.ID,
	ident *apitypes.WorklogID) (*apitypes.WorklogItem, error) {

	items, err := w.List(ctx, orgID, ident.Type())
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if *item.ID == *ident {
			return &item, nil
		}
	}

	return nil, nil
}

// Resolve attempts to resolve the worklog item in the given org with the given
// ident.
func (w *Worklog) Resolve(ctx context.Context, n *observer.Notifier,
	orgID *identity.ID, ident *apitypes.WorklogID) (*apitypes.WorklogResult, error) {

	item, err := w.Get(ctx, orgID, ident)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, nil
	}

	handler := w.handlers[item.Type()]
	result, err := handler.resolve(ctx, n, orgID, item)

	// We handle errors from the handler's resolve differently than regular
	// errors; for every other part of the daemon code they're non-errors, sent
	// back to the caller for display. The caller can then continue trying to
	// resolve other worklog items.
	if err != nil {
		result = &apitypes.WorklogResult{
			ID:      item.ID,
			State:   apitypes.ErrorWorklogResult,
			Message: handler.resolveErr() + ": " + err.Error(),
		}
	}

	return result, nil
}

type secretRotateHandler struct {
	engine *Engine
}

func (secretRotateHandler) resolveErr() string {
	// This won't happen, because rotation must be manual. Let's include an
	// error message just in case, though!
	return "Error rotating secret"
}

func (h *secretRotateHandler) list(ctx context.Context, org *envelope.Org) ([]apitypes.WorklogItem, error) {
	projects, err := h.engine.client.Projects.List(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	cgs := newCredentialGraphSet()
	for _, project := range projects {
		graphs, err := h.engine.client.CredentialGraph.Search(ctx,
			"/"+org.Body.Name+"/"+project.Body.Name+"/*/*/*/*",
			h.engine.session.AuthID())
		if err != nil {
			return nil, err
		}

		err = cgs.Add(graphs...)
		if err != nil {
			return nil, err
		}
	}

	needRotation, err := cgs.NeedRotation()
	if err != nil {
		return nil, err
	}

	var items []apitypes.WorklogItem
	for _, cred := range needRotation {
		item := apitypes.WorklogItem{
			Subject: cred.PathExp().String() + "/" + cred.Name(),
			Summary: "A user's access was revoked. This secret's value should be changed.",
		}
		item.CreateID(apitypes.SecretRotateWorklogType)

		items = append(items, item)
	}

	return items, nil
}

func (h *secretRotateHandler) resolve(ctx context.Context, n *observer.Notifier,
	orgID *identity.ID, item *apitypes.WorklogItem) (*apitypes.WorklogResult, error) {
	return &apitypes.WorklogResult{
		ID:      item.ID,
		State:   apitypes.ManualWorklogResult,
		Message: "Please set a new value for the secret at " + item.Subject,
	}, nil
}

type missingKeypairsHandler struct {
	engine *Engine
}

func (missingKeypairsHandler) resolveErr() string {
	return "Error generating keypairs"
}

func (h *missingKeypairsHandler) list(ctx context.Context, org *envelope.Org) ([]apitypes.WorklogItem, error) {
	encClaimed, sigClaimed, err := fetchRegistryKeyPairs(ctx, h.engine.client, org.ID)
	if err != nil {
		return nil, err
	}

	var items []apitypes.WorklogItem
	if encClaimed == nil || sigClaimed == nil {
		item := apitypes.WorklogItem{
			Subject: org.Body.Name,
			Summary: "Signing and encryption keypairs missing for org %s.",
		}

		if encClaimed != nil {
			item.Summary = "Signing keypair missing for org %s."
		} else if sigClaimed != nil {
			item.Summary = "Encryption keypair missing for org %s."
		}

		item.Summary = fmt.Sprintf(item.Summary, org.Body.Name)

		item.CreateID(apitypes.MissingKeypairsWorklogType)

		items = append(items, item)
	}

	return items, nil
}

func (h *missingKeypairsHandler) resolve(ctx context.Context, n *observer.Notifier,
	orgID *identity.ID, item *apitypes.WorklogItem) (*apitypes.WorklogResult, error) {
	err := h.engine.GenerateKeypairs(ctx, n, orgID)
	if err != nil {
		return nil, err
	}

	return &apitypes.WorklogResult{
		ID:      item.ID,
		State:   apitypes.SuccessWorklogResult,
		Message: "Keypairs generated.",
	}, nil
}

type inviteApproveHandler struct {
	engine *Engine
}

func (inviteApproveHandler) resolveErr() string {
	return "Error approving invite"
}

func (h *inviteApproveHandler) list(ctx context.Context, org *envelope.Org) ([]apitypes.WorklogItem, error) {
	invites, err := h.engine.client.OrgInvites.List(ctx, org.ID, []string{"accepted"}, "")
	if err != nil {
		// The user can be unauthorized because they don't have access to
		// invites, so they'll have nothing to do here.
		if apitypes.IsUnauthorizedError(err) {
			return nil, nil
		}

		return nil, err
	}

	var items []apitypes.WorklogItem
	for _, invite := range invites {
		summary := fmt.Sprintf("The invite for %s to org %s is ready for approval.",
			invite.Body.Email, org.Body.Name)
		item := apitypes.WorklogItem{
			Subject:   invite.Body.Email,
			Summary:   summary,
			SubjectID: invite.ID,
		}
		item.CreateID(apitypes.InviteApproveWorklogType)

		items = append(items, item)
	}

	return items, nil
}

func (h *inviteApproveHandler) resolve(ctx context.Context, n *observer.Notifier,
	orgID *identity.ID, item *apitypes.WorklogItem) (*apitypes.WorklogResult, error) {
	_, err := h.engine.ApproveInvite(ctx, n, item.SubjectID)
	if err != nil {
		return nil, err
	}

	return &apitypes.WorklogResult{
		ID:      item.ID,
		State:   apitypes.SuccessWorklogResult,
		Message: "User invite approved and finalized.",
	}, nil
}

type keyringMembersHandler struct {
	engine *Engine
}

func (keyringMembersHandler) resolveErr() string {
	return "Error adding user(s) to keyring"
}

func (h *keyringMembersHandler) list(ctx context.Context, org *envelope.Org) ([]apitypes.WorklogItem, error) {
	// Find all of the credential graphs in this org.
	// We need to get all credential graphs. To do this, we first need to know
	// their pathexps. Use keyring listing for this.
	//
	// The paths map takes care of eliminating multiple versions of the keyring;
	// the subsequent List call will return all versions.
	cgs := newCredentialGraphSet()
	paths := make(map[string]*pathexp.PathExp)
	keyrings, err := h.engine.client.Keyring.List(ctx, org.ID, nil)
	if err != nil {
		return nil, err
	}

	for _, k := range keyrings {
		path := k.GetKeyring().PathExp()
		paths[path.String()] = path
	}

	for _, pe := range paths {
		graphs, err := h.engine.client.CredentialGraph.List(ctx, "", pe, nil)
		if err != nil {
			return nil, err
		}

		err = cgs.Add(graphs...)
		if err != nil {
			return nil, err
		}
	}

	// To find out if any keyrings are missing members, we need walk through all
	// active versions of each keyring, to see if anyone is missing.
	// Inactive versions don't matter, as there is nothing there a user would
	// want to access.
	graphs, err := cgs.Active()
	if err != nil {
		return nil, err
	}

	claimTrees, err := h.engine.client.ClaimTree.List(ctx, org.ID, nil)
	if err != nil {
		return nil, err
	}

	// XXX: this logic will be much different when we selectively encode the
	// members of a keyring based on ACLs.
	members, err := getKeyringMembers(ctx, h.engine.client, org.ID)
	if err != nil {
		return nil, err
	}

	missing := make(map[string]apitypes.WorklogItem)
	for _, graph := range graphs {
		for _, member := range members {
			m, _, err := graph.FindMember(&member)
			if err != nil && err != registry.ErrMemberNotFound {
				return nil, err
			}

			if m != nil {
				continue
			}

			// We now know the user/machine should have access to this keyring,
			// but at the moment they do not. See if they do have a valid
			// encryption key we can use to add them to the keyring.
			targetPubKey, _ := findEncryptionPublicKey(claimTrees, org.ID, &member)
			if targetPubKey == nil { // nothing we can do for this user right now
				continue
			}

			users, err := h.engine.client.Profiles.ListByID(ctx, []identity.ID{member})
			if err != nil {
				return nil, err
			}

			name := users[0].Body.Username
			if _, ok := missing[name]; !ok {
				item := apitypes.WorklogItem{
					Subject:   name,
					Summary:   "This user is missing access to one or more secrets.",
					SubjectID: &member,
				}
				item.CreateID(apitypes.KeyringMembersWorklogType)
				missing[name] = item
			}
			// at least one member is missing. we can stop looking here.
			break
		}
	}

	// Always return the items in a consistent order.
	keys := make([]string, 0, len(missing))
	for k := range missing {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	items := make([]apitypes.WorklogItem, 0, len(missing))
	for _, k := range keys {
		items = append(items, missing[k])
	}

	return items, nil
}

func (h *keyringMembersHandler) resolve(ctx context.Context, n *observer.Notifier,
	orgID *identity.ID, item *apitypes.WorklogItem) (*apitypes.WorklogResult, error) {

	// Preamble. Get the current user's keypairs, and the org's claims for
	// pubkey lookup.
	sigID, encID, kp, err := fetchKeyPairs(ctx, h.engine.client, orgID)
	if err != nil {
		return nil, err
	}

	claimTrees, err := h.engine.client.ClaimTree.List(ctx, orgID, nil)
	if err != nil {
		return nil, err
	}

	// Get all keyrings and find the active ones. We can then add the user to
	// the ones they're missing from.
	cgs := newCredentialGraphSet()
	paths := make(map[string]*pathexp.PathExp)
	keyrings, err := h.engine.client.Keyring.List(ctx, orgID, nil)
	if err != nil {
		return nil, err
	}

	for _, k := range keyrings {
		path := k.GetKeyring().PathExp()
		paths[path.String()] = path
	}

	for _, pe := range paths {
		graphs, err := h.engine.client.CredentialGraph.List(ctx, "", pe, nil)
		if err != nil {
			return nil, err
		}

		err = cgs.Add(graphs...)
		if err != nil {
			return nil, err
		}
	}

	graphs, err := cgs.Active()
	if err != nil {
		return nil, err
	}

	for _, graph := range graphs {
		// See if the user in question is in this keyring
		m, _, err := graph.FindMember(item.SubjectID)
		if err != nil && err != registry.ErrMemberNotFound {
			return nil, err
		}

		if m != nil {
			continue
		}

		// We have a missing user.
		// Get the existing user's share of the master encryption key, find the
		// missing user's public key, clone the existing
		// user's copy of the master encryption key for this keyring, and
		// post the result. Now the user is a member!

		krm, mekshare, err := graph.FindMember(h.engine.session.AuthID())
		if err != nil {
			return nil, err
		}

		encPubKey, err := findEncryptionPublicKeyByID(claimTrees, orgID, krm.EncryptingKeyID)
		if err != nil {
			return nil, err
		}

		targetPubKey, err := findEncryptionPublicKey(claimTrees, orgID, item.SubjectID)
		if err != nil {
			return nil, err
		}

		encMek, nonce, err := h.engine.crypto.CloneMembership(ctx,
			*mekshare.Key.Value, *mekshare.Key.Nonce, &kp.Encryption,
			*encPubKey.Body.Key.Value, *targetPubKey.Body.Key.Value)
		if err != nil {
			return nil, err
		}

		key := &primitive.KeyringMemberKey{
			Algorithm: crypto.EasyBox,
			Nonce:     base64.NewValue(nonce),
			Value:     base64.NewValue(encMek),
		}

		keyring := graph.GetKeyring()
		switch k := keyring.(type) {
		case *envelope.KeyringV1:
			projectID := k.Body.ProjectID
			membership, err := newV1KeyringMember(ctx, h.engine.crypto, orgID, projectID,
				krm.KeyringID, item.SubjectID, targetPubKey.ID, encID, sigID, key, kp)
			if err != nil {
				return nil, err
			}

			_, err = h.engine.client.KeyringMember.Post(ctx, []envelope.KeyringMemberV1{*membership})
			if err != nil {
				return nil, err
			}

		case *envelope.Keyring:
			membership, err := newV2KeyringMember(ctx, h.engine.crypto, orgID, krm.KeyringID,
				item.SubjectID, targetPubKey.ID, encID, sigID, key, kp)
			if err != nil {
				return nil, err
			}
			err = h.engine.client.Keyring.Members.Post(ctx, *membership)
			if err != nil {
				return nil, err
			}
		}
	}

	return &apitypes.WorklogResult{
		ID:      item.ID,
		State:   apitypes.SuccessWorklogResult,
		Message: "User added to missing keyring(s).",
	}, nil
}
