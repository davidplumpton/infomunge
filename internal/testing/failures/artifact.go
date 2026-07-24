package failures

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"infomunge/internal/testing/teststore"
)

// Artifact records a minimized failing case from intensive testing.
type Artifact struct {
	ID                  int         `json:"id"`
	Timestamp           string      `json:"timestamp"`
	DetectedAt          string      `json:"detected_at,omitempty"`
	MinimizedAt         string      `json:"minimized_at,omitempty"`
	Property            string      `json:"property"`
	MinimizedExpression string      `json:"minimized_expression"`
	OriginalExpression  string      `json:"original_expression"`
	InputPayload        interface{} `json:"input_payload"`
	Expected            interface{} `json:"expected,omitempty"`
	Actual              interface{} `json:"actual,omitempty"`
	PanicStack          string      `json:"panic_stack,omitempty"`
	Seed                int64       `json:"seed"`
	Fingerprint         string      `json:"fingerprint"`
}

// SaveResult describes the canonical artifact selected by persistence.
type SaveResult struct {
	Artifact Artifact
	Path     string
	Saved    bool
}

// SaveArtifact writes an artifact JSON file to tmp/intensive-testing/failures.
// It returns the canonical artifact, including its assigned ID. If an artifact
// with the same fingerprint already exists, that artifact is returned and no
// new file is written.
func SaveArtifact(a Artifact) (SaveResult, error) {
	dir, err := teststore.Path("failures")
	if err != nil {
		return SaveResult{}, fmt.Errorf("resolve failures dir: %w", err)
	}
	return SaveArtifactToDir(dir, a)
}

// SaveArtifactToDir writes an artifact JSON file to the provided directory.
// It returns the canonical artifact, filepath, and whether a new file was
// written.
func SaveArtifactToDir(dir string, a Artifact) (SaveResult, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return SaveResult{}, fmt.Errorf("create failures dir: %w", err)
	}

	fingerprint := Fingerprint(a.MinimizedExpression, a.Property)
	existing, err := LoadArtifactsFromDir(dir)
	if err != nil {
		return SaveResult{}, err
	}
	for _, item := range existing {
		if item.Fingerprint == fingerprint {
			return SaveResult{Artifact: item}, nil
		}
	}

	nextID := nextSequentialID(existing)
	now := time.Now().UTC()

	a.ID = nextID
	a.Timestamp = now.Format(time.RFC3339)
	if strings.TrimSpace(a.DetectedAt) == "" {
		a.DetectedAt = a.Timestamp
	}
	if strings.TrimSpace(a.MinimizedAt) == "" {
		a.MinimizedAt = a.Timestamp
	}
	a.Fingerprint = fingerprint

	fileName := fmt.Sprintf(
		"%03d_%s_%s.json",
		nextID,
		sanitizeLabel(a.Property),
		now.Format("2006-01-02T15:04:05"),
	)
	path := filepath.Join(dir, fileName)

	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return SaveResult{}, fmt.Errorf("marshal artifact: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return SaveResult{}, fmt.Errorf("write artifact file: %w", err)
	}

	return SaveResult{
		Artifact: a,
		Path:     path,
		Saved:    true,
	}, nil
}

// LoadArtifacts reads all artifact JSON files from tmp/intensive-testing/failures.
func LoadArtifacts() ([]Artifact, error) {
	dir, err := teststore.Path("failures")
	if err != nil {
		return nil, fmt.Errorf("resolve failures dir: %w", err)
	}
	return LoadArtifactsFromDir(dir)
}

// LoadArtifactsFromDir reads all artifact JSON files from dir.
func LoadArtifactsFromDir(dir string) ([]Artifact, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read failures dir: %w", err)
	}

	artifacts := make([]Artifact, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read artifact %s: %w", entry.Name(), err)
		}

		var a Artifact
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, fmt.Errorf("unmarshal artifact %s: %w", entry.Name(), err)
		}

		if a.Fingerprint == "" {
			a.Fingerprint = Fingerprint(a.MinimizedExpression, a.Property)
		}
		if a.ID == 0 {
			if parsedID, ok := parseIDPrefix(entry.Name()); ok {
				a.ID = parsedID
			}
		}

		artifacts = append(artifacts, a)
	}

	sort.Slice(artifacts, func(i, j int) bool {
		if artifacts[i].ID == artifacts[j].ID {
			return artifacts[i].Timestamp < artifacts[j].Timestamp
		}
		return artifacts[i].ID < artifacts[j].ID
	})

	return artifacts, nil
}

// Fingerprint returns a stable identifier used for deduplicating failures.
func Fingerprint(minimizedExpression, property string) string {
	sum := sha256.Sum256([]byte(property + "\n" + minimizedExpression))
	return hex.EncodeToString(sum[:])
}

func nextSequentialID(existing []Artifact) int {
	maxID := 0
	for _, item := range existing {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func parseIDPrefix(name string) (int, bool) {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	parts := strings.SplitN(base, "_", 2)
	if len(parts) == 0 {
		return 0, false
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func sanitizeLabel(s string) string {
	if s == "" {
		return "unknown"
	}

	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}

	label := strings.Trim(b.String(), "_")
	if label == "" {
		return "unknown"
	}
	return strings.ToLower(label)
}
