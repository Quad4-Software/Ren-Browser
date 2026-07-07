// SPDX-License-Identifier: MIT
package plugins

import (
	"time"
)

func (m *Manager) SetNetworkRecorder(fn PluginNetworkRecorder) {
	m.mu.Lock()
	m.networkRecorder = fn
	m.mu.Unlock()
}

func (m *Manager) PluginHTTPFetch(pluginID string, req WasmHTTPRequest) (WasmHTTPResponse, error) {
	if err := consumeWasmFetchBudget(pluginID); err != nil {
		return WasmHTTPResponse{}, err
	}
	start := time.Now()
	resp, err := DoPluginHTTP(req)
	m.recordNetworkFetch(pluginID, req, resp, time.Since(start), err)
	return resp, err
}

func (m *Manager) recordNetworkFetch(pluginID string, req WasmHTTPRequest, resp WasmHTTPResponse, duration time.Duration, err error) {
	m.mu.RLock()
	fn := m.networkRecorder
	m.mu.RUnlock()
	if fn == nil {
		return
	}
	statusCode := resp.StatusCode
	bytes := len(resp.Body)
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
		statusCode = 0
		bytes = 0
	} else if resp.Error != "" {
		errMsg = resp.Error
	}
	fn(pluginID, req.Method, req.URL, statusCode, bytes, duration.Milliseconds(), errMsg)
}

func (m *Manager) loadRuntimeSettings(id string) RuntimeSettings {
	row, err := m.store.GetPlugin(id)
	if err != nil {
		return RuntimeSettings{}
	}
	return ParseRuntimeSettings(row.SettingsJSON)
}

func (m *Manager) saveRuntimeSettings(id string, enabled bool, settings RuntimeSettings) error {
	return m.store.UpsertPlugin(id, enabled, settings.JSON())
}

func (m *Manager) establishIntegrityHash(dir string, settings RuntimeSettings) (RuntimeSettings, error) {
	hash, err := ComputeDirIntegrityHash(dir)
	if err != nil {
		return settings, err
	}
	settings.IntegrityHash = hash
	settings.Tampered = false
	return settings, nil
}

func (m *Manager) verifyIntegrityOnLoad(id, dir string, settings RuntimeSettings) (RuntimeSettings, bool, string) {
	hash, err := ComputeDirIntegrityHash(dir)
	if err != nil {
		return settings, false, err.Error()
	}
	if settings.IntegrityHash == "" {
		settings.IntegrityHash = hash
		settings.Tampered = false
		return settings, true, ""
	}
	if hash != settings.IntegrityHash {
		settings.Tampered = true
		return settings, false, integrityTamperMessage
	}
	settings.Tampered = false
	return settings, true, ""
}
