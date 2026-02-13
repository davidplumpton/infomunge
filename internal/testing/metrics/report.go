package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"infomunge/internal/testing/failures"
)

const reportPath = "tmp/intensive-testing/report.json"

var coveragePattern = regexp.MustCompile(`coverage:\s+([0-9]+(?:\.[0-9]+)?)%`)

var panicCount atomic.Int64
var mutationKilled atomic.Int64
var mutationSurvived atomic.Int64

type Report struct {
	GeneratedAt string         `json:"generated_at"`
	Package     string         `json:"package"`
	Metrics     MetricSnapshot `json:"metrics"`
}

type Options struct {
	EnableCoverage bool
}

type MetricSnapshot struct {
	UniqueFailures            int      `json:"unique_failures"`
	ShrinkRatio               *float64 `json:"shrink_ratio,omitempty"`
	TimeToMinimalReproSeconds *float64 `json:"time_to_minimal_repro_seconds,omitempty"`
	PromotedRegressions       int      `json:"promoted_cucumber_regressions"`
	PanicCount                int64    `json:"panic_count"`
	MutationKilled            int64    `json:"mutation_killed"`
	MutationSurvived          int64    `json:"mutation_survived"`
	MutationKillRate          *float64 `json:"mutation_kill_rate,omitempty"`
	PreprocessorCoveragePct   *float64 `json:"preprocessor_coverage_pct,omitempty"`
	EvaluatorCoveragePct      *float64 `json:"evaluator_coverage_pct,omitempty"`
	PreprocessorCoverageDelta *float64 `json:"preprocessor_coverage_delta,omitempty"`
	EvaluatorCoverageDelta    *float64 `json:"evaluator_coverage_delta,omitempty"`
}

type reportFile struct {
	Latest  Report   `json:"latest"`
	History []Report `json:"history"`
}

// RecordPanic increments the panic counter for the active test process.
func RecordPanic() {
	panicCount.Add(1)
}

// RecordMutationOutcome records whether a non-equivalent mutation was killed.
func RecordMutationOutcome(killed bool) {
	if killed {
		mutationKilled.Add(1)
		return
	}
	mutationSurvived.Add(1)
}

// ReportAndPersist computes intensive-testing metrics, prints a summary to
// stdout, and writes tmp/intensive-testing/report.json for trend tracking.
func ReportAndPersist(pkg string, options Options) {
	snapshot := collectSnapshot(options)
	current := Report{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Package:     pkg,
		Metrics:     snapshot,
	}

	printSummary(current)
	if err := writeReportFile(current); err != nil {
		fmt.Printf("intensive-metrics: failed to write report: %v\n", err)
	}
}

func collectSnapshot(options Options) MetricSnapshot {
	artifacts, err := failures.LoadArtifacts()
	if err != nil {
		artifacts = nil
	}

	shrinkRatio := averageShrinkRatio(artifacts)
	timeToRepro := averageTimeToMinimalRepro(artifacts)
	promoted := countPromotedRegressions()

	killed := mutationKilled.Load()
	survived := mutationSurvived.Load()
	var killRate *float64
	if killed+survived > 0 {
		ratio := float64(killed) / float64(killed+survived)
		killRate = &ratio
	}

	var currentPre *float64
	var currentEval *float64
	if options.EnableCoverage {
		currentPre, currentEval = currentCoverage()
	}
	var deltaPre *float64
	var deltaEval *float64
	if previous, ok := readPreviousCoverage(); ok {
		if previous.PreprocessorCoveragePct != nil && currentPre != nil {
			d := *currentPre - *previous.PreprocessorCoveragePct
			deltaPre = &d
		}
		if previous.EvaluatorCoveragePct != nil && currentEval != nil {
			d := *currentEval - *previous.EvaluatorCoveragePct
			deltaEval = &d
		}
	}

	return MetricSnapshot{
		UniqueFailures:            countUniqueFailures(artifacts),
		ShrinkRatio:               shrinkRatio,
		TimeToMinimalReproSeconds: timeToRepro,
		PromotedRegressions:       promoted,
		PanicCount:                panicCount.Load(),
		MutationKilled:            killed,
		MutationSurvived:          survived,
		MutationKillRate:          killRate,
		PreprocessorCoveragePct:   currentPre,
		EvaluatorCoveragePct:      currentEval,
		PreprocessorCoverageDelta: deltaPre,
		EvaluatorCoverageDelta:    deltaEval,
	}
}

func countUniqueFailures(artifacts []failures.Artifact) int {
	seen := make(map[string]struct{}, len(artifacts))
	for _, a := range artifacts {
		fp := strings.TrimSpace(a.Fingerprint)
		if fp == "" {
			fp = failures.Fingerprint(a.MinimizedExpression, a.Property)
		}
		seen[fp] = struct{}{}
	}
	return len(seen)
}

