// SPDX-License-Identifier: MIT
package contract_test

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"renbrowser/internal/app"
)

var exportFnRe = regexp.MustCompile(`(?m)^export function (\w+)\(`)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func parseBindingExports(t *testing.T, relPath string) map[string]struct{} {
	t.Helper()
	path := filepath.Join(repoRoot(t), relPath)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", relPath, err)
	}
	out := make(map[string]struct{})
	for _, match := range exportFnRe.FindAllStringSubmatch(string(raw), -1) {
		out[match[1]] = struct{}{}
	}
	return out
}

func exportedMethods(t *testing.T, ptr any) map[string]struct{} {
	t.Helper()
	rt := reflect.TypeOf(ptr)
	out := make(map[string]struct{})
	for i := 0; i < rt.NumMethod(); i++ {
		name := rt.Method(i).Name
		if name == "" || name[0] < 'A' || name[0] > 'Z' {
			continue
		}
		out[name] = struct{}{}
	}
	return out
}

func diffSets(want, got map[string]struct{}) (missing, extra []string) {
	for name := range want {
		if _, ok := got[name]; !ok {
			missing = append(missing, name)
		}
	}
	for name := range got {
		if _, ok := want[name]; !ok {
			extra = append(extra, name)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	return missing, extra
}

func TestBrowserServiceBindingDrift(t *testing.T) {
	bindings := parseBindingExports(t, "frontend/bindings/renbrowser/internal/app/browserservice.ts")
	methods := exportedMethods(t, (*app.BrowserService)(nil))
	missing, extra := diffSets(methods, bindings)
	if len(missing) > 0 || len(extra) > 0 {
		t.Fatalf(
			"BrowserService binding drift\nmissing in bindings: %s\nextra in bindings: %s\nregenerate with: wails3 generate bindings",
			strings.Join(missing, ", "),
			strings.Join(extra, ", "),
		)
	}
}

func TestPluginHostBindingDrift(t *testing.T) {
	bindings := parseBindingExports(t, "frontend/bindings/renbrowser/internal/app/pluginhost.ts")
	methods := exportedMethods(t, (*app.PluginHost)(nil))
	missing, extra := diffSets(methods, bindings)
	if len(missing) > 0 || len(extra) > 0 {
		t.Fatalf(
			"PluginHost binding drift\nmissing in bindings: %s\nextra in bindings: %s\nregenerate with: wails3 generate bindings",
			strings.Join(missing, ", "),
			strings.Join(extra, ", "),
		)
	}
}
