//go:build js
// +build js

package runner

import (
	unifiederrors "infomunge/internal/errors"
)

// ModuleLoader handles loading and caching of modules in JS builds.
type ModuleLoader struct {
	BaseDir string             // Base directory for module resolution
	cache   map[string]*Module // module spec -> loaded module
	loading map[string]bool    // currently loading modules
}

// NewModuleLoader creates a module loader for JS builds with standard module support.
func NewModuleLoader(baseDir string) *ModuleLoader {
	return &ModuleLoader{
		BaseDir: baseDir,
		cache:   make(map[string]*Module),
		loading: make(map[string]bool),
	}
}

// Resolve returns metadata for standard modules in JS builds.
func (l *ModuleLoader) Resolve(moduleSpec string) (moduleName string, path string, err error) {
	if isStandardModule(moduleSpec) {
		return moduleNameFromSpec(moduleSpec), moduleSpec, nil
	}
	return "", "", unifiederrors.ParseError("imports are not supported in JS builds")
}

// Load supports standard modules and rejects file-based imports in JS builds.
func (l *ModuleLoader) Load(moduleSpec string) (*Module, error) {
	if !isStandardModule(moduleSpec) {
		return nil, unifiederrors.ParseError("imports are not supported in JS builds")
	}

	if m, ok := l.cache[moduleSpec]; ok {
		return m, nil
	}
	if l.loading[moduleSpec] {
		return nil, unifiederrors.ParseErrorf("cyclic import detected: %s", moduleSpec)
	}
	l.loading[moduleSpec] = true
	defer delete(l.loading, moduleSpec)

	content, err := standardModuleSource(moduleSpec)
	if err != nil {
		return nil, unifiederrors.WrapParsef(err, "failed to load standard module %s", moduleSpec)
	}
	ns, err := parseModuleContent(content, l)
	if err != nil {
		return nil, unifiederrors.WrapParsef(err, "failed to parse standard module %s", moduleSpec)
	}

	m := &Module{
		Name:      moduleNameFromSpec(moduleSpec),
		Namespace: ns,
	}
	l.cache[moduleSpec] = m
	return m, nil
}
