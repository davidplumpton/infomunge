package mutation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/cucumber/gherkin/go/v26"
	messages "github.com/cucumber/messages/go/v21"
)

var (
	namedInputRegex  = regexp.MustCompile(`(?i)^input\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+is:?$`)
	quotedValueRegex = regexp.MustCompile(`"([^"]*)"`)
)

// CorpusEntry is a single extracted scenario sample for mutation/differential testing.
type CorpusEntry struct {
	Script         string
	Inputs         map[string]string
	ExpectedOutput string
	SourceFile     string
	ScenarioName   string
}

// ExtractCorpus parses .feature files under featuresDir and extracts scenario
// script/input/expected-output triples.
func ExtractCorpus(featuresDir string) ([]CorpusEntry, error) {
	featureFiles, err := collectFeatureFiles(featuresDir)
	if err != nil {
		return nil, err
	}

	entries := make([]CorpusEntry, 0, len(featureFiles))
	for _, featureFile := range featureFiles {
		fileEntries, err := extractFileCorpus(featureFile)
		if err != nil {
			return nil, err
		}
		entries = append(entries, fileEntries...)
	}

	return entries, nil
}

func collectFeatureFiles(featuresDir string) ([]string, error) {
	var featureFiles []string
	err := filepath.WalkDir(featuresDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".feature") {
			featureFiles = append(featureFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk features dir %q: %w", featuresDir, err)
	}

	sort.Strings(featureFiles)
	return featureFiles, nil
}

func extractFileCorpus(featureFile string) ([]CorpusEntry, error) {
	contentBytes, err := os.ReadFile(featureFile)
	if err != nil {
		return nil, fmt.Errorf("read feature file %q: %w", featureFile, err)
	}

	doc, err := gherkin.ParseGherkinDocument(strings.NewReader(string(contentBytes)), func() string { return "id" })
	if err != nil {
		return nil, fmt.Errorf("parse feature file %q: %w", featureFile, err)
	}
	if doc == nil || doc.Feature == nil {
		return nil, nil
	}

	entries := make([]CorpusEntry, 0)
	featureBackground := make([]*messages.Step, 0)
	for _, child := range doc.Feature.Children {
		switch {
		case child.Background != nil:
			featureBackground = cloneSteps(child.Background.Steps)
		case child.Scenario != nil:
			entry, ok := extractScenarioEntry(featureFile, child.Scenario.Name, mergeSteps(featureBackground, child.Scenario.Steps))
			if ok {
				entries = append(entries, entry)
			}
		case child.Rule != nil:
			ruleEntries := extractRuleEntries(featureFile, child.Rule, featureBackground)
			entries = append(entries, ruleEntries...)
		}
	}

	return entries, nil
}

func extractRuleEntries(featureFile string, rule *messages.Rule, featureBackground []*messages.Step) []CorpusEntry {
	entries := make([]CorpusEntry, 0)
	ruleBackground := make([]*messages.Step, 0)
	for _, child := range rule.Children {
		switch {
		case child.Background != nil:
			ruleBackground = cloneSteps(child.Background.Steps)
		case child.Scenario != nil:
			combined := mergeSteps(featureBackground, ruleBackground)
			combined = mergeSteps(combined, child.Scenario.Steps)
			entry, ok := extractScenarioEntry(featureFile, child.Scenario.Name, combined)
			if ok {
				entries = append(entries, entry)
			}
		}
	}
	return entries
}

func extractScenarioEntry(featureFile, scenarioName string, steps []*messages.Step) (CorpusEntry, bool) {
	entry := CorpusEntry{
		Inputs:       map[string]string{},
		SourceFile:   featureFile,
		ScenarioName: scenarioName,
	}

	phase := ""
	for _, step := range steps {
		keyword := strings.TrimSpace(step.Keyword)
		switch {
		case strings.EqualFold(keyword, "Given"):
			phase = "given"
		case strings.EqualFold(keyword, "When"):
			phase = "when"
		case strings.EqualFold(keyword, "Then"):
			phase = "then"
		}

		text := strings.TrimSpace(step.Text)
		textLower := strings.ToLower(text)
		docString := ""
		if step.DocString != nil {
			docString = step.DocString.Content
		}

		if docString != "" {
			switch {
			case strings.Contains(textLower, "infomunge processes"):
				entry.Script = docString
			case strings.Contains(textLower, "following script") || strings.EqualFold(textLower, "the script"):
				entry.Script = docString
			case strings.Contains(textLower, "input content") && looksLikeScript(docString):
				entry.Script = docString
			case strings.Contains(textLower, "input payload is"):
				entry.Inputs["payload"] = docString
			case strings.Contains(textLower, "following json input"):
				entry.Inputs["payload"] = docString
			case strings.Contains(textLower, "following xml input"):
				entry.Inputs["payload"] = docString
			case strings.Contains(textLower, "following csv input"):
				entry.Inputs["payload"] = docString
			case strings.Contains(textLower, "following yaml input"):
				entry.Inputs["payload"] = docString
			case strings.Contains(textLower, "following properties input"):
				entry.Inputs["payload"] = docString
			default:
				if matches := namedInputRegex.FindStringSubmatch(text); len(matches) == 2 {
					entry.Inputs[matches[1]] = docString
				}
			}
		}

		if textLower == "the following input content:" && docString != "" && looksLikeScript(docString) {
			entry.Script = docString
		}

		if strings.HasPrefix(textLower, "the input ") {
			if value, ok := parseInlineInputValue(text); ok {
				entry.Inputs["payload"] = value
			}
		}

		if phase == "then" {
			if docString != "" {
				entry.ExpectedOutput = docString
			} else if value, ok := parseExpectedValue(text); ok {
				entry.ExpectedOutput = value
			}
		}
	}

	if strings.TrimSpace(entry.Script) == "" || strings.TrimSpace(entry.ExpectedOutput) == "" {
		return CorpusEntry{}, false
	}
	return entry, true
}

func mergeSteps(a, b []*messages.Step) []*messages.Step {
	out := make([]*messages.Step, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}

func cloneSteps(steps []*messages.Step) []*messages.Step {
	out := make([]*messages.Step, len(steps))
	copy(out, steps)
	return out
}

func looksLikeScript(content string) bool {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false
	}
	return strings.Contains(trimmed, "---") || strings.HasPrefix(trimmed, "%im") || strings.HasPrefix(trimmed, "%dw")
}

func parseInlineInputValue(stepText string) (string, bool) {
	matches := quotedValueRegex.FindStringSubmatch(stepText)
	if len(matches) == 2 {
		return matches[1], true
	}

	prefix := "the input "
	if strings.HasPrefix(strings.ToLower(stepText), prefix) {
		return strings.TrimSpace(stepText[len(prefix):]), true
	}

	return "", false
}

func parseExpectedValue(stepText string) (string, bool) {
	if matches := quotedValueRegex.FindStringSubmatch(stepText); len(matches) == 2 {
		return matches[1], true
	}

	lower := strings.ToLower(stepText)
	const marker = "should be "
	idx := strings.Index(lower, marker)
	if idx >= 0 {
		value := strings.TrimSpace(stepText[idx+len(marker):])
		if value != "" {
			return value, true
		}
	}

	return "", false
}
