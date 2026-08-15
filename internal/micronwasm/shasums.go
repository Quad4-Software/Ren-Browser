// SPDX-License-Identifier: MIT
package micronwasm

import (
	"fmt"
	"regexp"
	"strings"

	"renbrowser/internal/constants"
)

var shasumLine = regexp.MustCompile(`^([a-fA-F0-9]{64})\s+\*?(\S+)\s*$`)

func ParseShasums256ForFilename(text, filename string) (string, error) {
	if strings.TrimSpace(text) == "" || strings.TrimSpace(filename) == "" {
		return "", fmt.Errorf("empty shasums input")
	}
	for raw := range strings.SplitSeq(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m := shasumLine.FindStringSubmatch(line)
		if len(m) != 3 {
			continue
		}
		name := strings.TrimSpace(m[2])
		if name == filename || strings.HasSuffix(name, "/"+filename) {
			return strings.ToLower(m[1]), nil
		}
	}
	return "", fmt.Errorf("%s not listed in %s", filename, constants.MicronParserGoShasumsFilename)
}
