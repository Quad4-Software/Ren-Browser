// SPDX-License-Identifier: MIT

//go:build !wasmbundle

package plugins

import (
	"encoding/json"
	"fmt"
)

type DevLogFunc func(level, message, detail string)

func (m *Manager) SetDevLogger(fn DevLogFunc) {
	m.mu.Lock()
	m.devLog = fn
	m.mu.Unlock()
}

func (m *Manager) LogPluginError(pluginID, phase, message, detail string) {
	m.logPlugin(pluginID, phase, "error", message, detail)
}

func (m *Manager) FailPlugin(pluginID, phase string, err error) error {
	if err == nil {
		return nil
	}
	m.mu.Lock()
	p, ok := m.plugins[pluginID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("plugin %q not found", pluginID)
	}
	msg := err.Error()
	p.Error = msg
	p.Enabled = false
	m.plugins[pluginID] = p
	manifest := p.Manifest
	m.disableLocked(pluginID)
	m.mu.Unlock()

	_ = m.store.SetPluginEnabled(pluginID, false)
	m.logPlugin(pluginID, phase, "error", msg, errDetail(err))
	m.emitPluginError(pluginID, phase, msg)
	m.emitUnloaded(manifest)
	return err
}

func (m *Manager) logPlugin(pluginID, phase, level, message, detail string) {
	label := fmt.Sprintf("plugin %s", pluginID)
	if phase != "" {
		label = fmt.Sprintf("plugin %s [%s]", pluginID, phase)
	}
	m.mu.RLock()
	fn := m.devLog
	m.mu.RUnlock()
	if fn != nil {
		fn(level, label+": "+message, detail)
	}
}

func (m *Manager) emitPluginError(pluginID, phase, message string) {
	m.mu.RLock()
	app := m.app
	m.mu.RUnlock()
	if app == nil {
		return
	}
	payload, _ := json.Marshal(map[string]string{
		"pluginId": pluginID,
		"phase":    phase,
		"message":  message,
	})
	app.Event.Emit("plugin:error", string(payload))
}

func errDetail(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (m *Manager) wrapBeforeFetch(pluginID string, hook BeforeFetchHook) BeforeFetchHook {
	return func(ctx FetchContext) (FetchHookResult, error) {
		res, err := hook(ctx)
		if err != nil {
			_ = m.FailPlugin(pluginID, "before_fetch", err)
		}
		return res, err
	}
}

func (m *Manager) wrapAfterFetch(pluginID string, hook AfterFetchHook) AfterFetchHook {
	return func(ctx FetchContext, body []byte) ([]byte, error) {
		out, err := hook(ctx, body)
		if err != nil {
			_ = m.FailPlugin(pluginID, "after_fetch", err)
			return body, err
		}
		return out, nil
	}
}
