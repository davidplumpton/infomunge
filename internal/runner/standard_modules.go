package runner

import (
	"strings"

	coremodules "infomunge/modules/dw/core"
)

func isStandardModule(moduleSpec string) bool {
	return coremodules.IsStandardModule(moduleSpec)
}

func standardModuleSource(moduleSpec string) (string, error) {
	return coremodules.Source(moduleSpec)
}

func moduleNameFromSpec(moduleSpec string) string {
	parts := strings.Split(moduleSpec, "::")
	return parts[len(parts)-1]
}
