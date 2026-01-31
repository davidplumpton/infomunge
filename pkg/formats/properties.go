package formats

import (
	"strings"
)

func init() {
	RegisterReader("text/x-java-properties", readProperties)
	RegisterExtension(".properties", "text/x-java-properties")
}

func readProperties(content string) (interface{}, error) {
	result := make(Object)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		} else {
			parts = strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}
	return result, nil
}
