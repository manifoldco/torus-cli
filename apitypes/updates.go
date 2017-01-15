package apitypes

type Updates struct {
	NeedsUpdate bool   `json:"needs_update"`
	Version     string `json:"version"`
}
