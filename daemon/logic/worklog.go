package logic

import (
	"context"

	"github.com/manifoldco/torus-cli/apitypes"
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
	engine *Engine
}

// List returns the list of all outstanding worklog items for the given org
func (w *Worklog) List(ctx context.Context, orgID *identity.ID) ([]apitypes.WorklogItem, error) {
	var items []apitypes.WorklogItem

	org, err := w.engine.client.Orgs.Get(ctx, orgID)
	if err != nil {
		return nil, err
	}

	projects, err := w.engine.client.Projects.List(ctx, orgID)
	if err != nil {
		return nil, err
	}

	cgs := newCredentialGraphSet()
	orgName := org.Body.(*primitive.Org).Name
	for _, project := range projects {
		projName := project.Body.(*primitive.Project).Name
		graphs, err := w.engine.client.CredentialGraph.Search(ctx,
			"/"+orgName+"/"+projName+"/*/*/*/*", w.engine.session.AuthID())
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

	encClaimed, sigClaimed, err := fetchRegistryKeyPairs(ctx, w.engine.client, org.ID)
	if err != nil {
		return nil, err
	}

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

	invites, err := w.engine.client.OrgInvite.List(ctx, org.ID, []string{"accepted"}, "")
	if err != nil {
		return nil, err
	}

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

// Get returns a single worklog item for the given org with the given ident.
func (w *Worklog) Get(ctx context.Context, orgID *identity.ID,
	ident *apitypes.WorklogID) (*apitypes.WorklogItem, error) {

	items, err := w.List(ctx, orgID)
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

	switch item.Type() {
	case apitypes.SecretRotateWorklogType:
		return &apitypes.WorklogResult{
			ID:      item.ID,
			State:   apitypes.ManualWorklogResult,
			Message: "Please set a new value for the secret at " + item.Subject,
		}, nil
	case apitypes.MissingKeypairsWorklogType:
		err = w.engine.GenerateKeypair(ctx, n, orgID)
		if err != nil {
			return &apitypes.WorklogResult{
				ID:      item.ID,
				State:   apitypes.ErrorWorklogResult,
				Message: "Error generating keypairs: " + err.Error(),
			}, nil
		}

		return &apitypes.WorklogResult{
			ID:      item.ID,
			State:   apitypes.SuccessWorklogResult,
			Message: "Keypairs generated.",
		}, nil
	case apitypes.InviteApproveWorklogType:
		_, err = w.engine.ApproveInvite(ctx, n, item.SubjectID)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return &apitypes.WorklogResult{
				ID:      item.ID,
				State:   apitypes.ErrorWorklogResult,
				Message: "Error approving invite: " + err.Error(),
			}, nil
		}

		return &apitypes.WorklogResult{
			ID:      item.ID,
			State:   apitypes.SuccessWorklogResult,
			Message: "User invite approved and finalized.",
		}, nil

	}

	return nil, nil
}
