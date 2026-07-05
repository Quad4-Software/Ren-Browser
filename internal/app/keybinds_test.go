// SPDX-License-Identifier: MIT
package app

import "testing"

func TestMergeKeybindsPreservesDefaults(t *testing.T) {
	merged := mergeKeybinds(KeybindSettings{
		Bindings: map[string]string{
			"reload": "mod+shift+r",
		},
	})
	if merged.Bindings["reload"] != "mod+shift+r" {
		t.Fatalf("reload = %q, want mod+shift+r", merged.Bindings["reload"])
	}
	if merged.Bindings["focusUrl"] != DefaultKeybinds().Bindings["focusUrl"] {
		t.Fatalf("focusUrl default not preserved")
	}
}

func TestDecodeKeybindsEmptyUsesDefaults(t *testing.T) {
	got, err := decodeKeybinds("")
	if err != nil {
		t.Fatal(err)
	}
	if got.Bindings["settings"] != "mod+," {
		t.Fatalf("settings = %q, want mod+,", got.Bindings["settings"])
	}
}

func TestEncodeDecodeKeybindsRoundTrip(t *testing.T) {
	want := DefaultKeybinds()
	want.Bindings["devtools"] = "mod+shift+j"
	raw, err := encodeKeybinds(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeKeybinds(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Bindings["devtools"] != "mod+shift+j" {
		t.Fatalf("devtools = %q, want mod+shift+j", got.Bindings["devtools"])
	}
}
