// Package mcp handles MCP package updates from GitLab Package Registry.
// It uses internal/gitlab as the GitLab API client.
package mcp

import (
	"context"

	"github.com/ai-launcher/cli/internal/config"
	"github.com/ai-launcher/cli/internal/gitlab"
)

// UpdateMCP fetches package list from the registry and (later) installs/updates npm MCP packages.
// registryURL and token come from cfg or env; projectID can be passed when we support it.
func UpdateMCP(ctx context.Context, cfg *config.Config, registryURL, projectID string) error {
	if registryURL == "" {
		registryURL = cfg.MCPRegistry
	}
	token := "" // TODO: from cfg or env (e.g. GITLAB_TOKEN)
	client := gitlab.NewClient(registryURL, token)
	_, err := client.ListPackages(ctx, projectID)
	if err != nil {
		return err
	}
	// TODO: install/update npm packages from the list
	return nil
}
