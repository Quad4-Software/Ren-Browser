// SPDX-License-Identifier: MIT
package app

import (
	"errors"
	"log"

	"renbrowser/internal/rns"
)

// RestartReticulum stops the current Reticulum stack, loads a fresh one from the
// configuration file, starts it, and attaches it to the browser service.
func (s *BrowserService) RestartReticulum() error {
	s.mu.Lock()
	if s.shuttingDown {
		s.mu.Unlock()
		return errors.New("application is shutting down")
	}
	stack := s.stack
	s.mu.Unlock()

	if stack == nil {
		return errors.New("reticulum stack not initialized")
	}

	configPath := stack.ConfigPath()
	log.Printf("RestartReticulum: stopping current Reticulum stack (config: %s)...", configPath)

	// 1. Stop current stack
	if err := stack.Stop(); err != nil {
		log.Printf("RestartReticulum warning: failed to stop old stack: %v", err)
	}

	// 2. Create a new stack
	log.Println("RestartReticulum: initializing fresh Reticulum stack...")
	newStack, err := rns.NewStack(configPath)
	if err != nil {
		log.Printf("RestartReticulum error: failed to create new stack: %v", err)
		return err
	}

	// 3. Start new stack
	log.Println("RestartReticulum: starting fresh Reticulum stack...")
	if err := newStack.Start(); err != nil {
		log.Printf("RestartReticulum error: failed to start new stack: %v", err)
		return err
	}

	// 4. Attach new stack
	s.mu.Lock()
	s.stack = newStack
	s.mu.Unlock()

	s.bindPersistence()

	log.Println("RestartReticulum: Reticulum stack restarted successfully.")

	if s.app != nil {
		s.app.Event.Emit("rns:status", "online")
	}

	return nil
}
