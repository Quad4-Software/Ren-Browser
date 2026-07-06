// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"maps"
)

const keybindsSettingKey = "keybinds"

type KeybindSettings struct {
	Bindings map[string]string `json:"bindings"`
}

func DefaultKeybinds() KeybindSettings {
	return KeybindSettings{
		Bindings: map[string]string{
			"focusUrl":   "mod+l",
			"reload":     "mod+r",
			"devtools":   "mod+shift+i",
			"findInPage": "mod+f",
			"discovery":  "mod+shift+d",
			"settings":   "mod+,",
			"newTab":     "mod+t",
			"newWindow":  "mod+shift+n",
			"closeTab":   "mod+w",
			"fullscreen": "f11",
		},
	}
}

func mergeKeybinds(saved KeybindSettings) KeybindSettings {
	defaults := DefaultKeybinds()
	if saved.Bindings == nil {
		return defaults
	}
	out := KeybindSettings{Bindings: map[string]string{}}
	maps.Copy(out.Bindings, defaults.Bindings)
	for action, chord := range saved.Bindings {
		if chord != "" {
			out.Bindings[action] = chord
		}
	}
	return out
}

func encodeKeybinds(settings KeybindSettings) (string, error) {
	raw, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func decodeKeybinds(raw string) (KeybindSettings, error) {
	if raw == "" {
		return DefaultKeybinds(), nil
	}
	var settings KeybindSettings
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return DefaultKeybinds(), err
	}
	return mergeKeybinds(settings), nil
}
