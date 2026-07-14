package playground

import (
	"net/http"
	"os"
)

// Handler serves the standalone playground assets rooted at directory.
func Handler(directory string) http.Handler {
	return http.FileServer(http.Dir(directory))
}

// DirectoryExists reports whether directory names an existing directory.
func DirectoryExists(directory string) bool {
	info, err := os.Stat(directory)
	return err == nil && info.IsDir()
}
