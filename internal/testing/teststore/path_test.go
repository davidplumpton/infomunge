package teststore

import (
	"path/filepath"
	"testing"
)

func TestPathsResolveFromNestedPackageDirectory(t *testing.T) {
	repositoryRoot, err := RepositoryRoot()
	if err != nil {
		t.Fatalf("RepositoryRoot() before changing directory: %v", err)
	}

	nested := filepath.Join(repositoryRoot, "internal", "testing", "mutation")
	t.Chdir(nested)

	gotRepositoryRoot, err := RepositoryRoot()
	if err != nil {
		t.Fatalf("RepositoryRoot() from nested directory: %v", err)
	}
	if gotRepositoryRoot != repositoryRoot {
		t.Fatalf("RepositoryRoot() = %q, want %q", gotRepositoryRoot, repositoryRoot)
	}

	gotStoreRoot, err := Root()
	if err != nil {
		t.Fatalf("Root() from nested directory: %v", err)
	}
	wantStoreRoot := filepath.Join(repositoryRoot, "tmp", "intensive-testing")
	if gotStoreRoot != wantStoreRoot {
		t.Fatalf("Root() = %q, want %q", gotStoreRoot, wantStoreRoot)
	}

	gotFailuresPath, err := Path("failures")
	if err != nil {
		t.Fatalf("Path(failures) from nested directory: %v", err)
	}
	wantFailuresPath := filepath.Join(wantStoreRoot, "failures")
	if gotFailuresPath != wantFailuresPath {
		t.Fatalf("Path(failures) = %q, want %q", gotFailuresPath, wantFailuresPath)
	}
}
