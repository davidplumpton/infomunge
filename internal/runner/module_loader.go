//go:build !js
// +build !js

package runner

import (
	"os"
	"path/filepath"
	"strings"

	unifiederrors "infomunge/internal/errors"
)

// ModuleLoader handles loading and caching of modules.
type ModuleLoader struct {
	BaseDir     string             // Base directory for module resolution
	SearchPaths []string           // Directories to search for modules
	cache       map[string]*Module // path -> loaded module
	loading     map[string]bool    // Currently loading (for cycle detection)
}

// NewModuleLoader creates a new ModuleLoader with the given base directory.
func NewModuleLoader(baseDir string) *ModuleLoader {
	searchPaths := []string{baseDir}
	modulesDir := filepath.Join(baseDir, "modules")
	if _, err := os.Stat(modulesDir); err == nil {
		searchPaths = append(searchPaths, modulesDir)
	}

	return &ModuleLoader{
		BaseDir:     baseDir,
		SearchPaths: searchPaths,
		cache:       make(map[string]*Module),
		loading:     make(map[string]bool),
	}
}

// Resolve converts a module spec (e.g., "modules::MyModule") into a module name and file path.
func (l *ModuleLoader) Resolve(moduleSpec string) (moduleName string, path string, err error) {
	parts := strings.Split(moduleSpec, "::")
	moduleName = parts[len(parts)-1]

	relPath := filepath.Join(parts...) + ".im"

	for _, searchPath := range l.SearchPaths {
		candidate := filepath.Join(searchPath, relPath)
		if _, err := os.Stat(candidate); err == nil {
			return moduleName, candidate, nil
		}
	}

	return "", "", unifiederrors.IOErrorf("module %q not found (searched in %v)", moduleSpec, l.SearchPaths)
}

// Load loads a module by its spec and returns the Module struct.
func (l *ModuleLoader) Load(moduleSpec string) (*Module, error) {
	moduleName, path, err := l.Resolve(moduleSpec)
	if err != nil {
		return nil, err
	}

	if m, ok := l.cache[path]; ok {
		return m, nil
	}

	if l.loading[path] {
		return nil, unifiederrors.ParseErrorf("cyclic import detected: %s", moduleSpec)
	}
	l.loading[path] = true
	defer delete(l.loading, path)

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, unifiederrors.WrapIOf(err, "failed to read module %s", moduleSpec)
	}

	ns, err := parseModuleContent(string(content), l)
	if err != nil {
		return nil, unifiederrors.WrapParsef(err, "failed to parse module %s", moduleSpec)
	}

	m := &Module{
		Name:      moduleName,
		Namespace: ns,
	}
	l.cache[path] = m

	return m, nil
}
