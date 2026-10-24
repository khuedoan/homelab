package benchmarks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeReport(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		diagnostics string
		failures    int
	}{
		{
			name:     "plain JSON",
			output:   `{"total_fail": 3}`,
			failures: 3,
		},
		{
			name:        "kube-bench warning",
			output:      "W1105 04:42:35.548280 util.go:340]\nUnable to find the programs kubectl or kubelet in the PATH.\n{\"total_fail\": 28}\n",
			diagnostics: "W1105 04:42:35.548280 util.go:340]\nUnable to find the programs kubectl or kubelet in the PATH.\n",
			failures:    28,
		},
		{
			name:        "fio warning",
			output:      "fio: posix_fallocate fails: Not supported\n{\"total_fail\": 5}",
			diagnostics: "fio: posix_fallocate fails: Not supported\n",
			failures:    5,
		},
		{
			name:        "braces in diagnostics",
			output:      "warning: options {a,b} unavailable\n  {\"total_fail\": 7}\n\t",
			diagnostics: "warning: options {a,b} unavailable\n",
			failures:    7,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var report struct {
				Failures int `json:"total_fail"`
			}

			diagnostics, err := decodeReport(test.output, &report)

			require.NoError(t, err)
			require.Equal(t, test.diagnostics, diagnostics)
			require.Equal(t, test.failures, report.Failures)
		})
	}
}

func TestDecodeReportRejectsInvalidOutput(t *testing.T) {
	tests := map[string]string{
		"malformed JSON":     "warning\n{broken}",
		"truncated JSON":     "warning\n{\"total_fail\":",
		"trailing text":      "{}\nwarning",
		"second JSON object": "{}\n{}",
		"no JSON object":     "warning: options {a,b} unavailable\n",
		"empty output":       "",
	}

	for name, output := range tests {
		t.Run(name, func(t *testing.T) {
			var report map[string]any

			_, err := decodeReport(output, &report)

			require.Error(t, err)
		})
	}
}
