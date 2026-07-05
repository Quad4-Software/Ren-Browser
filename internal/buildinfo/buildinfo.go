package buildinfo

var (
	Version = "0.1.0"
	Commit  = "dev"
)

func BuildLabel() string {
	if Commit == "" || Commit == "dev" {
		return "dev"
	}
	return Commit
}
