package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/identity"
)

// WorklogClient views and resolves worklog items in the daemon.
type WorklogClient struct {
	client *apiRoundTripper
}

var errUnknownWorklogType = errors.New("Unknown worklog item type")

// List returns the list of all worklog items in the given org.
func (w *WorklogClient) List(ctx context.Context, orgID *identity.ID) ([]apitypes.WorklogItem, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}

	var resp []rawWorklogItem
	err := w.client.DaemonRoundTrip(ctx, "GET", "/worklog", v, nil, &resp, nil)
	if err != nil {
		return nil, err
	}

	out := make([]apitypes.WorklogItem, 0, len(resp))
	for _, w := range resp {
		err := w.setDetails()
		if err != nil {
			return nil, err
		}

		out = append(out, *w.WorklogItem)
	}
	return out, err
}

// Get returns the worklog item with the given id in the given org.
func (w *WorklogClient) Get(ctx context.Context, orgID *identity.ID, ident *apitypes.WorklogID) (*apitypes.WorklogItem, error) {
	var res rawWorklogItem
	err := w.singleItemWorker(ctx, "GET", orgID, ident, &res)
	if err != nil {
		return nil, err
	}

	err = res.setDetails()
	return res.WorklogItem, err
}

// Resolve resolves the worklog item with the given id in the given org.
func (w *WorklogClient) Resolve(ctx context.Context, orgID *identity.ID, ident *apitypes.WorklogID) (*apitypes.WorklogResult, error) {
	var res apitypes.WorklogResult
	err := w.singleItemWorker(ctx, "POST", orgID, ident, &res)
	return &res, err
}

func (w *WorklogClient) singleItemWorker(ctx context.Context, verb string, orgID *identity.ID, ident *apitypes.WorklogID, res interface{}) error {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}

	return w.client.DaemonRoundTrip(ctx, verb, "/worklog/"+ident.String(), v, nil, res, nil)
}

type rawWorklogItem struct {
	*apitypes.WorklogItem
	Details json.RawMessage `json:"details"`
}

func (w *rawWorklogItem) setDetails() error {
	if w == nil {
		return nil
	}

	switch w.Type() {
	case apitypes.SecretRotateWorklogType:
		w.WorklogItem.Details = &apitypes.SecretRotateWorklogDetails{}
	case apitypes.MissingKeypairsWorklogType:
		w.WorklogItem.Details = &apitypes.MissingKeypairsWorklogDetails{}
	case apitypes.InviteApproveWorklogType:
		w.WorklogItem.Details = &apitypes.InviteApproveWorklogDetails{}
	case apitypes.UserKeyringMembersWorklogType:
		fallthrough
	case apitypes.MachineKeyringMembersWorklogType:
		w.WorklogItem.Details = &apitypes.KeyringMembersWorklogDetails{}
	default:
		return errUnknownWorklogType
	}

	return json.Unmarshal(w.Details, w.WorklogItem.Details)
}
