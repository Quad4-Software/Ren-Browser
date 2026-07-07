// SPDX-License-Identifier: MIT
package rns

import (
	"errors"
	"sort"
)

var (
	errConfigNotLoaded   = errors.New("config not loaded")
	errInterfaceNotFound = errors.New("interface not found")
)

type Status struct {
	Online           bool   `json:"online"`
	NodeCount        int    `json:"nodeCount"`
	InterfaceCount   int    `json:"interfaceCount"`
	InterfacesOnline int    `json:"interfacesOnline"`
	ConfigPath       string `json:"configPath"`
}

type InterfaceInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
	Online  bool   `json:"online"`
	TxBytes uint64 `json:"txBytes"`
	RxBytes uint64 `json:"rxBytes"`
}

func (s *Stack) Status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st := Status{
		ConfigPath: s.ConfigPath(),
	}
	if s.handler != nil {
		st.NodeCount = len(s.handler.List())
	}
	if s.cfg != nil {
		st.InterfaceCount = len(s.cfg.Interfaces)
	}
	if s.transport != nil {
		for _, iface := range s.transport.GetInterfaces() {
			if iface != nil && iface.IsOnline() {
				st.InterfacesOnline++
			}
		}
	}
	st.Online = st.InterfacesOnline > 0
	return st
}

func (s *Stack) ListInterfaces() []InterfaceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]InterfaceInfo, 0)
	if s.cfg == nil {
		return out
	}

	online := map[string]bool{}
	stats := map[string]struct{ tx, rx uint64 }{}

	for name, iface := range s.running {
		if iface == nil {
			continue
		}
		online[name] = iface.IsOnline()
		stats[name] = struct{ tx, rx uint64 }{
			tx: iface.GetTxBytes(),
			rx: iface.GetRxBytes(),
		}
	}

	if s.transport != nil {
		for name, iface := range s.transport.GetInterfaces() {
			if iface == nil {
				continue
			}
			if _, tracked := stats[name]; tracked {
				continue
			}
			online[name] = iface.IsOnline()
			stats[name] = struct{ tx, rx uint64 }{
				tx: iface.GetTxBytes(),
				rx: iface.GetRxBytes(),
			}
		}
	}

	for name, cfg := range s.cfg.Interfaces {
		if cfg == nil {
			continue
		}
		effective := EffectiveInterfaceConfig(cfg) // FIXME(user1): show cfg.Type once BackboneClientInterface is vendored
		st := stats[name]
		out = append(out, InterfaceInfo{
			Name:    name,
			Type:    effective.Type,
			Enabled: cfg.Enabled,
			Online:  cfg.Enabled && online[name],
			TxBytes: st.tx,
			RxBytes: st.rx,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func (s *Stack) SetInterfaceEnabled(name string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.setInterfaceEnabled(name, enabled)
}
