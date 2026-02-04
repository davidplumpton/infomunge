package coremodules

import (
	"embed"

	unifiederrors "infomunge/internal/errors"
)

//go:embed *.im
var standardModuleFiles embed.FS

var standardModuleFileBySpec = map[string]string{
	"dw::core::Arrays":   "Arrays.im",
	"dw::core::Binaries": "Binaries.im",
	"dw::core::Numbers":  "Numbers.im",
	"dw::core::Objects":  "Objects.im",
	"dw::core::Strings":  "Strings.im",
}

// IsStandardModule reports whether the provided import spec references a built-in dw::core module.
func IsStandardModule(moduleSpec string) bool {
	_, ok := standardModuleFileBySpec[moduleSpec]
	return ok
}

// Source returns the module source for a supported built-in dw::core module.
func Source(moduleSpec string) (string, error) {
	fileName, ok := standardModuleFileBySpec[moduleSpec]
	if !ok {
		return "", unifiederrors.ParseErrorf("module %q is not a standard module", moduleSpec)
	}

	content, err := standardModuleFiles.ReadFile(fileName)
	if err != nil {
		return "", unifiederrors.WrapIOf(err, "failed to read standard module source for %s", moduleSpec)
	}
	return string(content), nil
}
