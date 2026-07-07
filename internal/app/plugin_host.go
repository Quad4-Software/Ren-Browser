// SPDX-License-Identifier: MIT
package app

import (
	"fmt"
	"strings"

	"renbrowser/internal/plugins"
)

type PluginSummary struct {
	ID                 string                     `json:"id"`
	Name               string                     `json:"name"`
	Version            string                     `json:"version"`
	Description        string                     `json:"description,omitempty"`
	Enabled            bool                       `json:"enabled"`
	Error              string                     `json:"error,omitempty"`
	SizeBytes          int64                      `json:"sizeBytes"`
	Signature          plugins.SignatureInfo      `json:"signature"`
	Security           plugins.SecurityAssessment `json:"security"`
	Tampered           bool                       `json:"tampered"`
	GrantedPermissions []string                   `json:"grantedPermissions"`
}

type PluginInstallPreview struct {
	ID                   string                     `json:"id"`
	Name                 string                     `json:"name"`
	Version              string                     `json:"version"`
	Description          string                     `json:"description,omitempty"`
	Permissions          []string                   `json:"permissions"`
	NetworkEndpoints     []string                   `json:"networkEndpoints"`
	RequiresNetworkFetch bool                       `json:"requiresNetworkFetch"`
	Signature            plugins.SignatureInfo      `json:"signature"`
	Security             plugins.SecurityAssessment `json:"security"`
	I18nLocales          []string                   `json:"i18nLocales"`
}

type PluginHost struct {
	manager *plugins.Manager
}

func NewPluginHost(manager *plugins.Manager) *PluginHost {
	return &PluginHost{manager: manager}
}

func (h *PluginHost) ListPlugins() []PluginSummary {
	if h.manager == nil {
		return nil
	}
	installed := h.manager.List()
	out := make([]PluginSummary, 0, len(installed))
	for _, p := range installed {
		signature := plugins.VerifyDirSignature(p.Dir)
		security := plugins.AssessExtension(p.Manifest, p.Dir, nil, signature)
		summary := PluginSummary{
			ID:                 p.Manifest.ID,
			Name:               p.Manifest.Name,
			Version:            p.Manifest.Version,
			Enabled:            p.Enabled,
			Error:              p.Error,
			Signature:          signature,
			Security:           security,
			Tampered:           p.Tampered,
			GrantedPermissions: p.GrantedPermissions,
		}
		if summary.ID == "" {
			summary.ID = p.Dir
		}
		if p.Manifest.Name != "" {
			summary.Description = p.Manifest.Description
		}
		if size, err := plugins.DirSizeBytes(p.Dir); err == nil {
			summary.SizeBytes = size
		}
		out = append(out, summary)
	}
	return out
}

func (h *PluginHost) GetPlugin(id string) (plugins.InstalledPlugin, error) {
	if h.manager == nil {
		return plugins.InstalledPlugin{}, fmt.Errorf("plugin host unavailable")
	}
	p, ok := h.manager.Get(id)
	if !ok {
		return plugins.InstalledPlugin{}, fmt.Errorf("plugin %q not found", id)
	}
	return p, nil
}

func (h *PluginHost) EnablePlugin(id string) error {
	if h.manager == nil {
		return fmt.Errorf("plugin host unavailable")
	}
	return h.manager.Enable(id)
}

func (h *PluginHost) DisablePlugin(id string) error {
	if h.manager == nil {
		return fmt.Errorf("plugin host unavailable")
	}
	return h.manager.Disable(id)
}

func pluginInstallPreviewFrom(p plugins.InstallPreview) PluginInstallPreview {
	m := p.Manifest
	return PluginInstallPreview{
		ID:                   m.ID,
		Name:                 m.Name,
		Version:              m.Version,
		Description:          m.Description,
		Permissions:          m.Permissions,
		NetworkEndpoints:     p.NetworkEndpoints,
		RequiresNetworkFetch: p.RequiresNetworkFetch,
		Signature:            p.Signature,
		Security:             p.Security,
		I18nLocales:          p.I18nLocales,
	}
}

func (h *PluginHost) PreviewPluginInstallFromZip(path string) (PluginInstallPreview, error) {
	if h.manager == nil {
		return PluginInstallPreview{}, fmt.Errorf("plugin host unavailable")
	}
	preview, err := plugins.PreviewInstallFromZip(path)
	if err != nil {
		return PluginInstallPreview{}, err
	}
	return pluginInstallPreviewFrom(preview), nil
}

