// SPDX-License-Identifier: MIT
package plugins

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"renbrowser/internal/db"
)

func TestFailPluginDisablesAndLogs(t *testing.T) {
	tmp := t.TempDir()
	database, err := db.Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	manager := NewManager(database)
	manager.SetPluginsDirForTest(filepath.Join(tmp, "plugins"))
	var logs []string
	manager.SetDevLogger(func(level, message, detail string) {
		logs = append(logs, level+":"+message)
	})

	src := filepath.Join("testdata", "hello-extension")
	installed, err := manager.InstallFromDir(src, nil)
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if !installed.Enabled {
		t.Fatal("expected plugin enabled after install")
	}

	err = manager.FailPlugin(installed.Manifest.ID, "test", errors.New("boom"))
	if err == nil {
		t.Fatal("expected error")
	}

	p, ok := manager.Get(installed.Manifest.ID)
	if !ok {
		t.Fatal("plugin missing")
	}
	if p.Enabled {
		t.Fatal("expected plugin disabled")
	}
	if p.Error != "boom" {
		t.Fatalf("error = %q", p.Error)
	}
	if len(logs) == 0 || !strings.Contains(logs[0], "boom") {
		t.Fatalf("logs = %v", logs)
	}
}
