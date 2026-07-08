// SPDX-License-Identifier: MIT
package rns

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"quad4/reticulum-go/pkg/common"
	"quad4/reticulum-go/pkg/debug"
	"quad4/reticulum-go/pkg/identity"
	"quad4/reticulum-go/pkg/interfaces"
	"quad4/reticulum-go/pkg/transport"

	"renbrowser/internal/nomadnet"
	"renbrowser/internal/paths"
)

type Stack struct {
	mu         sync.RWMutex
	cfg        *common.ReticulumConfig
	transport  *transport.Transport
	identity   *identity.Identity
	identities *IdentityRegistry
	handler    *nomadnet.AnnounceHandler
	browser    *nomadnet.Browser
	running    map[string]interfaces.Interface
	started    bool
}

func NewStack(configPath string) (*Stack, error) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}
	debug.SetDebugLevel(debug.DebugCritical)
	debug.Init()

	tr := transport.NewTransport(cfg)
	reg, ident, err := ensureTransportIdentity(tr, cfg)
	if err != nil {
		return nil, err
	}

	handler := nomadnet.NewAnnounceHandler()
	tr.RegisterAnnounceHandler(handler)

	return &Stack{
		cfg:        cfg,
		transport:  tr,
		identity:   ident,
		identities: reg,
		handler:    handler,
		browser:    nomadnet.NewBrowser(tr, handler),
		running:    make(map[string]interfaces.Interface),
	}, nil
}

func transportStorageDir(cfg *common.ReticulumConfig) string {
	if cfg != nil && cfg.ConfigPath != "" {
		return filepath.Join(filepath.Dir(cfg.ConfigPath), "storage")
	}
	return filepath.Join(DefaultConfigDir(), "storage")
}

func ensureTransportIdentity(tr *transport.Transport, cfg *common.ReticulumConfig) (*IdentityRegistry, *identity.Identity, error) {
	storageDir := transportStorageDir(cfg)
	reg, err := OpenIdentityRegistry(storageDir)
	if err != nil {
		return nil, nil, fmt.Errorf("identity registry: %w", err)
	}
	ident, err := reg.LoadActive()
	if err != nil {
		return nil, nil, fmt.Errorf("transport identity: %w", err)
	}
	tr.SetIdentity(ident)
	return reg, ident, nil
}

func (s *Stack) Identities() *IdentityRegistry {
	return s.identities
}

func (s *Stack) SwitchIdentity(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrIdentityIDInvalid
	}
	if s.identities == nil {
		return errors.New("identity registry not initialized")
	}
	ident, err := s.identities.SetActive(id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.identity = ident
	if s.transport != nil {
		s.transport.SetIdentity(ident)
	}
	s.mu.Unlock()
	return nil
}

func (s *Stack) IdentityHash() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.identity == nil {
		return ""
	}
	return s.identity.GetHexHash()
}

func (s *Stack) Identify(nodeHash string) error {
	s.mu.RLock()
	ident := s.identity
	s.mu.RUnlock()
	return s.browser.Identify(nodeHash, ident)
}

func (s *Stack) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}

	if err := s.transport.Start(); err != nil {
		return fmt.Errorf("transport start: %w", err)
	}

	for name, ifaceCfg := range s.cfg.Interfaces {
		if !ifaceCfg.Enabled {
			continue
		}
		iface, err := interfaces.NewFromConfig(name, EffectiveInterfaceConfig(ifaceCfg)) // FIXME(user1): drop EffectiveInterfaceConfig when BackboneClientInterface is vendored
		if err != nil {
			if s.cfg.PanicOnInterfaceErr {
				return fmt.Errorf("interface %s: %w", name, err)
			}
			continue
		}
		if err := iface.Start(); err != nil {
			if s.cfg.PanicOnInterfaceErr {
				return fmt.Errorf("start interface %s: %w", name, err)
			}
			continue
		}
		if err := s.transport.RegisterInterface(name, iface); err != nil {
			return fmt.Errorf("register interface %s: %w", name, err)
		}
		s.running[name] = iface
	}

	if err := s.transport.InitializePathRequestHandler(); err != nil {
		return fmt.Errorf("path handler: %w", err)
	}

	s.started = true
	return nil
}

func (s *Stack) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		return nil
	}
	if s.browser != nil {
		s.browser.Close()
	}
	err := s.transport.Close()
	s.running = make(map[string]interfaces.Interface)
	s.started = false
	return err
}

func (s *Stack) Handler() *nomadnet.AnnounceHandler {
	return s.handler
}

func (s *Stack) Browser() *nomadnet.Browser {
	return s.browser
}

func (s *Stack) Transport() *transport.Transport {
	return s.transport
}

func (s *Stack) ConfigPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cfg == nil {
		return ""
	}
	return s.cfg.ConfigPath
}

func (s *Stack) Config() *common.ReticulumConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *Stack) IsStarted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

func DefaultConfigDir() string {
	return filepath.Join(homeDir(), ".reticulum-go")
}

func homeDir() string {
	return paths.DataRoot()
}
