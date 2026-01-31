package version

// Version information
var (
	Version   = "0.1.0"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// String returns the full version string
func String() string {
	return Version
}

// FullString returns detailed version information
func FullString() string {
	return Version + " (commit: " + Commit + ", built: " + BuildDate + ")"
}
