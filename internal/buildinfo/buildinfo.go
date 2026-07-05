// SPDX-License-Identifier: MIT
package buildinfo

var (
	Version = "dev"
	Commit  = "dev"
)

func BuildLabel() string {
	if Commit == "" || Commit == "dev" {
		return "dev"
	}
	return Commit
}
