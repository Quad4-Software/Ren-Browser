// SPDX-License-Identifier: MIT
package plugins

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type SecurityFinding struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type SecurityAssessment struct {
	RiskLevel string            `json:"riskLevel"`
	Score     int               `json:"score"`
	Findings  []SecurityFinding `json:"findings"`
}

var (
	jsEvalRe         = regexp.MustCompile(`\beval\s*\(`)
	jsFunctionCtorRe = regexp.MustCompile(`\bnew\s+Function\s*\(`)
	jsAtobRe         = regexp.MustCompile(`\batob\s*\(`)
	jsDocWriteRe     = regexp.MustCompile(`\bdocument\.write\s*\(`)
)

func AssessExtension(manifest Manifest, dir string, embedded map[string][]byte, signature SignatureInfo) SecurityAssessment {
	var findings []SecurityFinding
	score := 0
	add := func(id, severity, message string, points int) {
		findings = append(findings, SecurityFinding{
			ID:       id,
			Severity: severity,
			Message:  message,
		})
		score += points
	}

	if strings.TrimSpace(manifest.ID) == "" {
		add("missing-id", "high", "extension manifest is missing an id", 40)
	}

	if signature.Present && !signature.Valid {
		add("invalid-signature", "high", "extension signature is present but invalid", 50)
	}

	if HasPermission(manifest, PermRenderUnsanitized) {
		add("unsanitized-render", "high", "extension can render unsanitized HTML", 35)
	}

	if HasPermission(manifest, PermNetworkFetch) && !signature.Valid {
		add("unsigned-network", "warn", "unsigned extension requests outbound network access", 20)
	}

	if HasPermission(manifest, PermNavigationWrite) && HasPermission(manifest, PermNetworkFetch) {
		add("nav-network-combo", "warn", "extension can navigate pages and make network requests", 15)
	}

	if len(manifest.Permissions) >= 6 {
		add("many-permissions", "warn", "extension requests many permissions", 10)
	}

	endpoints := CollectNetworkEndpoints(manifest, dir, embedded)
	if len(endpoints) > 15 {
		add("many-endpoints", "warn", "extension references many external network endpoints", 10)
	}
	if HasPermission(manifest, PermNetworkFetch) {
		undeclared := undeclaredNetworkEndpoints(manifest, endpoints)
		if len(undeclared) > 0 {
			add("undeclared-network", "warn", "extension code references network endpoints not declared in the manifest", 15)
		}
	}

	for _, pattern := range scanSuspiciousCode(manifest, dir, embedded) {
		add(pattern.id, pattern.severity, pattern.message, pattern.points)
	}

	risk := "low"
	switch {
	case score >= 50:
		risk = "high"
	case score >= 20:
		risk = "medium"
	}

	return SecurityAssessment{
		RiskLevel: risk,
		Score:     score,
		Findings:  findings,
	}
}

type suspiciousPattern struct {
	id       string
	severity string
	message  string
	points   int
}

func undeclaredNetworkEndpoints(manifest Manifest, endpoints []string) []string {
	declared := make(map[string]struct{})
	if manifest.Network != nil {
		for _, endpoint := range manifest.Network.Endpoints {
			for _, value := range extractURLsFromText(endpoint) {
				declared[normalizeEndpoint(value)] = struct{}{}
			}
			declared[normalizeEndpoint(endpoint)] = struct{}{}
		}
	}
	var out []string
	for _, endpoint := range endpoints {
		if !isHTTPURL(endpoint) {
			continue
		}
		if _, ok := declared[normalizeEndpoint(endpoint)]; ok {
			continue
		}
		out = append(out, endpoint)
	}
	return out
}

func scanSuspiciousCode(manifest Manifest, dir string, embedded map[string][]byte) []suspiciousPattern {
	seen := make(map[string]struct{})
	var out []suspiciousPattern
	record := func(id, severity, message string, points int) {
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, suspiciousPattern{id: id, severity: severity, message: message, points: points})
	}

	scanBytes := func(path string, data []byte) {
		if !shouldScanNetworkFile(path) {
			return
		}
		text := string(data)
		switch {
		case jsEvalRe.MatchString(text):
			record("js-eval", "high", "extension JavaScript uses eval()", 25)
		}
		if jsFunctionCtorRe.MatchString(text) {
			record("js-function-ctor", "warn", "extension JavaScript uses the Function constructor", 15)
		}
		if jsAtobRe.MatchString(text) {
			record("js-atob", "warn", "extension JavaScript uses atob() which may decode obfuscated payloads", 10)
		}
		if jsDocWriteRe.MatchString(text) {
			record("js-document-write", "warn", "extension JavaScript uses document.write()", 10)
		}
	}

	for path, data := range embedded {
		scanBytes(path, data)
	}
	if dir != "" {
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			data, readErr := os.ReadFile(path) // #nosec G304 G122 -- plugin dir from validated install path
			if readErr != nil {
				return nil
			}
			scanBytes(path, data)
			return nil
		})
	}
	_ = manifest
	return out
}
