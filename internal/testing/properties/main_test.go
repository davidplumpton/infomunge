package properties_test

import (
	"os"
	"testing"

	"infomunge/internal/testing/metrics"
)

func TestMain(m *testing.M) {
	code := m.Run()
	metrics.ReportAndPersist("properties", metrics.Options{EnableCoverage: false})
	os.Exit(code)
}
