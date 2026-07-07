// SPDX-License-Identifier: MIT
package app

import "errors"

func (s *BrowserService) PickPluginZip() (string, error) {
	if s.app == nil {
		return "", errors.New("file picker unavailable in server mode")
	}
	result, err := s.app.Dialog.OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true).
		SetTitle("Select extension archive").
		AddFilter("RenBrowser extension", "*.renplugin.zip").
		AddFilter("Zip archive", "*.zip").
		PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (s *BrowserService) PickPluginDir() (string, error) {
	if s.app == nil {
		return "", errors.New("file picker unavailable in server mode")
	}
	result, err := s.app.Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		SetTitle("Select extension folder").
		PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (s *BrowserService) PickPluginWasm() (string, error) {
	if s.app == nil {
		return "", errors.New("file picker unavailable in server mode")
	}
	result, err := s.app.Dialog.OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true).
		SetTitle("Select extension module").
		AddFilter("RenBrowser extension module", "*.wasm").
		AddFilter("WebAssembly module", "*.wasm").
		PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	return result, nil
}
