// SPDX-License-Identifier: MIT
package rns

import (
	"errors"
	"sort"

	"quad4/reticulum-go/pkg/sharedinstance"
)

var (
	errConfigNotLoaded   = errors.New("config not loaded")
	errInterfaceNotFound = errors.New("interface not found")
)

type Status struct {
	Online                    bool   `json:"online"`
	NodeCount                 int    `json:"nodeCount"`
	InterfaceCount            int    `json:"interfaceCount"`
	InterfacesOnline          int    `json:"interfacesOnline"`
	ConfigPath                string `json:"configPath"`
	EnableTransport           bool   `json:"enableTransport"`
	ShareInstance             bool   `json:"shareInstance"`
	ConnectedToSharedInstance bool   `json:"connectedToSharedInstance"`
	SharedInstanceMode        string `json:"sharedInstanceMode"`
	TransportActive           bool   `json:"transportActive"`
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
		ConfigPath:         s.ConfigPathUnlocked(),
		SharedInstanceMode: "disabled",
	}
	if s.cfg != nil {
		st.EnableTransport = s.cfg.EnableTransport
		st.ShareInstance = s.cfg.ShareInstance
		st.ConnectedToSharedInstance = s.cfg.ConnectedToSharedInstance
		st.InterfaceCount = len(s.cfg.Interfaces)
	}
	if s.transport != nil && s.transport.ConnectedToSharedInstance() {
		st.ConnectedToSharedInstance = true
	}
	if s.sharedInstance != nil {
		switch s.sharedInstance.Mode {
		case sharedinstance.ModeServer:
			st.SharedInstanceMode = "server"
		case sharedinstance.ModeClient:
			st.SharedInstanceMode = "client"
			st.ConnectedToSharedInstance = true
		}
	}
	st.TransportActive = st.EnableTransport && !st.ConnectedToSharedInstance
	if s.handler != nil {
		st.NodeCount = len(s.handler.List())
	}
	if s.transport != nil {
		for _, iface := range s.transport.GetInterfaces() {
			if iface != nil && iface.IsOnline() {
				st.InterfacesOnline++
			}
		}
	}
	st.Online = st.InterfacesOnline > 0 || st.ConnectedToSharedInstance
	return st
}

func (s *Stack) ConfigPathUnlocked() string {
	if s.cfg == nil {
		return ""
	}
	return s.cfg.ConfigPath
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
		st := stats[name]
		out = append(out, InterfaceInfo{
			Name:    name,
			Type:    cfg.Type,
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

func (s *Stack) DeleteInterface(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deleteInterface(name)
}
