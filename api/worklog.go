package api

import (
	"context"
	"net/url"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/identity"
)

// WorklogClient views and resolves worklog items in the daemon.
type WorklogClient struct {
	client *apiRoundTripper
}

// List returns the list of all worklog items in the given org.
func (w *WorklogClient) List(ctx context.Context, orgID *identity.ID) ([]apitypes.WorklogItem, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}

	var resp []apitypes.WorklogItem
	err := w.client.DaemonRoundTrip(ctx, "GET", "/worklog", v, nil, &resp, nil)
	return resp, err
}

// Get returns the worklog item with the given id in the given org.
func (w *WorklogClient) Get(ctx context.Context, orgID *identity.ID, ident *apitypes.WorklogID) (*apitypes.WorklogItem, error) {
	var res apitypes.WorklogItem
	err := w.singleItemWorker(ctx, "GET", orgID, ident, &res)
	return &res, err
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