func averageShrinkRatio(artifacts []failures.Artifact) *float64 {
	total := 0.0
	count := 0
	for _, a := range artifacts {
		orig := len(strings.TrimSpace(a.OriginalExpression))
		if orig == 0 {
			continue
		}
		minimized := len(strings.TrimSpace(a.MinimizedExpression))
		total += float64(minimized) / float64(orig)
		count++
	}
	if count == 0 {
		return nil
	}
	v := total / float64(count)
	return &v
}

func averageTimeToMinimalRepro(artifacts []failures.Artifact) *float64 {
	totalSeconds := 0.0
	count := 0
	for _, a := range artifacts {
		startRaw := strings.TrimSpace(a.DetectedAt)
		endRaw := strings.TrimSpace(a.MinimizedAt)
		if startRaw == "" || endRaw == "" {
			continue
		}
		start, err := time.Parse(time.RFC3339, startRaw)
		if err != nil {
			continue
		}
		end, err := time.Parse(time.RFC3339, endRaw)
		if err != nil {
			continue
		}
		if end.Before(start) {
			continue
		}
		totalSeconds += end.Sub(start).Seconds()
		count++
	}
	if count == 0 {
		return nil
	}
	v := totalSeconds / float64(count)
	return &v
}

func countPromotedRegressions() int {
	root, err := repoRoot()
	if err != nil {
		return 0
	}

	featuresDir := filepath.Join(root, "test", "features")
	entries, err := os.ReadDir(featuresDir)
	if err != nil {
		return 0
	}

	const marker = "Auto-generated from intensive testing failure"
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".feature" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(featuresDir, entry.Name()))
		if err != nil {
			continue
		}
		if bytes.Contains(data, []byte(marker)) {
			count++
		}
	}
	return count
}

func currentCoverage() (preprocessor *float64, evaluator *float64) {
	root, err := repoRoot()
	if err != nil {
		return nil, nil
	}

	preprocessor = runCoverage(root, "./internal/preprocessor")
	evaluator = runCoverage(root, "./internal/evaluator")
	return preprocessor, evaluator
}

func runCoverage(root, pkg string) *float64 {
	cmd := exec.Command("go", "test", "-cover", pkg)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	m := coveragePattern.FindStringSubmatch(string(out))
	if len(m) < 2 {
		return nil
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return nil
	}
	return &v
}

func readPreviousCoverage() (MetricSnapshot, bool) {
	root, err := repoRoot()
	if err != nil {
		return MetricSnapshot{}, false
	}
	data, err := os.ReadFile(filepath.Join(root, reportPath))
	if err != nil {
		return MetricSnapshot{}, false
	}
	var rf reportFile
	if err := json.Unmarshal(data, &rf); err != nil {
		return MetricSnapshot{}, false
	}
	return rf.Latest.Metrics, true
}

func writeReportFile(current Report) error {
	root, err := repoRoot()
	if err != nil {
		return err
	}

	path := filepath.Join(root, reportPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}

	var rf reportFile
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &rf)
	}
	if rf.Latest.GeneratedAt != "" {
		rf.History = append(rf.History, rf.Latest)
	}
	rf.Latest = current

	data, err := json.MarshalIndent(rf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

func printSummary(report Report) {
	m := report.Metrics
	fmt.Printf("\nintensive-metrics (%s)\n", report.Package)
	fmt.Printf("  unique failures: %d\n", m.UniqueFailures)
	fmt.Printf("  shrink ratio: %s\n", formatRatio(m.ShrinkRatio))
	fmt.Printf("  time-to-minimal-repro (s): %s\n", formatFloat(m.TimeToMinimalReproSeconds))
	fmt.Printf("  promoted cucumber regressions: %d\n", m.PromotedRegressions)
	fmt.Printf("  panic count: %d\n", m.PanicCount)
	fmt.Printf("  mutation kill rate: %s (killed=%d survived=%d)\n", formatRatio(m.MutationKillRate), m.MutationKilled, m.MutationSurvived)
	fmt.Printf("  preprocessor coverage: %s (delta %s)\n", formatPct(m.PreprocessorCoveragePct), formatSigned(m.PreprocessorCoverageDelta))
	fmt.Printf("  evaluator coverage: %s (delta %s)\n", formatPct(m.EvaluatorCoveragePct), formatSigned(m.EvaluatorCoverageDelta))
	fmt.Printf("  report file: %s\n\n", reportPath)
}

func repoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	cur := wd
	for {
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			return cur, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", fmt.Errorf("go.mod not found from %s", wd)
		}
		cur = parent
	}
}

func formatRatio(v *float64) string {
	if v == nil {
		return "n/a"
	}
	return fmt.Sprintf("%.2f", *v)
}

func formatPct(v *float64) string {
	if v == nil {
		return "n/a"
	}
	return fmt.Sprintf("%.2f%%", *v)
}

func formatFloat(v *float64) string {
	if v == nil {
		return "n/a"
	}
	return fmt.Sprintf("%.2f", *v)
}

func formatSigned(v *float64) string {
	if v == nil {
		return "n/a"
	}
	return fmt.Sprintf("%+.2f%%", *v)
}
