package logic

import (
	"context"
	"errors"
	"sort"

	"github.com/manifoldco/go-base64"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/pathexp"
	"github.com/manifoldco/torus-cli/primitive"
	"github.com/manifoldco/torus-cli/registry"

	"github.com/manifoldco/torus-cli/daemon/crypto"
	"github.com/manifoldco/torus-cli/daemon/observer"
)

var errManualResolve = errors.New("must be resolved manually")

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
	membersType := apitypes.UserKeyringMembersWorklogType | apitypes.MachineKeyringMembersWorklogType
	w := Worklog{
		engine: e,
		handlers: map[apitypes.WorklogType]worklogTypeHandler{
			apitypes.SecretRotateWorklogType:    &secretRotateHandler{engine: e},
			apitypes.MissingKeypairsWorklogType: &missingKeypairsHandler{engine: e},
			apitypes.InviteApproveWorklogType:   &inviteApproveHandler{engine: e},
			membersType:                         &keyringMembersHandler{engine: e},
		},
	}

	return w
}

type worklogTypeHandler interface {
	list(context.Context, *envelope.Org) ([]apitypes.WorklogItem, error)
	resolve(context.Context, *observer.Notifier, *identity.ID,
		*apitypes.WorklogItem) error
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
	orgID *identity.ID, ident *apitypes.WorklogID) error {

	item, err := w.Get(ctx, orgID, ident)
	if err != nil {
		return err
	}

	if item == nil {
		return nil
	}

	for t, h := range w.handlers {
		if t&item.Type() == 0 {
			continue
		}

		return h.resolve(ctx, n, orgID, item)
	}

	panic("worklog handler not found for type")
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
	for _, reason := range needRotation {
		var ids []identity.ID
		claimsByOwner := make(map[identity.ID]primitive.KeyringMemberClaim, len(reason.Reasons))
		for _, r := range reason.Reasons {
			ids = append(ids, *r.OwnerID)
			claimsByOwner[*r.OwnerID] = r
		}

		users, err := h.engine.client.Profiles.ListByID(ctx, ids)
		if err != nil {
			return nil, err
		}

		var reasons []apitypes.SecretRotateWorklogReason
		for _, user := range users {
			reasons = append(reasons, apitypes.SecretRotateWorklogReason{
				Username: user.Body.Username,
				Type:     claimsByOwner[*user.ID].Reason.Type,
			})
		}

		cred := reason.Credential
		item := apitypes.WorklogItem{
			Details: &apitypes.SecretRotateWorklogDetails{
				PathExp: cred.PathExp(),
				Name:    cred.Name(),
				Reasons: reasons,
			},
		}
		item.CreateID(apitypes.SecretRotateWorklogType)

		items = append(items, item)
	}

	return items, nil
}

func (h *secretRotateHandler) resolve(ctx context.Context, n *observer.Notifier,
	orgID *identity.ID, item *apitypes.WorklogItem) error {
	return errManualResolve
}

type missingKeypairsHandler struct {
	engine *Engine
}

func (missingKeypairsHandler) resolveErr() string {
	return "Error generating keypairs"
}

