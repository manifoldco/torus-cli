package logic

import (
	"context"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/envelope"
	"github.com/manifoldco/torus-cli/identity"
	"github.com/manifoldco/torus-cli/primitive"

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
		},
	}

	return w
}

type worklogTypeHandler interface {
	list(context.Context, *envelope.Unsigned) ([]apitypes.WorklogItem, error)
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

func (h *secretRotateHandler) list(ctx context.Context, org *envelope.Unsigned) ([]apitypes.WorklogItem, error) {
	projects, err := h.engine.client.Projects.List(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	cgs := newCredentialGraphSet()
	orgName := org.Body.(*primitive.Org).Name
	for _, project := range projects {
		projName := project.Body.(*primitive.Project).Name
		graphs, err := h.engine.client.CredentialGraph.Search(ctx,
			"/"+orgName+"/"+projName+"/*/*/*/*", h.engine.session.AuthID())
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
		base, err := baseCredential(&cred)
		if err != nil {
			return nil, err
		}

		item := apitypes.WorklogItem{
			Subject: base.PathExp.String() + "/" + base.Name,
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

func (h *missingKeypairsHandler) list(ctx context.Context, org *envelope.Unsigned) ([]apitypes.WorklogItem, error) {
	encClaimed, sigClaimed, err := fetchRegistryKeyPairs(ctx, h.engine.client, org.ID)
	if err != nil {
		return nil, err
	}

	var items []apitypes.WorklogItem
	if encClaimed == nil || sigClaimed == nil {
		item := apitypes.WorklogItem{
			Subject: org.Body.(*primitive.Org).Name,
			Summary: "Signing and Encryption keypairs missing for org.",
		}

		if encClaimed != nil {
			item.Summary = "Signing keypair missing for org."
		} else if sigClaimed != nil {
			item.Summary = "Encryption keypair missing for org."
		}

		item.CreateID(apitypes.MissingKeypairsWorklogType)

		items = append(items, item)
	}

	return items, nil
}

func (h *missingKeypairsHandler) resolve(ctx context.Context, n *observer.Notifier,
	orgID *identity.ID, item *apitypes.WorklogItem) (*apitypes.WorklogResult, error) {
	err := h.engine.GenerateKeypair(ctx, n, orgID)
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

func (h *inviteApproveHandler) list(ctx context.Context, org *envelope.Unsigned) ([]apitypes.WorklogItem, error) {
	invites, err := h.engine.client.OrgInvite.List(ctx, org.ID, []string{"accepted"}, "")
	if err != nil {
		return nil, err
	}

	var items []apitypes.WorklogItem
	for _, invite := range invites {
		item := apitypes.WorklogItem{
			Subject:   invite.Body.(*primitive.OrgInvite).Email,
			Summary:   "Org invite ready for approval",
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
