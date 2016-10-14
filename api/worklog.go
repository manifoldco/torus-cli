package api

import (
	"context"
	"net/url"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/identity"
)

// WorklogClient views and resolves worklog items in the daemon.
type WorklogClient struct {
	client *Client
}

// List returns the list of all worklog items in the given org.
func (w *WorklogClient) List(ctx context.Context, orgID *identity.ID) ([]apitypes.WorklogItem, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}

	req, _, err := w.client.NewRequest("GET", "/worklog", v, nil, false)
	if err != nil {
		return nil, err
	}

	var resp []apitypes.WorklogItem

	_, err = w.client.Do(ctx, req, &resp, nil, nil)
	return resp, err
}

// Get returns the worklog item with the given id in the given org.
func (w *WorklogClient) Get(ctx context.Context, orgID *identity.ID, ident *apitypes.WorklogID) (*apitypes.WorklogItem, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}

	req, _, err := w.client.NewRequest("GET", "/worklog/"+ident.String(), v, nil, false)
	if err != nil {
		return nil, err
	}

	var entry apitypes.WorklogItem

	_, err = w.client.Do(ctx, req, &entry, nil, nil)
	return &entry, err
}

// Resolve resolves the worklog item with the given id in the given org.
func (w *WorklogClient) Resolve(ctx context.Context, orgID *identity.ID, ident *apitypes.WorklogID) (*apitypes.WorklogResult, error) {
	v := &url.Values{}
	if orgID != nil {
		v.Set("org_id", orgID.String())
	}

	req, _, err := w.client.NewRequest("POST", "/worklog/"+ident.String(), v, nil, false)
	if err != nil {
		return nil, err
	}

	var res apitypes.WorklogResult

	_, err = w.client.Do(ctx, req, &res, nil, nil)
	return &res, nil
}
