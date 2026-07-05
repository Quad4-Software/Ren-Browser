// SPDX-License-Identifier: MIT
package app

import (
	"fmt"

	"renbrowser/internal/plugins"
)

type PluginSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
	Error       string `json:"error,omitempty"`
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
		summary := PluginSummary{
			ID:      p.Manifest.ID,
			Name:    p.Manifest.Name,
			Version: p.Manifest.Version,
			Enabled: p.Enabled,
			Error:   p.Error,
		}
		if summary.ID == "" {
			summary.ID = p.Dir
		}
		if p.Manifest.Name != "" {
			summary.Description = p.Manifest.Description
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

func (h *PluginHost) InstallPluginFromZip(path string) (plugins.InstalledPlugin, error) {
	if h.manager == nil {
		return plugins.InstalledPlugin{}, fmt.Errorf("plugin host unavailable")
	}
	return h.manager.InstallFromZip(path)
}

func (h *PluginHost) InstallPluginFromDir(path string) (plugins.InstalledPlugin, error) {
	if h.manager == nil {
		return plugins.InstalledPlugin{}, fmt.Errorf("plugin host unavailable")
	}
	return h.manager.InstallFromDir(path)
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
	if err := plugins.RequirePermission(p.Manifest, plugins.PermStoragePlugin); err != nil {
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
	if err := plugins.RequirePermission(p.Manifest, plugins.PermStoragePlugin); err != nil {
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
	if err := plugins.RequirePermission(p.Manifest, plugins.PermEventsEmit); err != nil {
		return err
	}
	h.manager.EmitEvent(pluginID, event, data)
	return nil
}
