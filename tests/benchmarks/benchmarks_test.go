package benchmarks

import (
	"os"
	"testing"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/khuedoan/homelab/tests/internal/testenv"
)

func TestBenchmarks(t *testing.T) {
	if os.Getenv("BENCHMARKS") != "1" {
		t.Skip("benchmarks require BENCHMARKS=1")
	}
	cluster := fixture.Connect(t, testenv.ForTest(t))
	t.Run("Storage", func(t *testing.T) { benchmarkStorage(t, cluster) })
	t.Run("Security", func(t *testing.T) { benchmarkSecurity(t, cluster) })
}