func (h *missingKeypairsHandler) list(ctx context.Context, org *envelope.Org) ([]apitypes.WorklogItem, error) {
	keypairs, err := h.engine.client.KeyPairs.List(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	encClaimed, err := keypairs.Select(org.ID, primitive.EncryptionKeyType)
	if err != nil {
		return nil, err
	}

	sigClaimed, err := keypairs.Select(org.ID, primitive.SigningKeyType)
	if err != nil {
		return nil, err
	}

	var items []apitypes.WorklogItem
	if encClaimed == nil || sigClaimed == nil {
		item := apitypes.WorklogItem{
			Details: &apitypes.MissingKeypairsWorklogDetails{
				Org:               org.Body.Name,
				EncryptionMissing: encClaimed == nil,
				SigningMissing:    sigClaimed == nil,
			},
		}

		item.CreateID(apitypes.MissingKeypairsWorklogType)

		items = append(items, item)
	}

	return items, nil
}

func (h *missingKeypairsHandler) resolve(ctx context.Context, n *observer.Notifier,
	orgID *identity.ID, item *apitypes.WorklogItem) error {
	return h.engine.GenerateKeypairs(ctx, n, orgID)
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

	teams, err := h.engine.client.Teams.List(ctx, org.ID, "", primitive.AnyTeamType)
	if err != nil {
		return nil, err
	}

	teamsByID := make(map[identity.ID]string)
	for _, t := range teams {
		teamsByID[*t.ID] = t.Body.Name
	}

	var items []apitypes.WorklogItem
	for _, invite := range invites {
		users, err := h.engine.client.Profiles.ListByID(ctx, []identity.ID{*invite.Body.InviteeID})
		if err != nil {
			return nil, err
		}

		var teamNames []string
		for _, t := range invite.Body.PendingTeams {
			teamNames = append(teamNames, teamsByID[t])
		}

		item := apitypes.WorklogItem{
			Details: &apitypes.InviteApproveWorklogDetails{
				InviteID: invite.ID,
				Email:    invite.Body.Email,
				Username: users[0].Body.Username,
				Name:     users[0].Body.Name,
				Org:      org.Body.Name,
				Teams:    teamNames,
			},
		}
		item.CreateID(apitypes.InviteApproveWorklogType)

		items = append(items, item)
	}

	return items, nil
}

func (h *inviteApproveHandler) resolve(ctx context.Context, n *observer.Notifier,
	orgID *identity.ID, item *apitypes.WorklogItem) error {
	_, err := h.engine.ApproveInvite(ctx, n, item.Details.(*apitypes.InviteApproveWorklogDetails).InviteID)
	return err
}

type keyringMembersHandler struct {
	engine *Engine
}

func (keyringMembersHandler) resolveErr() string {
	return "Error adding user(s) to keyring"
}

var errUserNotFound = errors.New("user not found")

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

	claimTree, err := h.engine.client.ClaimTree.Get(ctx, org.ID, nil)
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
	MembersLoop:
		for _, member := range members {
			for _, owner := range member.KeyOwnerIDs() {
				m, _, err := graph.FindMember(&owner)
				if err != nil && err != registry.ErrMemberNotFound {
					return nil, err
				}

				if m != nil {
					continue
				}

				// We now know the user/machine should have access to this keyring,
				// but at the moment they do not. See if they do have a valid
				// encryption key we can use to add them to the keyring.
				_, err = claimTree.FindActive(&owner, primitive.EncryptionKeyType)
				if err == registry.ErrMissingKeyForOwner {
					continue
				}
				if err != nil {
					return nil, err
				}

				// This member is missing access (user, or one or more machine
				// tokens).
				var name string
				var typ apitypes.WorklogType
				switch t := member.(type) {
				case *userKeyringMember:
					users, err := h.engine.client.Profiles.ListByID(ctx, []identity.ID{*member.GetID()})
					if err != nil {
						return nil, err
					}

					if len(users) == 0 {
						return nil, errUserNotFound
					}

					name = users[0].Body.Username
					typ = apitypes.UserKeyringMembersWorklogType
				case *machineKeyringMember:
					name = t.Machine.Body.Name
					typ = apitypes.MachineKeyringMembersWorklogType
				default:
					panic("Unknown keyring member type")
				}

				if _, ok := missing[name]; !ok {
					item := apitypes.WorklogItem{
						Details: &apitypes.KeyringMembersWorklogDetails{
							EntityID: member.GetID(),
							Name:     name,
							OwnerIDs: member.KeyOwnerIDs(),
						},
					}
					item.CreateID(typ)
					missing[name] = item
				}
				d := missing[name].Details.(*apitypes.KeyringMembersWorklogDetails)

				path := *graph.GetKeyring().PathExp()
				for _, o := range d.Keyrings {
					if o.Equal(&path) {
						continue MembersLoop
					}
				}
				d.Keyrings = append(d.Keyrings, path)

				continue MembersLoop
			}
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
	orgID *identity.ID, item *apitypes.WorklogItem) error {

	// Preamble. Get the current user's keypairs, and the org's claims for
	// pubkey lookup.
	keypairs, err := h.engine.client.KeyPairs.List(ctx, orgID)
	if err != nil {
		return err
	}

	sigID, encID, kp, err := fetchKeyPairs(keypairs, orgID)
	if err != nil {
		return err
	}

	claimTree, err := h.engine.client.ClaimTree.Get(ctx, orgID, nil)
	if err != nil {
		return err
	}

	cgs := newCredentialGraphSet()
	details := item.Details.(*apitypes.KeyringMembersWorklogDetails)
	for _, pe := range details.Keyrings {
		graphs, err := h.engine.client.CredentialGraph.List(ctx, "", &pe, nil)
		if err != nil {
			return err
		}

		err = cgs.Add(graphs...)
		if err != nil {
			return err
		}
	}

	graphs, err := cgs.Active()
	if err != nil {
		return err
	}

	for _, graph := range graphs {
		for _, ownerID := range details.OwnerIDs {
			// See if the user/machine token in question is in this keyring
			m, _, err := graph.FindMember(&ownerID)
			if err != nil && err != registry.ErrMemberNotFound {
				return err
			}

			if m != nil {
				continue
			}

			// We have a missing user/machine token.
			// Get the existing user's share of the master encryption key, find the
			// missing user/machine token's public key, clone the existing
			// user's copy of the master encryption key for this keyring, and
			// post the result. Now the user/machine token is a member!
			krm, mekshare, err := graph.FindMember(h.engine.session.AuthID())

			if err != nil {
				return err
			}

			encPubKeySegment, err := claimTree.Find(krm.EncryptingKeyID, false)
			if err != nil {
				return err
			}
			encPubKey := encPubKeySegment.PublicKey

			targetPubKeySegment, err := claimTree.FindActive(&ownerID, primitive.EncryptionKeyType)
			if err != nil {
				return err
			}
			targetPubKey := targetPubKeySegment.PublicKey

			encMek, nonce, err := h.engine.crypto.CloneMembership(ctx,
				*mekshare.Key.Value, *mekshare.Key.Nonce, &kp.Encryption,
				*encPubKey.Body.Key.Value, *targetPubKey.Body.Key.Value)
			if err != nil {
				return err
			}

			key := &primitive.KeyringMemberKey{
				Algorithm: crypto.EasyBox,
				Nonce:     base64.New(nonce),
				Value:     base64.New(encMek),
			}

			keyring := graph.GetKeyring()
			switch k := keyring.(type) {
			case *envelope.KeyringV1:
				projectID := k.Body.ProjectID
				membership, err := newV1KeyringMember(ctx, h.engine.crypto, orgID, projectID,
					krm.KeyringID, &ownerID, targetPubKey.ID, encID, sigID, key, kp)
				if err != nil {
					return err
				}

				_, err = h.engine.client.KeyringMember.Post(ctx, []envelope.KeyringMemberV1{*membership})
				if err != nil {
					return err
				}

			case *envelope.Keyring:
				membership, err := newV2KeyringMember(ctx, h.engine.crypto, orgID, krm.KeyringID,
					&ownerID, targetPubKey.ID, encID, sigID, key, kp)
				if err != nil {
					return err
				}
				err = h.engine.client.Keyring.Members.Post(ctx, *membership)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
