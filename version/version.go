package version

/*
|--------------------------------------------------------------------------
| Version
|--------------------------------------------------------------------------
|
| Defines the current Gofreight semver and Module() tag string.
| 
| Update Version when cutting a release; documentation and go install @v
| tags must stay in sync.
| 
| Central semver constants consumed by the CLI, generator scaffolds, and
| demo welcome pages.
| 
| Bump Version here when releasing; tag Git with v prefix matching
| Module().
| 
| Symbols defined here include: Version (exported value); Module (Module
| returns the Go module version tag (with "v" prefix).).
| 
*/

// Version is the current Gofreight release (SemVer, without "v" prefix).
const Version = "0.5.0"

// Module returns the Go module version tag (with "v" prefix).
func Module() string {
	return "v" + Version
}
