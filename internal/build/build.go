package build

// These variables are populated via -ldflags at build time by GoReleaser / Makefile.
// Default values act as sensible fallbacks for non-release local builds.
var (
	version = "dev" // kept here for completeness but main still exposes its own version
	commit  = "none"
	date    = "unknown"
	builtBy = "local"
)

// Info returns a struct-style map with build metadata.
func Info() map[string]string {
	return map[string]string{
		"version": version,
		"commit":  commit,
		"date":    date,
		"builtBy": builtBy,
	}
}
