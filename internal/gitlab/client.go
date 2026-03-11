// Package gitlab provides a client for GitLab Package Registry and optionally Releases.
// internal/mcp and other packages use it as a single point for GitLab API access.
package gitlab

import "context"

// Package represents an MCP (or other) package in the registry.
type Package struct {
	Name    string
	Version string
	URL     string // artifact download URL
}

// Client is the GitLab API client for Package Registry and optionally Releases.
type Client struct {
	BaseURL string
	Token   string
}

// NewClient creates a GitLab client. baseURL is the registry/repository URL, token for auth.
func NewClient(baseURL, token string) *Client {
	return &Client{BaseURL: baseURL, Token: token}
}

// ListPackages returns packages from the project's package registry. Stub: returns nil, nil.
func (c *Client) ListPackages(ctx context.Context, projectID string) ([]Package, error) {
	_ = ctx
	_ = projectID
	// TODO: HTTP GET to GitLab API, parse response
	return nil, nil
}

// GetRelease returns release info for a tag. Optional, for GitLab Releases. Stub: returns nil, nil.
func (c *Client) GetRelease(ctx context.Context, projectID, tag string) (*Release, error) {
	_ = ctx
	_ = projectID
	_ = tag
	// TODO: GitLab Releases API
	return nil, nil
}

// Release holds GitLab release metadata (for future use).
type Release struct {
	TagName string
	Assets  []ReleaseAsset
}

// ReleaseAsset is a single asset (e.g. binary) in a release.
type ReleaseAsset struct {
	Name string
	URL  string
}
