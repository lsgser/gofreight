package version

// Version is the current Gofreight release (SemVer, without "v" prefix).
const Version = "0.4.0"

// Module returns the Go module version tag (with "v" prefix).
func Module() string {
	return "v" + Version
}
