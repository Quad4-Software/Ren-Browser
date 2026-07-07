// SPDX-License-Identifier: MIT
package rns

import (
	"errors"
	"fmt"
	"maps"

	"quad4/reticulum-go/pkg/common"
	"quad4/reticulum-go/pkg/debug"
	"quad4/reticulum-go/pkg/interfaces"
	"quad4/reticulum-go/pkg/reticulumconfig"
)

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func interfaceConfigsEqualForReload(a, b *common.InterfaceConfig) bool {
	a = EffectiveInterfaceConfig(a)
	b = EffectiveInterfaceConfig(b)
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Type == b.Type &&
		a.Enabled == b.Enabled &&
		a.Address == b.Address &&
		a.TargetHost == b.TargetHost &&
		a.TargetPort == b.TargetPort &&
		a.Port == b.Port &&
		a.KISSFraming == b.KISSFraming &&
		a.I2PTunneled == b.I2PTunneled &&
		a.I2PConnectable == b.I2PConnectable &&
		a.I2PSAMAddress == b.I2PSAMAddress &&
		sliceEqual(a.I2PPeers, b.I2PPeers) &&
		a.GroupID == b.GroupID &&
		a.DiscoveryScope == b.DiscoveryScope &&
		a.DiscoveryPort == b.DiscoveryPort &&
		a.DataPort == b.DataPort &&
		a.MulticastAddrType == b.MulticastAddrType &&
		a.Interface == b.Interface &&
		a.NetworkName == b.NetworkName &&
		a.Passphrase == b.Passphrase &&
		a.IFACSize == b.IFACSize &&
		a.IFACNetname == b.IFACNetname &&
		a.IFACNetkey == b.IFACNetkey
}

func (s *Stack) tearDownInterface(iface interfaces.Interface) {
	if iface == nil {
		return
	}
	name := iface.GetName()
	s.transport.UnregisterInterface(name)
	if err := iface.Stop(); err != nil {
		debug.Log(debug.DebugVerbose, "interface stop", "name", name, "error", err)
	}
}

// ReloadInterfaces reconciles running interfaces against cfg without restarting transport.
func (s *Stack) ReloadInterfaces(cfg *common.ReticulumConfig) error {
	if cfg == nil {
		return errors.New("nil config")
	}
	if s.transport == nil {
		return errors.New("nil transport")
	}

	oldCfg := s.cfg
	oldByName := make(map[string]interfaces.Interface, len(s.running))
	maps.Copy(oldByName, s.running)

	for name, oldI := range oldByName {
		ic, inNew := cfg.Interfaces[name]
		if !inNew || !ic.Enabled {
			s.tearDownInterface(oldI)
			delete(s.running, name)
			continue
		}
		if !interfaceConfigsEqualForReload(oldCfg.Interfaces[name], ic) {
			s.tearDownInterface(oldI)
			delete(s.running, name)
		}
	}

	for name, ic := range cfg.Interfaces {
		if !ic.Enabled {
			continue
		}
		if _, ok := s.running[name]; ok {
			continue
		}
		niface, err := interfaces.NewFromConfig(name, EffectiveInterfaceConfig(ic))
		if err != nil {
			if cfg.PanicOnInterfaceErr {
				return fmt.Errorf("interface %s: %w", name, err)
			}
			debug.Log(debug.DebugCritical, "ReloadInterfaces: skip interface", "name", name, "error", err)
			continue
		}
		if err := niface.Start(); err != nil {
			if cfg.PanicOnInterfaceErr {
				return fmt.Errorf("start %s: %w", name, err)
			}
			debug.Log(debug.DebugCritical, "ReloadInterfaces: start failed", "name", name, "error", err)
			continue
		}
		ni, ok := niface.(common.NetworkInterface)
		if !ok {
			_ = niface.Stop()
			return fmt.Errorf("interface %s does not implement common.NetworkInterface", name)
		}
		if err := s.transport.ReplaceInterface(name, ni); err != nil {
			_ = niface.Stop()
			if cfg.PanicOnInterfaceErr {
				return err
			}
			debug.Log(debug.DebugCritical, "ReloadInterfaces: ReplaceInterface failed", "name", name, "error", err)
			continue
		}
		s.running[name] = niface
	}

	s.cfg = cfg
	s.transport.SetReticulumConfig(cfg)
	return nil
}

func (s *Stack) setInterfaceEnabled(name string, enabled bool) error {
	if s.cfg == nil {
		return errConfigNotLoaded
	}
	cfg, ok := s.cfg.Interfaces[name]
	if !ok || cfg == nil {
		return errInterfaceNotFound
	}
	cfg.Enabled = enabled
	if err := reticulumconfig.SaveConfig(s.cfg); err != nil {
		return err
	}
	if !s.started {
		return nil
	}
	return s.ReloadInterfaces(s.cfg)
}
