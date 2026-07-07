// SPDX-License-Identifier: MIT
package app

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"renbrowser/internal/rns"
)

func TestCreateIdentityValidation(t *testing.T) {
	svc := newTestBrowserService(t)

	_, err := svc.CreateIdentity("   ")
	if !errors.Is(err, rns.ErrIdentityNameEmpty) {
		t.Fatalf("err = %v", err)
	}

	record, err := svc.CreateIdentity("Work")
	if err != nil {
		t.Fatal(err)
	}
	if record.Name != "Work" || record.Hash == "" {
		t.Fatalf("record = %+v", record)
	}
}

func TestSetActiveIdentityValidation(t *testing.T) {
	svc := newTestBrowserService(t)

	_, err := svc.SetActiveIdentity("")
	if !errors.Is(err, rns.ErrIdentityIDInvalid) {
		t.Fatalf("empty id err = %v", err)
	}

	items, err := svc.ListIdentities()
	if err != nil || len(items) == 0 {
		t.Fatalf("list err=%v len=%d", err, len(items))
	}
	active := items[0]
	for _, item := range items {
		if item.Active {
			active = item
		}
	}
	created, err := svc.CreateIdentity("Alt")
	if err != nil {
		t.Fatal(err)
	}
	switched, err := svc.SetActiveIdentity(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !switched.Active || switched.ID != created.ID {
		t.Fatalf("switched = %+v", switched)
	}
	_, err = svc.SetActiveIdentity(created.ID)
	if !errors.Is(err, rns.ErrIdentityAlreadyActive) {
		t.Fatalf("already active err = %v", err)
	}
	_, err = svc.SetActiveIdentity(active.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteIdentityValidation(t *testing.T) {
	svc := newTestBrowserService(t)

	if err := svc.DeleteIdentity(""); !errors.Is(err, rns.ErrIdentityIDInvalid) {
		t.Fatalf("empty id err = %v", err)
	}
	items, err := svc.ListIdentities()
	if err != nil || len(items) != 1 {
		t.Fatalf("list err=%v len=%d", err, len(items))
	}
	if err := svc.DeleteIdentity(items[0].ID); !errors.Is(err, rns.ErrCannotDeleteLast) {
		t.Fatalf("delete last err = %v", err)
	}
}

func TestRenameIdentityValidation(t *testing.T) {
	svc := newTestBrowserService(t)
	items, err := svc.ListIdentities()
	if err != nil || len(items) == 0 {
		t.Fatal(err)
	}
	_, err = svc.RenameIdentity(items[0].ID, strings.Repeat("x", 200))
	if !errors.Is(err, rns.ErrIdentityNameTooLong) {
		t.Fatalf("long name err = %v", err)
	}
}

func TestExportIdentityValidation(t *testing.T) {
	svc := newTestBrowserService(t)
	if err := svc.ExportIdentity(""); !errors.Is(err, rns.ErrIdentityIDInvalid) {
		t.Fatalf("empty id err = %v", err)
	}
}

func TestListIdentitiesWithoutStack(t *testing.T) {
	svc, err := NewBrowserServiceWithOptions(nil, nil, ServiceOptions{
		ProfilePath: filepath.Join(t.TempDir(), "profile.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = svc.Store().Close() })

	_, err = svc.ListIdentities()
	if err == nil || !strings.Contains(err.Error(), "reticulum not initialized") {
		t.Fatalf("err = %v", err)
	}
}

func TestIdentityOpErrorPreservesSentinel(t *testing.T) {
	err := identityOpError("create identity", rns.ErrIdentityNameEmpty)
	if !errors.Is(err, rns.ErrIdentityNameEmpty) {
		t.Fatalf("err = %v", err)
	}
	wrapped := identityOpError("create identity", errors.New("disk full"))
	if errors.Is(wrapped, rns.ErrIdentityNameEmpty) {
		t.Fatalf("unexpected sentinel in wrapped err: %v", wrapped)
	}
	if !strings.Contains(wrapped.Error(), "create identity") {
		t.Fatalf("wrapped = %q", wrapped.Error())
	}
}