func (h *PluginHost) PreviewPluginInstallFromDir(path string) (PluginInstallPreview, error) {
	if h.manager == nil {
		return PluginInstallPreview{}, fmt.Errorf("plugin host unavailable")
	}
	preview, err := plugins.PreviewInstallFromDir(path)
	if err != nil {
		return PluginInstallPreview{}, err
	}
	return pluginInstallPreviewFrom(preview), nil
}

func (h *PluginHost) PreviewPluginInstallFromWasm(path string) (PluginInstallPreview, error) {
	if h.manager == nil {
		return PluginInstallPreview{}, fmt.Errorf("plugin host unavailable")
	}
	preview, err := plugins.PreviewInstallFromWasm(path)
	if err != nil {
		return PluginInstallPreview{}, err
	}
	return pluginInstallPreviewFrom(preview), nil
}

func (h *PluginHost) InstallPluginFromZip(path string, granted []string) (plugins.InstalledPlugin, error) {
	if h.manager == nil {
		return plugins.InstalledPlugin{}, fmt.Errorf("plugin host unavailable")
	}
	return h.manager.InstallFromZip(path, granted)
}

func (h *PluginHost) InstallPluginFromDir(path string, granted []string) (plugins.InstalledPlugin, error) {
	if h.manager == nil {
		return plugins.InstalledPlugin{}, fmt.Errorf("plugin host unavailable")
	}
	return h.manager.InstallFromDir(path, granted)
}

func (h *PluginHost) InstallPluginFromWasm(path string, granted []string) (plugins.InstalledPlugin, error) {
	if h.manager == nil {
		return plugins.InstalledPlugin{}, fmt.Errorf("plugin host unavailable")
	}
	return h.manager.InstallFromWasm(path, granted)
}

func (h *PluginHost) AddTrustedPublisher(identity, name string) error {
	if strings.TrimSpace(identity) == "" {
		return fmt.Errorf("publisher identity is required")
	}
	if h.manager == nil {
		return fmt.Errorf("plugin host unavailable")
	}
	return plugins.AddUserTrustedPublisherWithStore(h.manager.Store(), identity, name)
}

func (h *PluginHost) UninstallPlugin(id string) error {
	if h.manager == nil {
		return fmt.Errorf("plugin host unavailable")
	}
	return h.manager.Uninstall(id)
}

func (h *PluginHost) GetPluginStorage(id, key string) (string, error) {
	if h.manager == nil {
		return "", fmt.Errorf("plugin host unavailable")
	}
	p, ok := h.manager.Get(id)
	if !ok || !p.Enabled {
		return "", fmt.Errorf("plugin %q not enabled", id)
	}
	if err := plugins.RequireGrantedPermission(p.GrantedPermissions, p.Manifest, plugins.PermStoragePlugin); err != nil {
		return "", err
	}
	return h.manager.Storage().Get(id, key)
}

func (h *PluginHost) SetPluginStorage(id, key, value string) error {
	if h.manager == nil {
		return fmt.Errorf("plugin host unavailable")
	}
	p, ok := h.manager.Get(id)
	if !ok || !p.Enabled {
		return fmt.Errorf("plugin %q not enabled", id)
	}
	if err := plugins.RequireGrantedPermission(p.GrantedPermissions, p.Manifest, plugins.PermStoragePlugin); err != nil {
		return err
	}
	return h.manager.Storage().Set(id, key, value)
}

func (h *PluginHost) InvokeCommand(pluginID, commandID string, args map[string]string) error {
	if h.manager == nil {
		return fmt.Errorf("plugin host unavailable")
	}
	return h.manager.InvokeCommand(pluginID, commandID, args)
}

func (h *PluginHost) GetContributions() plugins.ContributionsView {
	if h.manager == nil {
		return plugins.ContributionsView{}
	}
	return h.manager.Registry().Contributions()
}

func (h *PluginHost) PluginsDir() string {
	if h.manager == nil {
		return ""
	}
	return h.manager.PluginsDir()
}

func (h *PluginHost) EmitPluginEvent(pluginID, event string, data any) error {
	if h.manager == nil {
		return fmt.Errorf("plugin host unavailable")
	}
	p, ok := h.manager.Get(pluginID)
	if !ok || !p.Enabled {
		return fmt.Errorf("plugin %q not enabled", pluginID)
	}
	if err := plugins.RequireGrantedPermission(p.GrantedPermissions, p.Manifest, plugins.PermEventsEmit); err != nil {
		return err
	}
	h.manager.EmitEvent(pluginID, event, data)
	return nil
}
