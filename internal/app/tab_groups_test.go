// SPDX-License-Identifier: MIT
package app_test

import (
	"strings"
	"testing"

	"renbrowser/internal/app"
)

func TestTabGroupsPersistAcrossReload(t *testing.T) {
	root := t.TempDir()
	svc := newTestServiceIn(t, root)

	want := []app.TabGroup{
		{ID: "g1", Name: "docs", Color: "#60a5fa", Collapsed: true},
		{ID: "g2", Name: "nodes", Color: "#34d399"},
	}
	svc.SetTabGroups(want)
	_ = svc.Store().Close()

	reloaded := newTestServiceIn(t, root)
	got := reloaded.GetTabGroups()
	if len(got) != 2 || got[0].ID != "g1" || got[0].Name != "docs" || !got[0].Collapsed {
		t.Fatalf("groups = %+v, want persisted g1/g2", got)
	}
}

func TestTabGroupsSanitize(t *testing.T) {
	svc := newTestService(t)

	longName := strings.Repeat("x", 80)
	got := svc.SetTabGroups([]app.TabGroup{
		{ID: "g1", Name: "  padded  ", Color: "#fff"},
		{ID: "g1", Name: "dupe", Color: "#000"},
		{ID: "  ", Name: "blank id"},
		{ID: "g2", Name: longName},
	})
	if len(got) != 2 {
		t.Fatalf("groups = %+v, want 2 entries", got)
	}
	if got[0].Name != "padded" {
		t.Fatalf("name = %q, want trimmed", got[0].Name)
	}
	if len(got[1].Name) != 48 {
		t.Fatalf("name len = %d, want 48", len(got[1].Name))
	}
}

func TestTabGroupsCappedAtMax(t *testing.T) {
	svc := newTestService(t)

	groups := make([]app.TabGroup, 0, 40)
	for i := range 40 {
		groups = append(groups, app.TabGroup{ID: strings.Repeat("g", 1) + strings.Repeat("0", i)})
	}
	got := svc.SetTabGroups(groups)
	if len(got) != 16 {
		t.Fatalf("groups = %d, want 16", len(got))
	}
}

func TestTabGroupsEmptyByDefault(t *testing.T) {
	svc := newTestService(t)

	if got := svc.GetTabGroups(); len(got) != 0 {
		t.Fatalf("groups = %+v, want empty", got)
	}
}
