package runner

import (
	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/readlimit"
)

const (
	// MaxScriptFileBytes limits script-like files read from disk by the CLI and runner.
	MaxScriptFileBytes = 1 * 1024 * 1024
	MaxModuleFileBytes = MaxScriptFileBytes
)

// ReadScriptFile reads a script file without allowing unbounded memory growth.
func ReadScriptFile(filePath string) (string, error) {
	content, tooLarge, err := readlimit.ReadFile(filePath, MaxScriptFileBytes)
	if err != nil {
		return "", unifiederrors.WrapIOf(err, "error reading script file: %s", filePath)
	}
	if tooLarge {
		return "", unifiederrors.ValidationErrorf("script file %s exceeds maximum size of %d bytes", filePath, MaxScriptFileBytes)
	}
	return string(content), nil
}

func readModuleFile(moduleSpec, filePath string) (string, error) {
	content, tooLarge, err := readlimit.ReadFile(filePath, MaxModuleFileBytes)
	if err != nil {
		return "", unifiederrors.WrapIOf(err, "failed to read module %s from %s", moduleSpec, filePath)
	}
	if tooLarge {
		return "", unifiederrors.ValidationErrorf("module file %s for %s exceeds maximum size of %d bytes", filePath, moduleSpec, MaxModuleFileBytes)
	}
	return string(content), nil
}
