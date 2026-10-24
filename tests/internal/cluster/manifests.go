package cluster

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// PodSpec renders a workload fixture and decodes its Kubernetes pod specification.
func PodSpec(t testing.TB, filename string, values any) corev1.PodSpec {
	t.Helper()
	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	tmpl, err := template.New(filepath.Base(filename)).Option("missingkey=error").Parse(string(data))
	require.NoError(t, err)
	var rendered bytes.Buffer
	require.NoError(t, tmpl.Execute(&rendered, values))
	var spec corev1.PodSpec
	require.NoError(t, yaml.UnmarshalStrict(rendered.Bytes(), &spec))
	return spec
}
