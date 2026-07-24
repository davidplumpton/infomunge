// Package teststore resolves repository-local paths used by intensive tests.
package teststore

import (
	"fmt"
	"os"
	"path/filepath"
)

const relativeRoot = "tmp/intensive-testing"

// RepositoryRoot returns the nearest ancestor containing the repository's
// go.mod. Go executes package tests from the package directory, so callers
// must not assume their working directory is the repository root.
func RepositoryRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	return repositoryRootFrom(wd)
}

// Root returns the repository-local directory for intensive-testing output.
func Root() (string, error) {
	root, err := RepositoryRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, relativeRoot), nil
}

// Path returns a path beneath the repository-local intensive-testing output
// directory.
func Path(elements ...string) (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{root}, elements...)...), nil
}

func repositoryRootFrom(start string) (string, error) {
	cur, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve working directory %s: %w", start, err)
	}

	for {
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			return cur, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", fmt.Errorf("go.mod not found from %s", start)
		}
		cur = parent
	}
}
