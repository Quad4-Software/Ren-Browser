// SPDX-License-Identifier: MIT
package buildinfo

var (
	Version   = "dev"
	Commit    = "dev"
	BuildTime = "unknown"
)

func BuildLabel() string {
	if Commit == "" || Commit == "dev" {
		return "dev"
	}
	return Commit
}

func PrintVersion(appName string) {
	println(appName + " " + Version)
	println("commit: " + Commit)
	println("build time: " + BuildTime)
}
