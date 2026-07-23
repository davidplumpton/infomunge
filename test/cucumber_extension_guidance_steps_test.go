package test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var extensionGuidePathPattern = regexp.MustCompile("`((?:internal|pkg|test|modules|docs)/[^`]+)`")

func (tc *testContext) everyRepositoryPathReferencedByTheExtensionGuideShouldExist() error {
	body, err := os.ReadFile("../docs/EXTENDING.md")
	if err != nil {
		return fmt.Errorf("read extension guide: %w", err)
	}

	references := extensionGuidePathPattern.FindAllStringSubmatch(string(body), -1)
	if len(references) == 0 {
		return fmt.Errorf("extension guide does not reference any repository paths")
	}

	missing := make(map[string]struct{})
	for _, match := range references {
		reference := match[1]
		matches, globErr := filepath.Glob(filepath.Join("..", reference))
		if globErr != nil {
			return fmt.Errorf("invalid extension guide path pattern %q: %w", reference, globErr)
		}
		if len(matches) == 0 {
			missing[reference] = struct{}{}
		}
	}

	if len(missing) == 0 {
		return nil
	}

	paths := make([]string, 0, len(missing))
	for path := range missing {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return fmt.Errorf("extension guide references missing repository paths: %s", strings.Join(paths, ", "))
}
