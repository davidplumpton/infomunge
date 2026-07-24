package version

// Version information
var (
	Version   = "0.1.1"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// String returns the full version string
func String() string {
	return Version
}
