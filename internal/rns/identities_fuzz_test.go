// SPDX-License-Identifier: MIT
package rns

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

func FuzzValidateIdentityID(f *testing.F) {
	valid, err := newIdentityID()
	if err != nil {
		f.Fatal(err)
	}
	f.Add(valid)
	f.Add("")
	f.Add("../escape")
	f.Add(strings.Repeat("z", identityIDLen))
	f.Fuzz(func(t *testing.T, id string) {
		err := validateIdentityID(id)
		if err == nil {
			if len(id) != identityIDLen {
				t.Fatalf("accepted wrong length %d", len(id))
			}
			if strings.ContainsAny(id, `/\`) || strings.Contains(id, "..") {
				t.Fatalf("accepted traversal id %q", id)
			}
			for _, c := range id {
				if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
					t.Fatalf("accepted non-hex id %q", id)
				}
			}
			return
		}
		if !errors.Is(err, ErrIdentityIDInvalid) {
			t.Fatalf("unexpected error %v", err)
		}
	})
}

func FuzzValidateIdentityName(f *testing.F) {
	f.Add("Default")
	f.Add("")
	f.Add(strings.Repeat("n", maxIdentityNameLen+1))
	f.Fuzz(func(t *testing.T, name string) {
		err := validateIdentityName(name)
		trimmed := strings.TrimSpace(name)
		switch {
		case trimmed == "":
			if !errors.Is(err, ErrIdentityNameEmpty) {
				t.Fatalf("name %q: err = %v", name, err)
			}
		case utf8.RuneCountInString(trimmed) > maxIdentityNameLen:
			if !errors.Is(err, ErrIdentityNameTooLong) {
				t.Fatalf("name %q: err = %v", name, err)
			}
		case err != nil:
			t.Fatalf("name %q: unexpected err = %v", name, err)
		}
	})
}

func FuzzParseIdentityRegistryJSON(f *testing.F) {
	f.Add([]byte(`{"version":1,"activeId":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","items":[{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","name":"Default","hash":"abc","createdAt":1}]}`))
	f.Add([]byte(`{"version":1,"activeId":"missing","items":[]}`))
	f.Add([]byte(`not json`))
	f.Fuzz(func(t *testing.T, raw []byte) {
		var data identityRegistryFile
		if err := json.Unmarshal(raw, &data); err != nil {
			return
		}
		reg := &IdentityRegistry{data: data}
		_ = reg.validateRegistryData()
	})
}

func FuzzValidateIdentityKeyData(f *testing.F) {
	f.Add(make([]byte, identityKeySize))
	f.Add([]byte("short"))
	f.Fuzz(func(t *testing.T, data []byte) {
		err := validateIdentityKeyData(data)
		if len(data) == identityKeySize {
			if err != nil {
				t.Fatalf("valid size rejected: %v", err)
			}
			return
		}
		if !errors.Is(err, ErrInvalidIdentityFile) {
			t.Fatalf("len=%d err=%v", len(data), err)
		}
	})
}
